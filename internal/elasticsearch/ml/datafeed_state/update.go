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
	"context"
	"fmt"

	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/datafeed"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func updateAfterMissedTransition(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	data MLDatafeedStateData,
	datafeedStats *estypes.DatafeedStats,
) (*MLDatafeedStateData, diag.Diagnostics) {
	datafeedID := data.DatafeedID.ValueString()
	var diags diag.Diagnostics

	statsAfterUpdate, getDiags := elasticsearch.GetDatafeedStats(ctx, client, datafeedID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return nil, diags
	}

	if statsAfterUpdate == nil {
		diags.AddError(
			"ML Datafeed not found",
			fmt.Sprintf("ML datafeed %s does not exist after successful update", datafeedID),
		)
		return nil, diags
	}

	if statsAfterUpdate.TimingStats == nil || datafeedStats.TimingStats == nil {
		diags.AddWarning("Expected Datafeed to contain timing stats",
			fmt.Sprintf("Stats for datafeed %s did not contain timing stats either before or after the update. Before %v - After %v", datafeedID, datafeedStats, statsAfterUpdate))
	} else if statsAfterUpdate.TimingStats.SearchCount <= datafeedStats.TimingStats.SearchCount {
		diags.AddError(
			"Datafeed did not successfully transition to the desired state",
			fmt.Sprintf("[%s] datafeed did not settle into the [%s] state. The current state is [%s]", datafeedID, data.State.ValueString(), statsAfterUpdate.State.String()),
		)
		return nil, diags
	}

	data.EffectiveSearchStart = timetypes.NewRFC3339Null()
	data.EffectiveSearchEnd = timetypes.NewRFC3339Null()

	return &data, diags
}

func performStateTransition(ctx context.Context, client *clients.ElasticsearchScopedClient, data MLDatafeedStateData, currentState datafeed.State) (bool, diag.Diagnostics) {
	datafeedID := data.DatafeedID.ValueString()
	desiredState := datafeed.State(data.State.ValueString())
	force := data.Force.ValueBool()

	timeout, parseErrs := data.Timeout.Parse()
	if parseErrs.HasError() {
		return false, parseErrs
	}

	if currentState == desiredState {
		tflog.Debug(ctx, fmt.Sprintf("ML datafeed %s is already in desired state %s", datafeedID, desiredState))
		return true, nil
	}

	switch desiredState {
	case datafeed.StateStarted:
		start := data.Start.ValueString()
		end := data.End.ValueString()

		diags := elasticsearch.StartDatafeed(ctx, client, datafeedID, start, end, timeout)
		if diags.HasError() {
			return false, diags
		}
	case datafeed.StateStopped:
		if diags := elasticsearch.StopDatafeed(ctx, client, datafeedID, force, timeout); diags.HasError() {
			return false, diags
		}
	default:
		return false, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Invalid state",
				fmt.Sprintf("Invalid state %s. Valid states are 'started' and 'stopped'", desiredState),
			),
		}
	}

	inDesiredState, diags := datafeed.WaitForDatafeedState(ctx, client, datafeedID, desiredState)
	if diags.HasError() {
		return false, diags
	}

	tflog.Info(ctx, fmt.Sprintf("ML datafeed %s successfully transitioned to state %s", datafeedID, desiredState))
	return inDesiredState, nil
}
