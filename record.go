package kintone

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode"
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
// Use [Quote] to quote string values in Query.
type GetRecordsRequest struct {
	App        int64    `json:"app"`
	Fields     []string `json:"fields,omitzero"`
	Query      string   `json:"query,omitzero"`
	TotalCount bool     `json:"totalCount,omitzero"`
}

// GetRecordsResponse is the response of [Client.GetRecords].
// TotalCount is set only when TotalCount of the request is true. Warning is
// the value of the X-Cybozu-Warning response header.
type GetRecordsResponse struct {
	Records    []Record `json:"records"`
	TotalCount int64    `json:"totalCount,string"`
	Warning    string   `json:"-"`
}

// GetRecords gets records matching the query.
func (c *Client) GetRecords(ctx context.Context, req GetRecordsRequest) (*GetRecordsResponse, error) {
	result, header, err := c.doHeader[GetRecordsResponse](ctx, http.MethodGet, "records.json", req)
	if err != nil {
		return nil, err
	}
	result.Warning = header.Get(warningHeader)
	return &result, nil
}

// Quote returns a double-quoted string for a value in a query, escaping double
// quotes, backslashes and control characters.
func Quote(s string) string {
	var sb strings.Builder
	sb.Grow(len(s) + 2)
	sb.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"', '\\':
			sb.WriteByte('\\')
			sb.WriteRune(r)
		case '\t':
			sb.WriteString(`\t`)
		case '\n':
			sb.WriteString(`\n`)
		case '\r':
			sb.WriteString(`\r`)
		default:
			if unicode.IsControl(r) {
				fmt.Fprintf(&sb, `\u%04x`, r)
			} else {
				sb.WriteRune(r)
			}
		}
	}
	sb.WriteByte('"')
	return sb.String()
}

// RecordRevision is the ID and the revision of a record.
type RecordRevision struct {
	ID       int64 `json:"id,string"`
	Revision int64 `json:"revision,string"`
}

// AddRecord adds a record.
func (c *Client) AddRecord(ctx context.Context, app int64, record Record) (RecordRevision, error) {
	return c.call(ctx, addRecordRequest(app, record))
}

func addRecordRequest(app int64, record Record) apiRequest[RecordRevision] {
	params := struct {
		App    int64  `json:"app"`
		Record Record `json:"record,omitzero"`
	}{app, record}
	return apiRequest[RecordRevision]{http.MethodPost, "record.json", params,
		decodeWith(func(r RecordRevision) (RecordRevision, error) { return r, nil })}
}

// AddRecords adds records.
func (c *Client) AddRecords(ctx context.Context, app int64, records []Record) ([]RecordRevision, error) {
	return c.call(ctx, addRecordsRequest(app, records))
}

func addRecordsRequest(app int64, records []Record) apiRequest[[]RecordRevision] {
	params := struct {
		App     int64    `json:"app"`
		Records []Record `json:"records"`
	}{app, records}
	return apiRequest[[]RecordRevision]{http.MethodPost, "records.json", params, decodeWith(decodeRecordRevisions)}
}

func decodeRecordRevisions(r struct {
	IDs       []string `json:"ids"`
	Revisions []string `json:"revisions"`
}) ([]RecordRevision, error) {
	revisions := make([]RecordRevision, len(r.IDs))
	for i, id := range r.IDs {
		var err error
		if revisions[i].ID, err = strconv.ParseInt(id, 10, 64); err != nil {
			return nil, fmt.Errorf("kintone: decode response: %w", err)
		}
		if i < len(r.Revisions) {
			if revisions[i].Revision, err = strconv.ParseInt(r.Revisions[i], 10, 64); err != nil {
				return nil, fmt.Errorf("kintone: decode response: %w", err)
			}
		}
	}
	return revisions, nil
}

// RecordUpdate is an update of a record, specified by either ID or UpdateKey.
// The revision is not checked if Revision is zero.
type RecordUpdate struct {
	ID        int64      `json:"id,omitzero"`
	UpdateKey *UpdateKey `json:"updateKey,omitzero"`
	Record    Record     `json:"record,omitzero"`
	Revision  int64      `json:"revision,omitzero"`
}

// UpdateKey specifies a record by the value of a field.
type UpdateKey struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

// UpdateRecord updates a record and returns the new revision.
func (c *Client) UpdateRecord(ctx context.Context, app int64, update RecordUpdate) (int64, error) {
	return c.call(ctx, updateRecordRequest(app, update))
}

func updateRecordRequest(app int64, update RecordUpdate) apiRequest[int64] {
	params := struct {
		App int64 `json:"app"`
		RecordUpdate
	}{app, update}
	return apiRequest[int64]{http.MethodPut, "record.json", params, decodeWith(decodeRevision)}
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
	return c.call(ctx, updateRecordsRequest(app, updates, false))
}

// UpsertRecords updates records, or adds them if they do not exist.
// To upsert a single record, call it with one update.
func (c *Client) UpsertRecords(ctx context.Context, app int64, updates []RecordUpdate) ([]UpdatedRecord, error) {
	return c.call(ctx, updateRecordsRequest(app, updates, true))
}

func updateRecordsRequest(app int64, updates []RecordUpdate, upsert bool) apiRequest[[]UpdatedRecord] {
	params := struct {
		App     int64          `json:"app"`
		Records []RecordUpdate `json:"records"`
		Upsert  bool           `json:"upsert,omitzero"`
	}{app, updates, upsert}
	return apiRequest[[]UpdatedRecord]{http.MethodPut, "records.json", params,
		decodeWith(func(r struct {
			Records []UpdatedRecord `json:"records"`
		}) ([]UpdatedRecord, error) {
			return r.Records, nil
		})}
}

// DeleteRecords deletes records.
// The revisions are not checked if revisions is nil.
func (c *Client) DeleteRecords(ctx context.Context, app int64, ids, revisions []int64) error {
	_, err := c.call(ctx, deleteRecordsRequest(app, ids, revisions))
	return err
}

func deleteRecordsRequest(app int64, ids, revisions []int64) apiRequest[struct{}] {
	params := struct {
		App       int64   `json:"app"`
		IDs       []int64 `json:"ids"`
		Revisions []int64 `json:"revisions,omitzero"`
	}{app, ids, revisions}
	return apiRequest[struct{}]{http.MethodDelete, "records.json", params, nil}
}
