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
	"fmt"
	"slices"

	"github.com/hashicorp/go-version"
)

var (
	SLOSupportsGroupByMinVersion                = version.Must(version.NewVersion("8.10.0"))
	SLOSupportsMultipleGroupByMinVersion        = version.Must(version.NewVersion("8.14.0"))
	SLOSupportsPreventInitialBackfillMinVersion = version.Must(version.NewVersion("8.15.0"))
	SLOSupportsDataViewIDMinVersion             = version.Must(version.NewVersion("8.15.0"))
	// SLOSettingsSyncFieldMinVersion is the first stack where the Kibana SLO API accepts
	// body.settings.syncField. Older Kibana returns HTTP 400: excess keys (syncField).
	SLOSettingsSyncFieldMinVersion = version.Must(version.NewVersion("8.18.0"))
)

// SLOKqlAccTestConstraints is the supported stack version range for acceptance tests
// that exercise kql_custom_indicator, general settings, and `enabled` (not timeslice-only
// features). Matches other SLO acc coverage: 8.9+ for the SLO API, excluding 8.11.x due to
// known Kibana SLO bugs. Use with versionutils.CheckIfVersionMeetsConstraints.
var SLOKqlAccTestConstraints = mustKqlAccConstraint(">=8.9.0,!=8.11.0,!=8.11.1,!=8.11.2,!=8.11.3,!=8.11.4")

// SLOKqlFleetAccTestConstraints is for KQL SLO acc steps that also use group_by (e.g. Fleet-style
// configs in TestAccResourceSlo_kql_custom_indicator_basic). group_by is rejected below 8.10.0;
// 8.11.x exclusions match SLOKqlAccTestConstraints.
var SLOKqlFleetAccTestConstraints = mustKqlAccConstraint(">=8.10.0,!=8.11.0,!=8.11.1,!=8.11.2,!=8.11.3,!=8.11.4")

func mustKqlAccConstraint(s string) version.Constraints {
	c, err := version.NewConstraint(s)
	if err != nil {
		panic(fmt.Errorf("invalid SLO KQL acc test constraint %q: %w", s, err))
	}
	return c
}

// indicatorAddressToType maps Terraform block names to Kibana API indicator type strings.
var indicatorAddressToType = map[string]string{
	"apm_latency_indicator":      "sli.apm.transactionDuration",
	"apm_availability_indicator": "sli.apm.transactionErrorRate",
	"kql_custom_indicator":       "sli.kql.custom",
	"metric_custom_indicator":    "sli.metric.custom",
	"histogram_custom_indicator": "sli.histogram.custom",
	"timeslice_metric_indicator": "sli.metric.timeslice",
}

// Timeslice metric aggregation types - single source of truth (matches kbapi
// SLOsTimesliceMetric*Metric unions: basic "with field", doc_count, and percentile
// are distinct; there is no value_count timeslice arm).
var (
	timesliceMetricAggregationsBasic = []string{
		"sum", "avg", "min", "max", "last_value", "cardinality", "std_deviation",
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

// Aggregations for which the optional `percentile` attribute must not be set (all except percentile).
var timesliceMetricAggregationsWithoutPercentile = slices.Concat(
	timesliceMetricAggregationsBasic,
	[]string{timesliceMetricAggregationDocCount},
)
