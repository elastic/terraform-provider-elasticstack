package dashboard

import _ "embed"

//go:embed descriptions/xy_layer_type.md
var xyLayerTypeDescription string

//go:embed descriptions/reference_line_icon.md
var referenceLineIconDescription string

//go:embed descriptions/tagcloud_metric.md
var tagcloudMetricDescription string

//go:embed descriptions/heatmap_x_axis.md
var heatmapXAxisDescription string

//go:embed descriptions/heatmap_y_axis.md
var heatmapYAxisDescription string

//go:embed descriptions/region_map_region.md
var regionMapRegionDescription string

//go:embed descriptions/gauge_metric.md
var gaugeMetricDescription string

//go:embed descriptions/metric_chart_dataset.md
var metricChartDatasetDescription string

//go:embed descriptions/metric_chart_metrics.md
var metricChartMetricsDescription string

//go:embed descriptions/metric_chart_metric_config.md
var metricChartMetricConfigDescription string

//go:embed descriptions/metric_chart_breakdown_by.md
var metricChartBreakdownByDescription string
