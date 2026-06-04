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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/datafeed"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func writeMLDatafeedState(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	req entitycore.WriteRequest[MLDatafeedStateData],
) (entitycore.WriteResult[MLDatafeedStateData], diag.Diagnostics) {
	data := req.Plan
	var diags diag.Diagnostics

	datafeedID := data.DatafeedID.ValueString()
	desiredState := data.State.ValueString()

	datafeedStats, statsDiags := elasticsearch.GetDatafeedStats(ctx, client, datafeedID)
	diags.Append(statsDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[MLDatafeedStateData]{}, diags
	}

	if datafeedStats == nil {
		diags.AddError(
			"ML Datafeed not found",
			fmt.Sprintf("ML datafeed %s does not exist", datafeedID),
		)
		return entitycore.WriteResult[MLDatafeedStateData]{}, diags
	}

	inDesiredState, transitionDiags := performStateTransition(ctx, client, data, datafeed.State(datafeedStats.State.String()))
	diags.Append(transitionDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[MLDatafeedStateData]{}, diags
	}

	compID, idDiags := client.ID(ctx, datafeedID)
	diags.Append(idDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[MLDatafeedStateData]{}, diags
	}

	data.ID = types.StringValue(compID.String())

	var finalData MLDatafeedStateData
	if inDesiredState {
		result, found, readDiags := readMLDatafeedState(ctx, client, datafeedID, data)
		diags.Append(readDiags...)
		if diags.HasError() {
			return entitycore.WriteResult[MLDatafeedStateData]{}, diags
		}
		if !found {
			diags.AddError("Failed to read datafeed stats after update", fmt.Sprintf("The datafeed was successfully transitioned to the %s state, but could not be read after this change", desiredState))
			return entitycore.WriteResult[MLDatafeedStateData]{}, diags
		}
		finalData = result
	} else {
		updated, updateDiags := updateAfterMissedTransition(ctx, client, data, datafeedStats)
		diags.Append(updateDiags...)
		if diags.HasError() {
			return entitycore.WriteResult[MLDatafeedStateData]{}, diags
		}
		if updated == nil {
			diags.AddError("Failed to read datafeed stats after update", fmt.Sprintf("The datafeed was successfully transitioned to the %s state, but could not be read after this change", desiredState))
			return entitycore.WriteResult[MLDatafeedStateData]{}, diags
		}
		finalData = *updated
	}

	return entitycore.WriteResult[MLDatafeedStateData]{Model: finalData}, diags
}
