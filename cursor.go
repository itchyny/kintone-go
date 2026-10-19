package kintone

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"net/http"
	"slices"
	"time"
)

// CreateCursorRequest is the request of [Client.CreateCursor].
type CreateCursorRequest struct {
	App    int64    `json:"app"`
	Fields []string `json:"fields,omitempty"`
	Query  string   `json:"query,omitempty"`
	Size   int64    `json:"size,omitempty"`
}

// Cursor is a cursor to get records. Warning is the value of the
// X-Cybozu-Warning response header.
type Cursor struct {
	ID         string `json:"id"`
	TotalCount int64  `json:"totalCount,string"`
	Warning    string `json:"-"`
}

// CreateCursor creates a cursor.
func (c *Client) CreateCursor(ctx context.Context, req CreateCursorRequest) (*Cursor, error) {
	result, header, err := c.doHeader[Cursor](ctx, http.MethodPost, "records/cursor.json", req)
	if err != nil {
		return nil, err
	}
	result.Warning = header.Get(warningHeader)
	return &result, nil
}

// GetRecordsByCursor gets the next records of the cursor.
// next reports whether there are more records to get.
func (c *Client) GetRecordsByCursor(ctx context.Context, id string) (records []Record, next bool, err error) {
	params := struct {
		ID string `json:"id"`
	}{id}
	result, err := c.get[struct {
		Records []Record `json:"records"`
		Next    bool     `json:"next"`
	}](ctx, "records/cursor.json", params)
	if err != nil {
		return nil, false, err
	}
	return result.Records, result.Next, nil
}

// DeleteCursor deletes the cursor.
func (c *Client) DeleteCursor(ctx context.Context, id string) error {
	params := struct {
		ID string `json:"id"`
	}{id}
	_, err := c.delete[struct{}](ctx, "records/cursor.json", params)
	return err
}

// AllRecords iterates over the records using a cursor. If Size of the request
// is zero, 500 is used. The cursor is deleted when the iteration stops before
// getting all the records. If the cursor has a warning, a [*WarningError] is
// yielded after all the records.
//
//	for record, err := range client.AllRecords(ctx, kintone.CreateCursorRequest{App: app}) {
//		if err != nil {
//			return err
//		}
//		fmt.Println(record["$id"].Value)
//	}
func (c *Client) AllRecords(ctx context.Context, req CreateCursorRequest) iter.Seq2[Record, error] {
	return func(yield func(Record, error) bool) {
		if req.Size == 0 {
			req.Size = 500
		}
		cursor, err := c.CreateCursor(ctx, req)
		if err != nil {
			yield(nil, err)
			return
		}
		next := true
		defer func() {
			if next {
				ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
				defer cancel()
				_ = c.DeleteCursor(ctx, cursor.ID)
			}
		}()
		for next {
			var records []Record
			records, next, err = c.GetRecordsByCursor(ctx, cursor.ID)
			if err != nil {
				next = true
				yield(nil, err)
				return
			}
			for _, record := range records {
				if !yield(record, nil) {
					return
				}
			}
		}
		if cursor.Warning != "" {
			yield(nil, &WarningError{Message: cursor.Warning})
		}
	}
}

const warningHeader = "X-Cybozu-Warning"

// WarningError is a warning in the X-Cybozu-Warning response header.
type WarningError struct {
	Message string
}

func (e *WarningError) Error() string {
	return "kintone: " + e.Message
}

// AllRecordsByIDRequest is the request of [Client.AllRecordsByID].
// Condition is the query condition, without order by, limit and offset.
type AllRecordsByIDRequest struct {
	App       int64
	Fields    []string
	Condition string
}

var allRecordsByIDLimit = 500

// AllRecordsByID iterates over the records in ascending order of the record
// ID, using [Client.GetRecords] with the condition on the record ID instead of
// a cursor. If Fields of the request does not contain $id, it is added. If any
// of the responses has a warning, a [*WarningError] is yielded after all the
// records.
func (c *Client) AllRecordsByID(ctx context.Context, req AllRecordsByIDRequest) iter.Seq2[Record, error] {
	return func(yield func(Record, error) bool) {
		fields := req.Fields
		if len(fields) > 0 && !slices.Contains(fields, "$id") {
			fields = append(slices.Clip(fields), "$id")
		}
		var condition string
		if req.Condition != "" {
			condition = "(" + req.Condition + ") and "
		}
		lastID, warning := "0", ""
		for {
			res, err := c.GetRecords(ctx, GetRecordsRequest{
				App:    req.App,
				Fields: fields,
				Query:  fmt.Sprintf("%s$id > %s order by $id asc limit %d", condition, lastID, allRecordsByIDLimit),
			})
			if err != nil {
				yield(nil, err)
				return
			}
			if warning == "" {
				warning = res.Warning
			}
			for _, record := range res.Records {
				id, ok := record["$id"].Value.(string)
				if !ok {
					yield(nil, errors.New("kintone: record without $id"))
					return
				}
				lastID = id
				if !yield(record, nil) {
					return
				}
			}
			if len(res.Records) < allRecordsByIDLimit {
				break
			}
		}
		if warning != "" {
			yield(nil, &WarningError{Message: warning})
		}
	}
}

func paginate[T any](offset int64, get func(offset int64) ([]T, bool, error)) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for {
			items, more, err := get(offset)
			if err != nil {
				var zero T
				yield(zero, err)
				return
			}
			for _, item := range items {
				if !yield(item, nil) {
					return
				}
			}
			if !more || len(items) == 0 {
				return
			}
			offset += int64(len(items))
		}
	}
}
