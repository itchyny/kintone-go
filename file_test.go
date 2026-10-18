package kintone

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type uploadedFile struct {
	path        string
	fieldName   string
	fileName    string
	contentType string
	content     string
}

func newUploadServer(t *testing.T, status int, body string) (*httptest.Server, *uploadedFile) {
	t.Helper()
	var got uploadedFile
	ts := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.path = r.URL.Path
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "multipart/form-data" {
			t.Errorf("Content-Type: got %q", r.Header.Get("Content-Type"))
		}
		reader, err := r.MultipartReader()
		if err != nil {
			t.Errorf("MultipartReader returned error: %v", err)
			return
		}
		part, err := reader.NextPart()
		if err != nil {
			t.Errorf("NextPart returned error: %v", err)
			return
		}
		b, _ := io.ReadAll(part)
		got = uploadedFile{r.URL.Path, part.FormName(), part.FileName(), part.Header.Get("Content-Type"), string(b)}
		if _, err := reader.NextPart(); err != io.EOF {
			t.Errorf("NextPart should return io.EOF but got: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	return ts, &got
}

func TestClientUploadFile(t *testing.T) {
	testCases := []struct {
		name            string
		fileName        string
		opts            []Option
		wantPath        string
		wantContentType string
	}{
		{
			name:            "text file",
			fileName:        "hello.txt",
			wantPath:        "/k/v1/file.json",
			wantContentType: "text/plain; charset=utf-8",
		},
		{
			name:            "Japanese file name",
			fileName:        "見積書.pdf",
			wantPath:        "/k/v1/file.json",
			wantContentType: "application/pdf",
		},
		{
			name:            "unknown extension",
			fileName:        "data.unknown-extension",
			wantPath:        "/k/v1/file.json",
			wantContentType: "application/octet-stream",
		},
		{
			name:            "guest space",
			fileName:        "hello.txt",
			opts:            []Option{WithGuestSpace(3)},
			wantPath:        "/k/guest/3/v1/file.json",
			wantContentType: "text/plain; charset=utf-8",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newUploadServer(t, http.StatusOK, `{"fileKey": "c15b3870-7505-4ab6-9d8d-b9bdbc74f5d6"}`)
			c := newTestClient(t, ts, tc.opts...)
			fileKey, err := c.UploadFile(t.Context(), tc.fileName, strings.NewReader("Hello, world"))
			if err != nil {
				t.Fatalf("UploadFile returned error: %v", err)
			}
			if want := "c15b3870-7505-4ab6-9d8d-b9bdbc74f5d6"; fileKey != want {
				t.Errorf("file key: got %q, want %q", fileKey, want)
			}
			want := uploadedFile{tc.wantPath, "file", tc.fileName, tc.wantContentType, "Hello, world"}
			if *got != want {
				t.Errorf("uploaded file: got %+v, want %+v", *got, want)
			}
		})
	}
}

type errorReader struct{ err error }

func (r errorReader) Read([]byte) (int, error) { return 0, r.err }

func TestClientUploadFileReadError(t *testing.T) {
	ts := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		_, _ = io.WriteString(w, `{"fileKey": "x"}`)
	}))
	c := newTestClient(t, ts)
	readErr := errors.New("read error")
	if _, err := c.UploadFile(t.Context(), "a.txt", errorReader{readErr}); !errors.Is(err, readErr) {
		t.Errorf("UploadFile should return the read error but got: %v", err)
	}
}

func TestClientUploadFileError(t *testing.T) {
	ts, _ := newUploadServer(t, http.StatusRequestEntityTooLarge,
		`{"code": "TEST_ERROR", "id": "abc123", "message": "Test error."}`)
	c := newTestClient(t, ts)
	if _, err := c.UploadFile(t.Context(), "a.txt", strings.NewReader("x")); !IsErrorCode(err, "TEST_ERROR") {
		t.Errorf("UploadFile should return TEST_ERROR error but got: %v", err)
	}
}

func TestClientDownloadFile(t *testing.T) {
	testCases := []struct {
		name     string
		opts     []Option
		wantPath string
	}{
		{
			name:     "file",
			wantPath: "/k/v1/file.json",
		},
		{
			name:     "guest space",
			opts:     []Option{WithGuestSpace(3)},
			wantPath: "/k/guest/3/v1/file.json",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var gotMethod, gotPath, gotFileKey string
			ts := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod, gotPath, gotFileKey = r.Method, r.URL.Path, r.URL.Query().Get("fileKey")
				w.Header().Set("Content-Type", "image/png")
				_, _ = io.WriteString(w, "\x89PNG")
			}))
			c := newTestClient(t, ts, tc.opts...)
			body, contentType, err := c.DownloadFile(t.Context(), "2012+02/06 1155587E")
			if err != nil {
				t.Fatalf("DownloadFile returned error: %v", err)
			}
			defer body.Close()
			b, err := io.ReadAll(body)
			if err != nil {
				t.Fatalf("ReadAll returned error: %v", err)
			}
			if gotMethod != http.MethodGet || gotPath != tc.wantPath || gotFileKey != "2012+02/06 1155587E" {
				t.Errorf("request: got %s %s fileKey=%q", gotMethod, gotPath, gotFileKey)
			}
			if string(b) != "\x89PNG" {
				t.Errorf("content: got %q, want %q", b, "\x89PNG")
			}
			if contentType != "image/png" {
				t.Errorf("content type: got %q, want %q", contentType, "image/png")
			}
		})
	}
}

func TestClientDownloadFileError(t *testing.T) {
	ts, _ := newTestServer(t, http.StatusNotFound,
		`{"code": "TEST_ERROR", "id": "abc123", "message": "Test error."}`)
	c := newTestClient(t, ts)
	body, _, err := c.DownloadFile(t.Context(), "x")
	if !IsErrorCode(err, "TEST_ERROR") {
		t.Errorf("DownloadFile should return TEST_ERROR error but got: %v", err)
	}
	if body != nil {
		t.Error("body should be nil on error")
	}
}
