package kintone

import (
	"context"
	"encoding/json/jsontext"
	"fmt"
	"net/http"
	"slices"
	"strings"
)

// BulkRequest is a bulk request, which sends the requests at once. Use the
// methods to add requests, and [BulkRequest.Send] to send them. The results
// of the requests are available after sending the bulk request.
type BulkRequest struct {
	client   *Client
	requests []bulkItem
}

type bulkItem struct {
	method string
	api    string
	params any
	set    func(jsontext.Value) error
}

// BulkResult is the result of a request in a bulk request.
type BulkResult[T any] struct {
	value T
}

// Value returns the result of the request. It returns the zero value before
// the bulk request is sent successfully.
func (r *BulkResult[T]) Value() T {
	return r.value
}

// NewBulkRequest creates a bulk request.
func (c *Client) NewBulkRequest() *BulkRequest {
	return &BulkRequest{client: c}
}

func (b *BulkRequest) add[T any](r apiRequest[T]) *BulkResult[T] {
	result := &BulkResult[T]{}
	b.requests = append(b.requests, bulkItem{
		method: r.method,
		api:    b.client.apiPath(r.path),
		params: r.params,
		set: func(res jsontext.Value) error {
			if r.decode == nil {
				return nil
			}
			value, err := r.decode(res)
			if err != nil {
				return err
			}
			result.value = value
			return nil
		},
	})
	return result
}

// AddRecord adds a request to add a record. See [Client.AddRecord].
func (b *BulkRequest) AddRecord(app int64, record Record) *BulkResult[RecordRevision] {
	return b.add(addRecordRequest(app, record))
}

// AddRecords adds a request to add records. See [Client.AddRecords].
func (b *BulkRequest) AddRecords(app int64, records []Record) *BulkResult[[]RecordRevision] {
	return b.add(addRecordsRequest(app, records))
}

// UpdateRecord adds a request to update a record. See [Client.UpdateRecord].
func (b *BulkRequest) UpdateRecord(app int64, update RecordUpdate) *BulkResult[int64] {
	return b.add(updateRecordRequest(app, update))
}

// UpdateRecords adds a request to update records. See [Client.UpdateRecords].
func (b *BulkRequest) UpdateRecords(app int64, updates []RecordUpdate) *BulkResult[[]UpdatedRecord] {
	return b.add(updateRecordsRequest(app, updates, false))
}

// UpsertRecords adds a request to upsert records. See [Client.UpsertRecords].
func (b *BulkRequest) UpsertRecords(app int64, updates []RecordUpdate) *BulkResult[[]UpdatedRecord] {
	return b.add(updateRecordsRequest(app, updates, true))
}

// DeleteRecords adds a request to delete records. See [Client.DeleteRecords].
func (b *BulkRequest) DeleteRecords(app int64, ids, revisions []int64) *BulkResult[struct{}] {
	return b.add(deleteRecordsRequest(app, ids, revisions))
}

// UpdateStatus adds a request to update the status of a record. See
// [Client.UpdateStatus].
func (b *BulkRequest) UpdateStatus(app int64, update StatusUpdate) *BulkResult[int64] {
	return b.add(updateStatusRequest(app, update))
}

// UpdateStatuses adds a request to update the statuses of records. See
// [Client.UpdateStatuses].
func (b *BulkRequest) UpdateStatuses(app int64, updates []StatusUpdate) *BulkResult[[]RecordRevision] {
	return b.add(updateStatusesRequest(app, updates))
}

// UpdateAssignees adds a request to update the assignees of a record. See
// [Client.UpdateAssignees].
func (b *BulkRequest) UpdateAssignees(app int64, update AssigneesUpdate) *BulkResult[int64] {
	return b.add(updateAssigneesRequest(app, update))
}

