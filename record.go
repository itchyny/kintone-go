package kintone

import (
	"context"
	"fmt"
	"strconv"
)

// Record is a kintone record, keyed by the field code.
type Record map[string]Field

// GetRecord gets a record.
func (c *Client) GetRecord(ctx context.Context, app, id int64) (Record, error) {
	params := struct {
		App int64 `json:"app"`
		ID  int64 `json:"id"`
	}{app, id}
	result, err := c.get[struct {
		Record Record `json:"record"`
	}](ctx, "record.json", params)
	if err != nil {
		return nil, err
	}
	return result.Record, nil
}

// GetRecordsRequest is the request of [Client.GetRecords].
type GetRecordsRequest struct {
	App        int64    `json:"app"`
	Fields     []string `json:"fields,omitempty"`
	Query      string   `json:"query,omitempty"`
	TotalCount bool     `json:"totalCount,omitempty"`
}

// GetRecordsResponse is the response of [Client.GetRecords].
// TotalCount is set only when TotalCount of the request is true.
type GetRecordsResponse struct {
	Records    []Record `json:"records"`
	TotalCount int64    `json:"totalCount,string"`
}

// GetRecords gets records matching the query.
func (c *Client) GetRecords(ctx context.Context, req GetRecordsRequest) (*GetRecordsResponse, error) {
	result, err := c.get[GetRecordsResponse](ctx, "records.json", req)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// RecordRevision is the ID and the revision of a record.
type RecordRevision struct {
	ID       int64 `json:"id,string"`
	Revision int64 `json:"revision,string"`
}

// AddRecord adds a record.
func (c *Client) AddRecord(ctx context.Context, app int64, record Record) (RecordRevision, error) {
	params := struct {
		App    int64  `json:"app"`
		Record Record `json:"record,omitempty"`
	}{app, record}
	return c.post[RecordRevision](ctx, "record.json", params)
}

// AddRecords adds records.
func (c *Client) AddRecords(ctx context.Context, app int64, records []Record) ([]RecordRevision, error) {
	params := struct {
		App     int64    `json:"app"`
		Records []Record `json:"records"`
	}{app, records}
	result, err := c.post[struct {
		IDs       []string `json:"ids"`
		Revisions []string `json:"revisions"`
	}](ctx, "records.json", params)
	if err != nil {
		return nil, err
	}
	revisions := make([]RecordRevision, len(result.IDs))
	for i, id := range result.IDs {
		if revisions[i].ID, err = strconv.ParseInt(id, 10, 64); err != nil {
			return nil, fmt.Errorf("kintone: decode response: %w", err)
		}
		if i < len(result.Revisions) {
			if revisions[i].Revision, err = strconv.ParseInt(result.Revisions[i], 10, 64); err != nil {
				return nil, fmt.Errorf("kintone: decode response: %w", err)
			}
		}
	}
	return revisions, nil
}

// RecordUpdate is an update of a record, specified by either ID or UpdateKey.
// The revision is not checked if Revision is zero.
type RecordUpdate struct {
	ID        int64      `json:"id,omitempty"`
	UpdateKey *UpdateKey `json:"updateKey,omitempty"`
	Record    Record     `json:"record,omitempty"`
	Revision  int64      `json:"revision,omitempty"`
}

// UpdateKey specifies a record by the value of a field.
type UpdateKey struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

// UpdateRecord updates a record and returns the new revision.
func (c *Client) UpdateRecord(ctx context.Context, app int64, update RecordUpdate) (int64, error) {
	params := struct {
		App int64 `json:"app"`
		RecordUpdate
	}{app, update}
	result, err := c.put[struct {
		Revision int64 `json:"revision,string"`
	}](ctx, "record.json", params)
	if err != nil {
		return 0, err
	}
	return result.Revision, nil
}

// UpdatedRecord is a record updated by [Client.UpdateRecords] or
// [Client.UpsertRecords].
type UpdatedRecord struct {
	ID        int64     `json:"id,string"`
	Revision  int64     `json:"revision,string"`
	Operation Operation `json:"operation"`
}

// Operation is the operation done on a record by [Client.UpsertRecords].
type Operation string

// Operations.
const (
	OperationInsert Operation = "INSERT"
	OperationUpdate Operation = "UPDATE"
)

// UpdateRecords updates records.
func (c *Client) UpdateRecords(ctx context.Context, app int64, updates []RecordUpdate) ([]UpdatedRecord, error) {
	return c.updateRecords(ctx, app, updates, false)
}

// UpsertRecords updates records, or adds them if they do not exist.
// To upsert a single record, call it with one update.
func (c *Client) UpsertRecords(ctx context.Context, app int64, updates []RecordUpdate) ([]UpdatedRecord, error) {
	return c.updateRecords(ctx, app, updates, true)
}

func (c *Client) updateRecords(ctx context.Context, app int64, updates []RecordUpdate, upsert bool) ([]UpdatedRecord, error) {
	params := struct {
		App     int64          `json:"app"`
		Records []RecordUpdate `json:"records"`
		Upsert  bool           `json:"upsert,omitempty"`
	}{app, updates, upsert}
	result, err := c.put[struct {
		Records []UpdatedRecord `json:"records"`
	}](ctx, "records.json", params)
	if err != nil {
		return nil, err
	}
	return result.Records, nil
}

// DeleteRecords deletes records.
// The revisions are not checked if revisions is nil.
func (c *Client) DeleteRecords(ctx context.Context, app int64, ids, revisions []int64) error {
	params := struct {
		App       int64   `json:"app"`
		IDs       []int64 `json:"ids"`
		Revisions []int64 `json:"revisions,omitempty"`
	}{app, ids, revisions}
	_, err := c.delete[struct{}](ctx, "records.json", params)
	return err
}
