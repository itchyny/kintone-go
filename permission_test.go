package kintone

import (
	"net/http"
	"reflect"
	"testing"
)

func TestClientEvaluateRecordPermissions(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{
		"rights": [
			{
				"id": "2",
				"record": {"viewable": true, "editable": false, "deletable": false},
				"fields": {
					"title": {"viewable": true, "editable": false},
					"note": {"viewable": false, "editable": false}
				}
			},
			{
				"id": "3",
				"record": {"viewable": true, "editable": true, "deletable": true},
				"fields": {
					"title": {"viewable": true, "editable": true},
					"note": {"viewable": true, "editable": true}
				}
			}
		]
	}`)
	c := newTestClient(t, ts)
	res, err := c.EvaluateRecordPermissions(t.Context(), 1, []int64{2, 3})
	if err != nil {
		t.Fatalf("EvaluateRecordPermissions returned error: %v", err)
	}
	assertRequest(t, got, http.MethodGet, "/k/v1/records/acl/evaluate.json", `{"app": 1, "ids": [2, 3]}`)
	want := []RecordPermission{
		{
			ID:     2,
			Record: RecordRights{Viewable: true},
			Fields: map[string]FieldRights{
				"title": {Viewable: true},
				"note":  {},
			},
		},
		{
			ID:     3,
			Record: RecordRights{Viewable: true, Editable: true, Deletable: true},
			Fields: map[string]FieldRights{
				"title": {Viewable: true, Editable: true},
				"note":  {Viewable: true, Editable: true},
			},
		},
	}
	if !reflect.DeepEqual(res, want) {
		t.Errorf("response: got %+v, want %+v", res, want)
	}
}
