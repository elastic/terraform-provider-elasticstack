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

package lensmetric

import (
	"maps"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Copied from dashboard/descriptions/*.md (Go embed disallows ".." paths outside the package dir).
const (
	metricChartDatasetDescription = "Dataset configuration as JSON. Can be a data view dataset (`type: 'dataview'`), " +
		"index dataset (`type: 'index'`), ES|QL dataset (`type: 'esql'`), or table ES|QL dataset (`type: 'tableESQLDatasetType`)."

	metricChartMetricsDescription = "Array of metrics to display (1-2 items). Each metric can be a primary metric (displays prominently) or secondary metric " +
		"(displays as comparison). Metrics can use field operations (count, unique count, min, max, avg, median, std dev, sum, last value, percentile, percentile ranks), " +
		"pipeline operations (differences, moving average, cumulative sum, counter rate), formula operations, or for ES|QL datasets, column-based value operations."

	metricChartMetricConfigDescription = "Metric configuration as JSON. For primary metrics: includes type ('primary'), operation, format, alignments, icon, " +
		"and optional fields like sub_label, fit, color, apply_color_to, and background_chart. For secondary metrics: includes type ('secondary'), operation, format, " +
		"and optional fields like label, prefix, compare, and color."

	metricChartBreakdownByDescription = "Breakdown configuration as JSON. Groups metrics by a dimension. Can use operations like date histogram, terms, histogram, range, filters, " +
		"or for ES|QL datasets, value operations with columns. Includes optional columns count and collapse_by configuration."
)

func metricChartSchemaAttrs(includePresentation bool) map[string]schema.Attribute {
	attrs := lenscommon.LensChartBaseAttributes()
	attrs["data_source_json"] = lenscommon.DataSourceJSONAttribute(metricChartDatasetDescription)
	attrs["query"] = lenscommon.QueryAttribute(
		"Query configuration for filtering data. Required for non-ES|QL datasets.",
	)
	attrs["metrics"] = lenscommon.JSONConfigItemList(
		metricChartMetricsDescription,
		metricChartMetricConfigDescription,
		lenscommon.PopulateMetricChartMetricDefaults, true,
		listvalidator.SizeAtMost(2),
	)
	attrs["breakdown_by_json"] = schema.StringAttribute{
		MarkdownDescription: metricChartBreakdownByDescription,
		CustomType:          jsontypes.NormalizedType{},
		Optional:            true,
	}
	if includePresentation {
		maps.Copy(attrs, lenscommon.LensChartPresentationAttributes())
	}
	return attrs
}
