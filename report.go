package kintone

import "context"

// Report is the setting of a graph of an app.
type Report struct {
	ID             int64               `json:"id,string,omitzero"`
	Name           string              `json:"name,omitzero"`
	ChartType      ChartType           `json:"chartType"`
	ChartMode      ChartMode           `json:"chartMode,omitzero"`
	Index          int                 `json:"index,string"`
	Groups         []ReportGroup       `json:"groups,omitzero"`
	Aggregations   []ReportAggregation `json:"aggregations,omitzero"`
	FilterCond     *string             `json:"filterCond,omitzero"`
	Sorts          []ReportSort        `json:"sorts,omitzero"`
	PeriodicReport *PeriodicReport     `json:"periodicReport,omitzero"`
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
	Code string   `json:"code"`
	Per  TimeUnit `json:"per,omitzero"`
}

// TimeUnit is a unit of time, used for the groups and the periodic reports of graphs.
type TimeUnit string

// Time units.
const (
	TimeUnitYear    TimeUnit = "YEAR"
	TimeUnitQuarter TimeUnit = "QUARTER"
	TimeUnitMonth   TimeUnit = "MONTH"
	TimeUnitWeek    TimeUnit = "WEEK"
	TimeUnitDay     TimeUnit = "DAY"
	TimeUnitHour    TimeUnit = "HOUR"
	TimeUnitMinute  TimeUnit = "MINUTE"
)

// ReportAggregation is an aggregation of a graph. Code is empty for COUNT.
type ReportAggregation struct {
	Type AggregationType `json:"type"`
	Code string          `json:"code,omitzero"`
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
	By    ReportSortBy `json:"by"`
	Order SortOrder    `json:"order"`
}

// ReportSortBy is the target of a sort of a graph.
type ReportSortBy string

// Report sort targets.
const (
	ReportSortByTotal  ReportSortBy = "TOTAL"
	ReportSortByGroup1 ReportSortBy = "GROUP1"
	ReportSortByGroup2 ReportSortBy = "GROUP2"
	ReportSortByGroup3 ReportSortBy = "GROUP3"
)

// SortOrder is the order of a sort.
type SortOrder string

// Sort orders.
const (
	SortOrderAsc  SortOrder = "ASC"
	SortOrderDesc SortOrder = "DESC"
)

// PeriodicReport is the setting of the periodic report of a graph.
type PeriodicReport struct {
	Active *bool        `json:"active,omitzero"`
	Period ReportPeriod `json:"period,omitzero"`
}

// ReportPeriod is the period of a periodic report.
type ReportPeriod struct {
	Every      TimeUnit       `json:"every"`
	Month      string         `json:"month,omitzero"`
	Time       string         `json:"time,omitzero"`
	Pattern    QuarterPattern `json:"pattern,omitzero"`
	DayOfMonth string         `json:"dayOfMonth,omitzero"`
	DayOfWeek  DayOfWeek      `json:"dayOfWeek,omitzero"`
	Minute     string         `json:"minute,omitzero"`
}

// QuarterPattern is the months of a quarterly periodic report.
type QuarterPattern string

// Quarter patterns.
const (
	QuarterPatternJanAprJulOct QuarterPattern = "JAN_APR_JUL_OCT"
	QuarterPatternFebMayAugNov QuarterPattern = "FEB_MAY_AUG_NOV"
	QuarterPatternMarJunSepDec QuarterPattern = "MAR_JUN_SEP_DEC"
)

// DayOfWeek is a day of the week.
type DayOfWeek string

// Days of the week.
const (
	DayOfWeekSunday    DayOfWeek = "SUNDAY"
	DayOfWeekMonday    DayOfWeek = "MONDAY"
	DayOfWeekTuesday   DayOfWeek = "TUESDAY"
	DayOfWeekWednesday DayOfWeek = "WEDNESDAY"
	DayOfWeekThursday  DayOfWeek = "THURSDAY"
	DayOfWeekFriday    DayOfWeek = "FRIDAY"
	DayOfWeekSaturday  DayOfWeek = "SATURDAY"
)

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
		Revision int64             `json:"revision,omitzero"`
	}{app, reports, revision}
	result, err := c.put[UpdatedReports](ctx, "preview/app/reports.json", params)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
