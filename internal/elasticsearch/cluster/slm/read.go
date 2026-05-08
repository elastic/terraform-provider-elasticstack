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

package slm

import (
	"context"
	"encoding/json"
	"fmt"

	esclients "github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func readSlm(ctx context.Context, client *esclients.ElasticsearchScopedClient, resourceID string, state Data) (Data, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	slm, sdkDiags := elasticsearch.GetSlm(ctx, client, resourceID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return state, false, diags
	}

	if slm == nil {
		tflog.Warn(ctx, fmt.Sprintf(`SLM policy "%s" not found, removing from state`, resourceID))
		return state, false, diags
	}

	data, diags := mapSlmToData(ctx, slm, resourceID, state)
	if diags.HasError() {
		return state, false, diags
	}
	return data, true, diags
}

func mapSlmToData(ctx context.Context, slm *elasticsearch.SlmPolicy, resourceID string, state Data) (Data, diag.Diagnostics) {
	var diags diag.Diagnostics
	var data Data

	data.Name = types.StringValue(resourceID)
	data.Repository = types.StringValue(slm.Repository)
	data.Schedule = types.StringValue(slm.Schedule)
	data.SnapshotName = types.StringValue(slm.Name)

	// Retention
	if slm.Retention != nil {
		if slm.Retention.ExpireAfter != nil {
			data.ExpireAfter = types.StringValue(*slm.Retention.ExpireAfter)
		} else {
			data.ExpireAfter = types.StringNull()
		}
		if slm.Retention.MaxCount != nil {
			data.MaxCount = types.Int64Value(int64(*slm.Retention.MaxCount))
		} else {
			data.MaxCount = types.Int64Null()
		}
		if slm.Retention.MinCount != nil {
			data.MinCount = types.Int64Value(int64(*slm.Retention.MinCount))
		} else {
			data.MinCount = types.Int64Null()
		}
	} else {
		data.ExpireAfter = types.StringNull()
		data.MaxCount = types.Int64Null()
		data.MinCount = types.Int64Null()
	}

	// Config
	if c := slm.Config; c != nil {
		if c.ExpandWildcards != "" {
			data.ExpandWildcards = types.StringValue(c.ExpandWildcards)
		} else {
			data.ExpandWildcards = types.StringValue(defaultExpandWildcards)
		}

		if c.IncludeGlobalState != nil {
			data.IncludeGlobalState = types.BoolValue(*c.IncludeGlobalState)
		} else {
			data.IncludeGlobalState = types.BoolValue(true)
		}
		if c.IgnoreUnavailable != nil {
			data.IgnoreUnavailable = types.BoolValue(*c.IgnoreUnavailable)
		} else {
			data.IgnoreUnavailable = types.BoolValue(false)
		}
		if c.Partial != nil {
			data.Partial = types.BoolValue(*c.Partial)
		} else {
			data.Partial = types.BoolValue(false)
		}

		// Indices: when the API omits indices, derive the value from the passed state to avoid plan diffs.
		switch {
		case len(c.Indices) > 0:
			indicesList, listDiags := types.ListValueFrom(ctx, types.StringType, c.Indices)
			diags.Append(listDiags...)
			if diags.HasError() {
				return state, diags
			}
			data.Indices = indicesList
		case state.Indices.IsNull() || state.Indices.IsUnknown():
			data.Indices = types.ListNull(types.StringType)
		case len(state.Indices.Elements()) == 0:
			data.Indices, _ = types.ListValueFrom(ctx, types.StringType, []string{})
		default:
			data.Indices, _ = types.ListValueFrom(ctx, types.StringType, []string{})
		}

		// FeatureStates: when the API omits feature states, derive the value from the passed state to avoid plan diffs.
		switch {
		case len(c.FeatureStates) > 0:
			featureStatesSet, setDiags := types.SetValueFrom(ctx, types.StringType, c.FeatureStates)
			diags.Append(setDiags...)
			if diags.HasError() {
				return state, diags
			}
			data.FeatureStates = featureStatesSet
		case state.FeatureStates.IsNull() || state.FeatureStates.IsUnknown():
			data.FeatureStates = types.SetNull(types.StringType)
		case len(state.FeatureStates.Elements()) == 0:
			data.FeatureStates, _ = types.SetValueFrom(ctx, types.StringType, []string{})
		default:
			data.FeatureStates, _ = types.SetValueFrom(ctx, types.StringType, []string{})
		}

		// Metadata
		if c.Metadata != nil {
			meta := make(map[string]any)
			for k, v := range c.Metadata {
				var val any
				if err := json.Unmarshal(v, &val); err != nil {
					diags.AddError("Failed to unmarshal metadata", fmt.Sprintf("failed to unmarshal metadata key %q: %s", k, err))
					return state, diags
				}
				meta[k] = val
			}
			metaBytes, err := json.Marshal(meta)
			if err != nil {
				diags.AddError("Failed to marshal metadata", err.Error())
				return state, diags
			}
			data.Metadata = jsontypes.NewNormalizedValue(string(metaBytes))
		} else {
			data.Metadata = jsontypes.NewNormalizedNull()
		}
	} else {
		data.ExpandWildcards = types.StringValue(defaultExpandWildcards)
		data.IncludeGlobalState = types.BoolValue(true)
		data.IgnoreUnavailable = types.BoolValue(false)
		data.Partial = types.BoolValue(false)
		data.Indices = types.ListNull(types.StringType)
		data.FeatureStates = types.SetNull(types.StringType)
		data.Metadata = jsontypes.NewNormalizedNull()
	}

	// Preserve envelope-managed fields from state.
	data.ID = state.ID
	data.ElasticsearchConnection = state.ElasticsearchConnection

	return data, diags
}
