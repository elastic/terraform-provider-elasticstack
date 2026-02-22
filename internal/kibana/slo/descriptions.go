package slo

import _ "embed"

//go:embed descriptions/budgeting_method.md
var budgetingMethodDescription string

//go:embed descriptions/time_window.md
var timeWindowDescription string

//go:embed descriptions/objective.md
var objectiveDescription string

//go:embed descriptions/timeslice_metric_aggregation.md
var timesliceMetricAggregationDescription string

//go:embed descriptions/timeslice_metric_field.md
var timesliceMetricFieldDescription string
