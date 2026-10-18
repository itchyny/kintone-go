package kintone

import "context"

// RecordAccess is the access rights of the user on a record and its fields.
type RecordAccess struct {
	ID     int64                        `json:"id,string"`
	Record RecordAccessRights           `json:"record"`
	Fields map[string]FieldAccessRights `json:"fields"`
}

// RecordAccessRights is the access rights on a record.
type RecordAccessRights struct {
	Viewable  bool `json:"viewable"`
	Editable  bool `json:"editable"`
	Deletable bool `json:"deletable"`
}

// FieldAccessRights is the access rights on a field.
type FieldAccessRights struct {
	Viewable bool `json:"viewable"`
	Editable bool `json:"editable"`
}

// EvaluateRecordPermissions gets the permissions of the user on the records.
func (c *Client) EvaluateRecordPermissions(ctx context.Context, app int64, ids []int64) ([]RecordAccess, error) {
	params := struct {
		App int64   `json:"app"`
		IDs []int64 `json:"ids"`
	}{app, ids}
	result, err := c.get[struct {
		Rights []RecordAccess `json:"rights"`
	}](ctx, "records/acl/evaluate.json", params)
	if err != nil {
		return nil, err
	}
	return result.Rights, nil
}
