package kintone

import (
	"context"
	"iter"
	"time"
)

// CreateCursorRequest is the request of [Client.CreateCursor].
type CreateCursorRequest struct {
	App    int64    `json:"app"`
	Fields []string `json:"fields,omitempty"`
	Query  string   `json:"query,omitempty"`
	Size   int64    `json:"size,omitempty"`
}

// Cursor is a cursor to get records.
type Cursor struct {
	ID         string `json:"id"`
	TotalCount int64  `json:"totalCount,string"`
}

// CreateCursor creates a cursor.
func (c *Client) CreateCursor(ctx context.Context, req CreateCursorRequest) (*Cursor, error) {
	result, err := c.post[Cursor](ctx, "records/cursor.json", req)
	if err != nil {
		return nil, err
	}
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
// getting all the records.
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
	}
}
