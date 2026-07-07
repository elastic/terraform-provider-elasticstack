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

	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/datafeed"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type MLDatafeedStateData struct {
	entitycore.ElasticsearchConnectionField
	entitycore.ResourceTimeoutsField
	ID                   types.String         `tfsdk:"id"`
	DatafeedID           types.String         `tfsdk:"datafeed_id"`
	State                types.String         `tfsdk:"state"`
	Force                types.Bool           `tfsdk:"force"`
	Timeout              customtypes.Duration `tfsdk:"datafeed_timeout"`
	Start                timetypes.RFC3339    `tfsdk:"start"`
	End                  timetypes.RFC3339    `tfsdk:"end"`
	EffectiveSearchStart timetypes.RFC3339    `tfsdk:"effective_search_start"`
	EffectiveSearchEnd   timetypes.RFC3339    `tfsdk:"effective_search_end"`
}

func (d MLDatafeedStateData) GetID() types.String         { return d.ID }
func (d MLDatafeedStateData) GetResourceID() types.String { return d.DatafeedID }

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

// SetEffectiveSearchIntervalFromAPI populates EffectiveSearchStart /
// EffectiveSearchEnd from running_state.search_interval. Start and End are not
// touched: they round-trip from config (or prior state) so explicitly-supplied
// values are never overwritten by the API response.
func (d *MLDatafeedStateData) SetEffectiveSearchIntervalFromAPI(datafeedStats *estypes.DatafeedStats) diag.Diagnostics {
	var diags diag.Diagnostics

	d.EffectiveSearchStart = timetypes.NewRFC3339Null()
	d.EffectiveSearchEnd = timetypes.NewRFC3339Null()

	if datafeed.State(datafeedStats.State.String()) != datafeed.StateStarted {
		return diags
	}

	if datafeedStats.RunningState == nil {
		diags.AddWarning(
			"Running state was empty for a started datafeed",
			"The Elasticsearch API returned an empty running state for a Datafeed which was successfully started. Ignoring effective search interval response values.",
		)
		return diags
	}

	if datafeedStats.RunningState.SearchInterval != nil {
		startTime, startDiags := timeInSameLocation(datafeedStats.RunningState.SearchInterval.StartMs, d.Start)
		endTime, endDiags := timeInSameLocation(datafeedStats.RunningState.SearchInterval.EndMs, d.End)
		diags.Append(startDiags...)
		diags.Append(endDiags...)
		if diags.HasError() {
			return diags
		}

		d.EffectiveSearchStart = timetypes.NewRFC3339TimeValue(startTime)
		d.EffectiveSearchEnd = timetypes.NewRFC3339TimeValue(endTime)
	}

	if datafeedStats.RunningState.RealTimeConfigured {
		d.EffectiveSearchEnd = timetypes.NewRFC3339Null()
	}

	return diags
}
