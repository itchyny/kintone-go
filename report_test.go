package kintone

import (
	"net/http"
	"reflect"
	"testing"
)

func TestClientGetReports(t *testing.T) {
	testCases := []struct {
		name     string
		req      GetAppSettingsRequest
		wantPath string
		wantBody string
	}{
		{
			name:     "production",
			req:      GetAppSettingsRequest{App: 1, Lang: "user"},
			wantPath: "/k/v1/app/reports.json",
			wantBody: `{"app": 1, "lang": "user"}`,
		},
		{
			name:     "preview",
			req:      GetAppSettingsRequest{App: 1, Preview: true},
			wantPath: "/k/v1/preview/app/reports.json",
			wantBody: `{"app": 1}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{
				"reports": {
					"Monthly sales": {
						"chartType": "COLUMN",
						"chartMode": "STACKED",
						"id": "20",
						"name": "Monthly sales",
						"index": "0",
						"groups": [{"code": "date", "per": "MONTH"}, {"code": "category"}],
						"aggregations": [{"type": "SUM", "code": "amount"}],
						"filterCond": "",
						"sorts": [{"by": "GROUP1", "order": "ASC"}],
						"periodicReport": {
							"active": true,
							"period": {"every": "MONTH", "dayOfMonth": "END_OF_MONTH", "time": "09:00"}
						}
					},
					"Count": {
						"chartType": "PIE",
						"id": "21",
						"name": "Count",
						"index": "1",
						"groups": [{"code": "status"}],
						"aggregations": [{"type": "COUNT"}],
						"filterCond": "",
						"sorts": [{"by": "TOTAL", "order": "DESC"}],
						"periodicReport": null
					}
				},
				"revision": "6"
			}`)
			c := newTestClient(t, ts)
			res, err := c.GetReports(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("GetReports returned error: %v", err)
			}
			assertRequest(t, got, http.MethodGet, tc.wantPath, tc.wantBody)
			want := &Reports{
				Reports: map[string]Report{
					"Monthly sales": {
						ID:           20,
						Name:         "Monthly sales",
						ChartType:    ChartTypeColumn,
						ChartMode:    ChartModeStacked,
						Index:        0,
						Groups:       []ReportGroup{{Code: "date", Per: "MONTH"}, {Code: "category"}},
						Aggregations: []ReportAggregation{{Type: AggregationTypeSum, Code: "amount"}},
						Sorts:        []ReportSort{{By: "GROUP1", Order: "ASC"}},
						PeriodicReport: &PeriodicReport{
							Active: new(true),
							Period: ReportPeriod{Every: "MONTH", DayOfMonth: "END_OF_MONTH", Time: "09:00"},
						},
					},
					"Count": {
						ID:           21,
						Name:         "Count",
						ChartType:    ChartTypePie,
						Index:        1,
						Groups:       []ReportGroup{{Code: "status"}},
						Aggregations: []ReportAggregation{{Type: AggregationTypeCount}},
						Sorts:        []ReportSort{{By: "TOTAL", Order: "DESC"}},
					},
				},
				Revision: 6,
			}
			if !reflect.DeepEqual(res, want) {
				t.Errorf("response: got %+v, want %+v", res, want)
			}
		})
	}
}

func TestClientUpdateReports(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{
		"revision": "7",
		"reports": {"Count": {"id": "21"}}
	}`)
	c := newTestClient(t, ts)
	res, err := c.UpdateReports(t.Context(), 1, map[string]Report{
		"Count": {
			Name:         "Count",
			ChartType:    ChartTypeTable,
			Index:        0,
			Groups:       []ReportGroup{{Code: "status"}},
			Aggregations: []ReportAggregation{{Type: AggregationTypeCount}},
			Sorts:        []ReportSort{{By: "TOTAL", Order: "DESC"}},
			PeriodicReport: &PeriodicReport{
				Active: new(false),
				Period: ReportPeriod{Every: "HOUR", Minute: "30"},
			},
		},
	}, 0)
	if err != nil {
		t.Fatalf("UpdateReports returned error: %v", err)
	}
	assertRequest(t, got, http.MethodPut, "/k/v1/preview/app/reports.json", `{
		"app": 1,
		"reports": {
			"Count": {
				"name": "Count",
				"chartType": "TABLE",
				"index": "0",
				"groups": [{"code": "status"}],
				"aggregations": [{"type": "COUNT"}],
				"sorts": [{"by": "TOTAL", "order": "DESC"}],
				"periodicReport": {"active": false, "period": {"every": "HOUR", "minute": "30"}}
			}
		}
	}`)
	want := &UpdatedReports{
		Reports:  map[string]UpdatedReport{"Count": {ID: 21}},
		Revision: 7,
	}
	if !reflect.DeepEqual(res, want) {
		t.Errorf("response: got %+v, want %+v", res, want)
	}
}
