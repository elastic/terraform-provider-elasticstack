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

package trainedmodelalias

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	elasticsearch "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func readTrainedModelAlias(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state TFModel) (TFModel, bool, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	alias := resourceID
	if alias == "" {
		diags.AddError("Invalid resource ID", "model_alias cannot be empty")
		return state, false, diags
	}

	tflog.Debug(ctx, fmt.Sprintf("Reading ML trained model alias: %s", alias))

	modelID, found, readDiags := elasticsearch.GetMLTrainedModelAlias(ctx, client, alias)
	diags.Append(readDiags...)
	if diags.HasError() {
		return state, false, diags
	}

	// Fallback: if alias resolution fails but we have a model_id from state,
	// verify the underlying model still exists. The GetTrainedModels API may
	// not resolve aliases on all Elasticsearch versions/setups, but querying
	// by the known model_id works.
	if !found && !state.ModelID.IsNull() && !state.ModelID.IsUnknown() {
		modelIDFromState := state.ModelID.ValueString()
		_, modelFound, modelDiags := elasticsearch.GetMLTrainedModelAlias(ctx, client, modelIDFromState)
		diags.Append(modelDiags...)
		if diags.HasError() {
			return state, false, diags
		}
		if modelFound {
			out := state
			out.ModelID = types.StringValue(modelIDFromState)
			tflog.Debug(ctx, fmt.Sprintf("Alias lookup failed but model %s exists; treating alias as found", modelIDFromState))
			return out, true, diags
		}
	}

	if !found {
		return state, false, nil
	}

	out := state
	out.ModelID = types.StringValue(modelID)

	tflog.Debug(ctx, fmt.Sprintf("Successfully read ML trained model alias: %s -> %s", alias, modelID))
	return out, true, diags
}
