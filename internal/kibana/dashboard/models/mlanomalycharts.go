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

package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type MlAnomalyChartsConfigModel struct {
	JobIDs            []types.String                          `tfsdk:"job_ids"`
	MaxSeriesToPlot   types.Int64                             `tfsdk:"max_series_to_plot"`
	SeverityThreshold []MlAnomalyChartsSeverityThresholdModel `tfsdk:"severity_threshold"`
	TimeRange         *TimeRangeModel                         `tfsdk:"time_range"`
	Title             types.String                            `tfsdk:"title"`
	Description       types.String                            `tfsdk:"description"`
	HideTitle         types.Bool                              `tfsdk:"hide_title"`
	HideBorder        types.Bool                              `tfsdk:"hide_border"`
}

type MlAnomalyChartsSeverityThresholdModel struct {
	Severity types.String `tfsdk:"severity"`
	Min      types.Int64  `tfsdk:"min"`
	Max      types.Int64  `tfsdk:"max"`
}
