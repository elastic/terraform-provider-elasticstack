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

package osquerypack

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// postReadOsqueryPack preserves write-only fields that the Kibana API does not
// return in GET responses. Currently this is limited to queries.*.saved_query_id,
// which is accepted on create/update but omitted from read responses.
func postReadOsqueryPack(
	ctx context.Context,
	req entitycore.KibanaPostReadRequest[osqueryPackModel],
) (osqueryPackModel, diag.Diagnostics) {
	state := req.State
	prior := req.Prior

	merged, diags := mergeSavedQueryIDs(ctx, state.Queries, prior.Queries)
	if diags.HasError() {
		return state, diags
	}
	state.Queries = merged
	return state, nil
}

// mergeSavedQueryIDs copies saved_query_id values from priorQueries into
// stateQueries for each query name where the API returned a null value.
func mergeSavedQueryIDs(ctx context.Context, stateQueries, priorQueries types.Map) (types.Map, diag.Diagnostics) {
	var diags diag.Diagnostics

	if stateQueries.IsNull() || stateQueries.IsUnknown() {
		return stateQueries, nil
	}
	if priorQueries.IsNull() || priorQueries.IsUnknown() {
		return stateQueries, nil
	}

	var priorMap map[string]queryModel
	diags.Append(priorQueries.ElementsAs(ctx, &priorMap, false)...)
	if diags.HasError() {
		return stateQueries, diags
	}

	var stateMap map[string]basetypes.ObjectValue
	diags.Append(stateQueries.ElementsAs(ctx, &stateMap, false)...)
	if diags.HasError() {
		return stateQueries, diags
	}

	changed := false
	for name, stateObj := range stateMap {
		prior, hasPrior := priorMap[name]
		if !hasPrior || prior.SavedQueryID.IsNull() || prior.SavedQueryID.IsUnknown() {
			continue
		}

		var stateQuery queryModel
		d := stateObj.As(ctx, &stateQuery, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return stateQueries, diags
		}

		if !stateQuery.SavedQueryID.IsNull() {
			continue
		}

		stateQuery.SavedQueryID = prior.SavedQueryID
		merged, d := types.ObjectValueFrom(ctx, queryAttrTypes(), stateQuery)
		diags.Append(d...)
		if diags.HasError() {
			return stateQueries, diags
		}

		stateMap[name] = merged
		changed = true
	}

	if !changed {
		return stateQueries, diags
	}

	result, d := types.MapValueFrom(ctx, queryMapElemType(), stateMap)
	diags.Append(d...)
	return result, diags
}
