package kintone

import "context"

// RecordPermission is the permissions of the user on a record and its fields.
type RecordPermission struct {
	ID     int64                  `json:"id,string"`
	Record RecordRights           `json:"record"`
	Fields map[string]FieldRights `json:"fields"`
}

// RecordRights is the rights on a record.
type RecordRights struct {
	Viewable  bool `json:"viewable"`
	Editable  bool `json:"editable"`
	Deletable bool `json:"deletable"`
}

// FieldRights is the rights on a field.
type FieldRights struct {
	Viewable bool `json:"viewable"`
	Editable bool `json:"editable"`
}

// EvaluateRecordPermissions gets the permissions of the user on the records.
func (c *Client) EvaluateRecordPermissions(ctx context.Context, app int64, ids []int64) ([]RecordPermission, error) {
	params := struct {
		App int64   `json:"app"`
		IDs []int64 `json:"ids"`
	}{app, ids}
	result, err := c.get[struct {
		Rights []RecordPermission `json:"rights"`
	}](ctx, "records/acl/evaluate.json", params)
	if err != nil {
		return nil, err
	}
	return result.Rights, nil
}
