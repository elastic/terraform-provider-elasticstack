// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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
