package kintone

import "context"

// Report is the setting of a graph of an app.
type Report struct {
	ID             int64               `json:"id,string,omitempty"`
	Name           string              `json:"name,omitempty"`
	ChartType      ChartType           `json:"chartType"`
	ChartMode      ChartMode           `json:"chartMode,omitempty"`
	Index          int                 `json:"index,string"`
	Groups         []ReportGroup       `json:"groups,omitempty"`
	Aggregations   []ReportAggregation `json:"aggregations,omitempty"`
	FilterCond     string              `json:"filterCond,omitempty"`
	Sorts          []ReportSort        `json:"sorts,omitempty"`
	PeriodicReport *PeriodicReport     `json:"periodicReport,omitempty"`
}

// ChartType is the type of a graph.
type ChartType string

// Chart types.
const (
	ChartTypeBar        ChartType = "BAR"
	ChartTypeColumn     ChartType = "COLUMN"
	ChartTypePie        ChartType = "PIE"
	ChartTypeLine       ChartType = "LINE"
	ChartTypePivotTable ChartType = "PIVOT_TABLE"
	ChartTypeTable      ChartType = "TABLE"
	ChartTypeArea       ChartType = "AREA"
	ChartTypeSpline     ChartType = "SPLINE"
	ChartTypeSplineArea ChartType = "SPLINE_AREA"
)

// ChartMode is the display mode of a graph.
type ChartMode string

// Chart modes.
const (
	ChartModeNormal     ChartMode = "NORMAL"
	ChartModeStacked    ChartMode = "STACKED"
	ChartModePercentage ChartMode = "PERCENTAGE"
)

// ReportGroup is a group of a graph. Per is the time unit of a date or time
// field.
type ReportGroup struct {
	Code string `json:"code"`
	Per  string `json:"per,omitempty"`
}

// ReportAggregation is an aggregation of a graph. Code is empty for COUNT.
type ReportAggregation struct {
	Type AggregationType `json:"type"`
	Code string          `json:"code,omitempty"`
}

// AggregationType is the type of an aggregation.
type AggregationType string

// Aggregation types.
const (
	AggregationTypeCount   AggregationType = "COUNT"
	AggregationTypeSum     AggregationType = "SUM"
	AggregationTypeAverage AggregationType = "AVERAGE"
	AggregationTypeMax     AggregationType = "MAX"
	AggregationTypeMin     AggregationType = "MIN"
)

// ReportSort is a sort of a graph.
type ReportSort struct {
	By    string `json:"by"`
	Order string `json:"order"`
}

// PeriodicReport is the setting of the periodic report of a graph.
type PeriodicReport struct {
	Active *bool        `json:"active,omitempty"`
	Period ReportPeriod `json:"period"`
}

// ReportPeriod is the period of a periodic report.
type ReportPeriod struct {
	Every      string `json:"every"`
	Month      string `json:"month,omitempty"`
	Time       string `json:"time,omitempty"`
	Pattern    string `json:"pattern,omitempty"`
	DayOfMonth string `json:"dayOfMonth,omitempty"`
	DayOfWeek  string `json:"dayOfWeek,omitempty"`
	Minute     string `json:"minute,omitempty"`
}

// Reports is the graphs of an app, keyed by the graph name.
type Reports struct {
	Reports  map[string]Report `json:"reports"`
	Revision int64             `json:"revision,string"`
}

// GetReports gets the graphs of an app.
func (c *Client) GetReports(ctx context.Context, req GetAppSettingsRequest) (*Reports, error) {
	return c.getAppSettings[Reports](ctx, "app/reports.json", req.Preview, req)
}

// UpdatedReports is the response of [Client.UpdateReports], which has the
// graph IDs keyed by the graph name.
type UpdatedReports struct {
	Reports  map[string]UpdatedReport `json:"reports"`
	Revision int64                    `json:"revision,string"`
}

// UpdatedReport is a graph updated by [Client.UpdateReports].
type UpdatedReport struct {
	ID int64 `json:"id,string"`
}

// UpdateReports updates the graphs of an app in the preview environment. The
// keys of reports are the graph names before the update. The revision is not
// checked if revision is zero.
func (c *Client) UpdateReports(ctx context.Context, app int64, reports map[string]Report, revision int64) (*UpdatedReports, error) {
	params := struct {
		App      int64             `json:"app"`
		Reports  map[string]Report `json:"reports"`
		Revision int64             `json:"revision,omitempty"`
	}{app, reports, revision}
	result, err := c.put[UpdatedReports](ctx, "preview/app/reports.json", params)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
