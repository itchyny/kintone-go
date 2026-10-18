package kintone

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"path/filepath"
)

// UploadFile uploads a file and returns the file key. The content type of the
// file is detected from the extension of the name.
func (c *Client) UploadFile(ctx context.Context, name string, r io.Reader) (string, error) {
	pr, pw := io.Pipe()
	w := multipart.NewWriter(pw)
	go func() {
		h := textproto.MIMEHeader{}
		h.Set("Content-Disposition", multipart.FileContentDisposition("file", name))
		contentType := mime.TypeByExtension(filepath.Ext(name))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		h.Set("Content-Type", contentType)
		part, err := w.CreatePart(h)
		if err == nil {
			_, err = io.Copy(part, r)
		}
		if err == nil {
			err = w.Close()
		}
		pw.CloseWithError(err)
	}()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint("file.json").String(), pr)
	if err != nil {
		pr.Close()
		return "", fmt.Errorf("kintone: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	res, err := c.send(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	var result struct {
		FileKey string `json:"fileKey"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("kintone: decode response: %w", err)
	}
	return result.FileKey, nil
}

// DownloadFile downloads the file of the file key and returns the content and
// the content type. The caller must close the content.
func (c *Client) DownloadFile(ctx context.Context, fileKey string) (io.ReadCloser, string, error) {
	u := c.endpoint("file.json")
	u.RawQuery = url.Values{"fileKey": {fileKey}}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, "", fmt.Errorf("kintone: %w", err)
	}
	res, err := c.send(req)
	if err != nil {
		return nil, "", err
	}
	return res.Body, res.Header.Get("Content-Type"), nil
}