// Send sends the bulk request. If a request in the bulk request fails, it
// returns a [*BulkError] with the index of the request. Other errors are
// returned as is.
func (b *BulkRequest) Send(ctx context.Context) error {
	type request struct {
		Method  string `json:"method"`
		API     string `json:"api"`
		Payload any    `json:"payload"`
	}
	requests := make([]request, len(b.requests))
	for i, r := range b.requests {
		requests[i] = request{r.method, r.api, r.params}
	}
	params := struct {
		Requests []request `json:"requests"`
	}{requests}
	result, err := b.client.do[struct {
		Results []jsontext.Value `json:"results"`
	}](ctx, http.MethodPost, "bulkRequest.json", params)
	if err != nil {
		return err
	}
	if len(result.Results) != len(b.requests) {
		return fmt.Errorf("kintone: decode response: got %d results for %d requests",
			len(result.Results), len(b.requests))
	}
	for i, r := range b.requests {
		if err := r.set(result.Results[i]); err != nil {
			return err
		}
	}
	return nil
}

var (
	allRecordsChunkSize = 100
	allRecordsBulkSize  = 20
)

// AllRecordsError is an error of the methods to process all records, such as
// [Client.AddAllRecords]. Processed is the number of the records processed
// before the error, and Err is the error of the bulk request.
type AllRecordsError struct {
	Processed int
	Err       error
}

func (e *AllRecordsError) Error() string {
	return fmt.Sprintf("kintone: processed %d records: %s", e.Processed, strings.TrimPrefix(e.Err.Error(), "kintone: "))
}

// Unwrap returns the error of the bulk request.
func (e *AllRecordsError) Unwrap() error {
	return e.Err
}

// AddAllRecords adds records by splitting them into bulk requests. If a bulk
// request fails, it returns the records added by the earlier bulk requests and
// an [*AllRecordsError].
func (c *Client) AddAllRecords(ctx context.Context, app int64, records []Record) ([]RecordRevision, error) {
	results, err := c.sendAllRecords(ctx, len(records), func(i, j int) apiRequest[[]RecordRevision] {
		return addRecordsRequest(app, records[i:j])
	})
	return slices.Concat(results...), err
}

// UpdateAllRecords updates records by splitting them into bulk requests. If a
// bulk request fails, it returns the records updated by the earlier bulk
// requests and an [*AllRecordsError].
func (c *Client) UpdateAllRecords(ctx context.Context, app int64, updates []RecordUpdate) ([]UpdatedRecord, error) {
	results, err := c.sendAllRecords(ctx, len(updates), func(i, j int) apiRequest[[]UpdatedRecord] {
		return updateRecordsRequest(app, updates[i:j], false)
	})
	return slices.Concat(results...), err
}

// UpsertAllRecords upserts records by splitting them into bulk requests. If a
// bulk request fails, it returns the records upserted by the earlier bulk
// requests and an [*AllRecordsError].
func (c *Client) UpsertAllRecords(ctx context.Context, app int64, updates []RecordUpdate) ([]UpdatedRecord, error) {
	results, err := c.sendAllRecords(ctx, len(updates), func(i, j int) apiRequest[[]UpdatedRecord] {
		return updateRecordsRequest(app, updates[i:j], true)
	})
	return slices.Concat(results...), err
}

// DeleteAllRecords deletes records by splitting them into bulk requests. The
// revisions are not checked if revisions is nil. If a bulk request fails, it
// returns an [*AllRecordsError].
func (c *Client) DeleteAllRecords(ctx context.Context, app int64, ids, revisions []int64) error {
	if revisions != nil && len(revisions) != len(ids) {
		return fmt.Errorf("kintone: got %d revisions for %d ids", len(revisions), len(ids))
	}
	_, err := c.sendAllRecords(ctx, len(ids), func(i, j int) apiRequest[struct{}] {
		if revisions == nil {
			return deleteRecordsRequest(app, ids[i:j], nil)
		}
		return deleteRecordsRequest(app, ids[i:j], revisions[i:j])
	})
	return err
}

func (c *Client) sendAllRecords[T any](ctx context.Context, n int, request func(i, j int) apiRequest[T]) ([]T, error) {
	var values []T
	for i := 0; i < n; {
		bulk := c.NewBulkRequest()
		var results []*BulkResult[T]
		j := i
		for range allRecordsBulkSize {
			if j == n {
				break
			}
			k := min(j+allRecordsChunkSize, n)
			results = append(results, bulk.add(request(j, k)))
			j = k
		}
		if err := bulk.Send(ctx); err != nil {
			return values, &AllRecordsError{Processed: i, Err: err}
		}
		for _, r := range results {
			values = append(values, r.Value())
		}
		i = j
	}
	return values, nil
}
