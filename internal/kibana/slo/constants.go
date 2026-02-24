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

package slo

import (
	"slices"

	"github.com/hashicorp/go-version"
)

var (
	SLOSupportsGroupByMinVersion                = version.Must(version.NewVersion("8.10.0"))
	SLOSupportsMultipleGroupByMinVersion        = version.Must(version.NewVersion("8.14.0"))
	SLOSupportsPreventInitialBackfillMinVersion = version.Must(version.NewVersion("8.15.0"))
	SLOSupportsDataViewIDMinVersion             = version.Must(version.NewVersion("8.15.0"))
)

// indicatorAddressToType maps Terraform block names to Kibana API indicator type strings.
var indicatorAddressToType = map[string]string{
	"apm_latency_indicator":      "sli.apm.transactionDuration",
	"apm_availability_indicator": "sli.apm.transactionErrorRate",
	"kql_custom_indicator":       "sli.kql.custom",
	"metric_custom_indicator":    "sli.metric.custom",
	"histogram_custom_indicator": "sli.histogram.custom",
	"timeslice_metric_indicator": "sli.metric.timeslice",
}

// Timeslice metric aggregation types - single source of truth
var (
	timesliceMetricAggregationsBasic = []string{
		"sum", "avg", "min", "max", "value_count", "last_value",
		"cardinality", "std_deviation",
	}
	timesliceMetricAggregationPercentile = "percentile"
	timesliceMetricAggregationDocCount   = "doc_count"
)

// Derived: all valid aggregations (built from Basic + Percentile + DocCount)
var timesliceMetricAggregations = slices.Concat(
	timesliceMetricAggregationsBasic,
	[]string{timesliceMetricAggregationPercentile, timesliceMetricAggregationDocCount},
)

// Derived: aggregations that require field (Basic + Percentile; doc_count does not)
var timesliceMetricAggregationsWithField = slices.Concat(
	timesliceMetricAggregationsBasic,
	[]string{timesliceMetricAggregationPercentile},
)
