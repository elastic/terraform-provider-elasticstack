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

package datafeedstate

import (
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/datafeed"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type MLDatafeedStateData struct {
	ID                      types.String         `tfsdk:"id"`
	ElasticsearchConnection types.List           `tfsdk:"elasticsearch_connection"`
	DatafeedID              types.String         `tfsdk:"datafeed_id"`
	State                   types.String         `tfsdk:"state"`
	Force                   types.Bool           `tfsdk:"force"`
	Timeout                 customtypes.Duration `tfsdk:"datafeed_timeout"`
	Start                   timetypes.RFC3339    `tfsdk:"start"`
	End                     timetypes.RFC3339    `tfsdk:"end"`
	Timeouts                timeouts.Value       `tfsdk:"timeouts"`
}

func timeInSameLocation(ms int64, source timetypes.RFC3339) (time.Time, diag.Diagnostics) {
	t := time.UnixMilli(ms)
	if !typeutils.IsKnown(source) {
		return t, nil
	}

	sourceTime, diags := source.ValueRFC3339Time()
	if diags.HasError() {
		return t, diags
	}

	t = t.In(sourceTime.Location())
	return t, nil
}

func (d *MLDatafeedStateData) SetStartAndEndFromAPI(datafeedStats *models.DatafeedStats) diag.Diagnostics {
	var diags diag.Diagnostics

	if datafeed.State(datafeedStats.State) == datafeed.StateStarted {
		if datafeedStats.RunningState == nil {
			diags.AddWarning(
				"Running state was empty for a started datafeed",
				"The Elasticsearch API returned an empty running state for a Datafeed which was successfully started. Ignoring start and end response values.",
			)
			return diags
		}

		if datafeedStats.RunningState.SearchInterval != nil {
			start, timeDiags := timeInSameLocation(datafeedStats.RunningState.SearchInterval.StartMS, d.Start)
			diags.Append(timeDiags...)
			if diags.HasError() {
				return diags
			}

			end, timeDiags := timeInSameLocation(datafeedStats.RunningState.SearchInterval.EndMS, d.End)
			diags.Append(timeDiags...)
			if diags.HasError() {
				return diags
			}

			d.Start = timetypes.NewRFC3339TimeValue(start)
			d.End = timetypes.NewRFC3339TimeValue(end)
		}

		if datafeedStats.RunningState.RealTimeConfigured {
			d.End = timetypes.NewRFC3339Null()
		}
	}

	if d.Start.IsUnknown() {
		d.Start = timetypes.NewRFC3339Null()
	}

	if d.End.IsUnknown() {
		d.End = timetypes.NewRFC3339Null()
	}

	return diags
}
