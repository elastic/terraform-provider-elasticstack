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

package datafeed

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Datafeed represents the Terraform resource model for ML datafeeds
type Datafeed struct {
	ID                      types.String                                      `tfsdk:"id"`
	ElasticsearchConnection types.List                                        `tfsdk:"elasticsearch_connection"`
	DatafeedID              types.String                                      `tfsdk:"datafeed_id"`
	JobID                   types.String                                      `tfsdk:"job_id"`
	Indices                 types.List                                        `tfsdk:"indices"`
	Query                   jsontypes.Normalized                              `tfsdk:"query"`
	Aggregations            jsontypes.Normalized                              `tfsdk:"aggregations"`
	ScriptFields            customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"script_fields"`
	RuntimeMappings         jsontypes.Normalized                              `tfsdk:"runtime_mappings"`
	ScrollSize              types.Int64                                       `tfsdk:"scroll_size"`
	ChunkingConfig          types.Object                                      `tfsdk:"chunking_config"`
	Frequency               types.String                                      `tfsdk:"frequency"`
	QueryDelay              types.String                                      `tfsdk:"query_delay"`
	DelayedDataCheckConfig  types.Object                                      `tfsdk:"delayed_data_check_config"`
	MaxEmptySearches        types.Int64                                       `tfsdk:"max_empty_searches"`
	IndicesOptions          types.Object                                      `tfsdk:"indices_options"`
}

// ChunkingConfig represents the chunking configuration
type ChunkingConfig struct {
	Mode     types.String `tfsdk:"mode"`
	TimeSpan types.String `tfsdk:"time_span"`
}

// DelayedDataCheckConfig represents the delayed data check configuration
type DelayedDataCheckConfig struct {
	Enabled     types.Bool   `tfsdk:"enabled"`
	CheckWindow types.String `tfsdk:"check_window"`
}

// IndicesOptions represents the indices options for search
type IndicesOptions struct {
	ExpandWildcards   ExpandWildcardsValue `tfsdk:"expand_wildcards"`
	IgnoreUnavailable types.Bool           `tfsdk:"ignore_unavailable"`
	AllowNoIndices    types.Bool           `tfsdk:"allow_no_indices"`
	IgnoreThrottled   types.Bool           `tfsdk:"ignore_throttled"`
}

// toDatafeedRequest converts the Terraform model to a DatafeedRequest.
func (m *Datafeed) toDatafeedRequest(ctx context.Context) (elasticsearch.DatafeedRequest, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	req := elasticsearch.DatafeedRequest{
		JobID:   m.JobID.ValueString(),
		Indices: typeutils.ListTypeToSliceString(ctx, m.Indices, path.Root("indices"), &diags),
	}
	if diags.HasError() {
		return elasticsearch.DatafeedRequest{}, diags
	}

	if typeutils.IsKnown(m.Query) {
		req.Query = json.RawMessage(m.Query.ValueString())
	}

	if typeutils.IsKnown(m.Aggregations) {
		req.Aggregations = json.RawMessage(m.Aggregations.ValueString())
	}

	if typeutils.IsKnown(m.ScriptFields) {
		req.ScriptFields = json.RawMessage(m.ScriptFields.ValueString())
	}

	if typeutils.IsKnown(m.RuntimeMappings) {
		req.RuntimeMappings = json.RawMessage(m.RuntimeMappings.ValueString())
	}

	if typeutils.IsKnown(m.ScrollSize) {
		v := int(m.ScrollSize.ValueInt64())
		req.ScrollSize = &v
	}

	if typeutils.IsKnown(m.Frequency) {
		req.Frequency = m.Frequency.ValueString()
	}

	if typeutils.IsKnown(m.QueryDelay) {
		req.QueryDelay = m.QueryDelay.ValueString()
	}

	if typeutils.IsKnown(m.MaxEmptySearches) {
		v := int(m.MaxEmptySearches.ValueInt64())
		req.MaxEmptySearches = &v
	}

	if typeutils.IsKnown(m.ChunkingConfig) {
		var chunkingConfig ChunkingConfig
		diags.Append(m.ChunkingConfig.As(ctx, &chunkingConfig, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return elasticsearch.DatafeedRequest{}, diags
		}
		cc := &elasticsearch.DatafeedChunkingConfig{
			Mode: chunkingConfig.Mode.ValueString(),
		}
		if chunkingConfig.Mode.ValueString() == "manual" && typeutils.IsKnown(chunkingConfig.TimeSpan) {
			cc.TimeSpan = chunkingConfig.TimeSpan.ValueString()
		}
		req.ChunkingConfig = cc
	}

	if typeutils.IsKnown(m.DelayedDataCheckConfig) {
		var delayedDataCheckConfig DelayedDataCheckConfig
		diags.Append(m.DelayedDataCheckConfig.As(ctx, &delayedDataCheckConfig, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return elasticsearch.DatafeedRequest{}, diags
		}
		ddcc := &elasticsearch.DatafeedDelayedDataCheckConfig{
			Enabled: delayedDataCheckConfig.Enabled.ValueBool(),
		}
		if typeutils.IsKnown(delayedDataCheckConfig.CheckWindow) {
			ddcc.CheckWindow = delayedDataCheckConfig.CheckWindow.ValueString()
		}
		req.DelayedDataCheckConfig = ddcc
	}

	if typeutils.IsKnown(m.IndicesOptions) {
		var indicesOptions IndicesOptions
		diags.Append(m.IndicesOptions.As(ctx, &indicesOptions, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return elasticsearch.DatafeedRequest{}, diags
		}
		ioReq := &elasticsearch.DatafeedIndicesOptions{}
		if typeutils.IsKnown(indicesOptions.ExpandWildcards) {
			var expandWildcards []string
			diags.Append(indicesOptions.ExpandWildcards.ElementsAs(ctx, &expandWildcards, false)...)
			if diags.HasError() {
				return elasticsearch.DatafeedRequest{}, diags
			}
			ioReq.ExpandWildcards = expandWildcards
		}
		if typeutils.IsKnown(indicesOptions.IgnoreUnavailable) {
			v := indicesOptions.IgnoreUnavailable.ValueBool()
			ioReq.IgnoreUnavailable = &v
		}
		if typeutils.IsKnown(indicesOptions.AllowNoIndices) {
			v := indicesOptions.AllowNoIndices.ValueBool()
			ioReq.AllowNoIndices = &v
		}
		if typeutils.IsKnown(indicesOptions.IgnoreThrottled) {
			v := indicesOptions.IgnoreThrottled.ValueBool()
			ioReq.IgnoreThrottled = &v
		}
		req.IndicesOptions = ioReq
	}

	return req, diags
}

// FromAPIModel populates the Terraform model from a typed API response.
func (m *Datafeed) FromAPIModel(ctx context.Context, apiModel *elasticsearch.MLDatafeedResponse) fwdiags.Diagnostics {
	var diags fwdiags.Diagnostics

	m.DatafeedID = types.StringValue(apiModel.DatafeedId)
	m.JobID = types.StringValue(apiModel.JobId)

	// Convert indices (prefer Indices, fall back to Indexes)
	indices := apiModel.Indices
	if len(indices) == 0 {
		indices = apiModel.Indexes
	}
	if len(indices) > 0 {
		indicesList, diag := types.ListValueFrom(ctx, types.StringType, indices)
		diags.Append(diag...)
		m.Indices = indicesList
	} else {
		m.Indices = types.ListNull(types.StringType)
	}

	// Convert query. We use the raw JSON bytes returned by the Elasticsearch API
	// (QueryRaw) rather than re-marshaling apiModel.Query (types.Query), because
	// the typed struct normalises term shorthand to the verbose value form which
	// would cause a permanent diff in state.
	if len(apiModel.QueryRaw) > 0 {
		m.Query = jsontypes.NewNormalizedValue(string(apiModel.QueryRaw))
	} else {
		// Fallback: marshal the typed struct (may normalise term queries).
		queryJSON, err := json.Marshal(apiModel.Query)
		if err != nil {
			diags.AddError("Failed to marshal query", err.Error())
			return diags
		}
		m.Query = jsontypes.NewNormalizedValue(string(queryJSON))
	}

	// Convert aggregations
	if len(apiModel.Aggregations) > 0 {
		aggregationsJSON, err := json.Marshal(apiModel.Aggregations)
		if err != nil {
			diags.AddError("Failed to marshal aggregations", err.Error())
			return diags
		}
		m.Aggregations = jsontypes.NewNormalizedValue(string(aggregationsJSON))
	} else {
		m.Aggregations = jsontypes.NewNormalizedNull()
	}

	// Convert script_fields
	if len(apiModel.ScriptFields) > 0 {
		scriptFieldsJSON, err := json.Marshal(apiModel.ScriptFields)
		if err != nil {
			diags.AddError("Failed to marshal script_fields", err.Error())
			return diags
		}
		m.ScriptFields = customtypes.NewJSONWithDefaultsValue(string(scriptFieldsJSON), populateScriptFieldsDefaults)
	} else {
		m.ScriptFields = customtypes.NewJSONWithDefaultsNull(populateScriptFieldsDefaults)
	}

	// Convert runtime_mappings
	if len(apiModel.RuntimeMappings) > 0 {
		runtimeMappingsJSON, err := json.Marshal(apiModel.RuntimeMappings)
		if err != nil {
			diags.AddError("Failed to marshal runtime_mappings", err.Error())
			return diags
		}
		m.RuntimeMappings = jsontypes.NewNormalizedValue(string(runtimeMappingsJSON))
	} else {
		m.RuntimeMappings = jsontypes.NewNormalizedNull()
	}

	// Convert scroll_size
	if apiModel.ScrollSize != nil {
		m.ScrollSize = types.Int64Value(int64(*apiModel.ScrollSize))
	} else {
		m.ScrollSize = types.Int64Null()
	}

	// Convert frequency (types.Duration is a string alias that marshals to JSON string)
	if apiModel.Frequency != nil {
		freqJSON, err := json.Marshal(apiModel.Frequency)
		if err != nil {
			diags.AddError("Failed to marshal frequency", err.Error())
			return diags
		}
		// Unmarshal the JSON string to get the actual string value
		var freqStr string
		if err := json.Unmarshal(freqJSON, &freqStr); err != nil {
			// freqJSON might be the raw value itself
			freqStr = string(freqJSON)
		}
		m.Frequency = types.StringValue(freqStr)
	} else {
		m.Frequency = types.StringNull()
	}

	// Convert query_delay
	if apiModel.QueryDelay != nil {
		delayJSON, err := json.Marshal(apiModel.QueryDelay)
		if err != nil {
			diags.AddError("Failed to marshal query_delay", err.Error())
			return diags
		}
		var delayStr string
		if err := json.Unmarshal(delayJSON, &delayStr); err != nil {
			delayStr = string(delayJSON)
		}
		m.QueryDelay = types.StringValue(delayStr)
	} else {
		m.QueryDelay = types.StringNull()
	}

	// Convert max_empty_searches
	if apiModel.MaxEmptySearches != nil {
		m.MaxEmptySearches = types.Int64Value(int64(*apiModel.MaxEmptySearches))
	} else {
		m.MaxEmptySearches = types.Int64Null()
	}

	// Convert chunking_config
	if apiModel.ChunkingConfig != nil {
		chunkingConfigTF := ChunkingConfig{
			Mode: types.StringValue(apiModel.ChunkingConfig.Mode.String()),
		}
		// Only set TimeSpan if mode is "manual" and TimeSpan is not nil/empty
		if apiModel.ChunkingConfig.Mode.String() == "manual" && apiModel.ChunkingConfig.TimeSpan != nil {
			tsJSON, err := json.Marshal(apiModel.ChunkingConfig.TimeSpan)
			if err != nil {
				diags.AddError("Failed to marshal chunking_config.time_span", err.Error())
				return diags
			}
			var tsStr string
			if err := json.Unmarshal(tsJSON, &tsStr); err != nil {
				tsStr = string(tsJSON)
			}
			if tsStr != "" {
				chunkingConfigTF.TimeSpan = types.StringValue(tsStr)
			} else {
				chunkingConfigTF.TimeSpan = types.StringNull()
			}
		} else {
			chunkingConfigTF.TimeSpan = types.StringNull()
		}

		chunkingConfigObj, diag := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"mode":      types.StringType,
			"time_span": types.StringType,
		}, chunkingConfigTF)
		diags.Append(diag...)
		m.ChunkingConfig = chunkingConfigObj
	} else {
		m.ChunkingConfig = types.ObjectNull(map[string]attr.Type{
			"mode":      types.StringType,
			"time_span": types.StringType,
		})
	}

	// Convert delayed_data_check_config
	// The typed API DelayedDataCheckConfig has Enabled (bool, not pointer) and CheckWindow (Duration)
	delayedDataCheckConfigTF := DelayedDataCheckConfig{
		Enabled: types.BoolValue(apiModel.DelayedDataCheckConfig.Enabled),
	}
	if apiModel.DelayedDataCheckConfig.CheckWindow != nil {
		cwJSON, err := json.Marshal(apiModel.DelayedDataCheckConfig.CheckWindow)
		if err != nil {
			diags.AddError("Failed to marshal delayed_data_check_config.check_window", err.Error())
			return diags
		}
		var cwStr string
		if err := json.Unmarshal(cwJSON, &cwStr); err != nil {
			cwStr = string(cwJSON)
		}
		delayedDataCheckConfigTF.CheckWindow = types.StringValue(cwStr)
	} else {
		delayedDataCheckConfigTF.CheckWindow = types.StringNull()
	}
	delayedDataCheckConfigObj, diag := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"enabled":      types.BoolType,
		"check_window": types.StringType,
	}, delayedDataCheckConfigTF)
	diags.Append(diag...)
	m.DelayedDataCheckConfig = delayedDataCheckConfigObj

	// Convert indices_options
	if apiModel.IndicesOptions != nil {
		indicesOptionsTF := IndicesOptions{}
		if len(apiModel.IndicesOptions.ExpandWildcards) > 0 {
			elems := make([]attr.Value, len(apiModel.IndicesOptions.ExpandWildcards))
			for i, s := range apiModel.IndicesOptions.ExpandWildcards {
				elems[i] = types.StringValue(s.String())
			}
			expandWildcardsVal, diag := NewExpandWildcardsValue(elems)
			diags.Append(diag...)
			indicesOptionsTF.ExpandWildcards = expandWildcardsVal
		} else {
			indicesOptionsTF.ExpandWildcards = NewExpandWildcardsNull()
		}
		if apiModel.IndicesOptions.IgnoreUnavailable != nil {
			indicesOptionsTF.IgnoreUnavailable = types.BoolValue(*apiModel.IndicesOptions.IgnoreUnavailable)
		} else {
			indicesOptionsTF.IgnoreUnavailable = types.BoolNull()
		}
		if apiModel.IndicesOptions.AllowNoIndices != nil {
			indicesOptionsTF.AllowNoIndices = types.BoolValue(*apiModel.IndicesOptions.AllowNoIndices)
		} else {
			indicesOptionsTF.AllowNoIndices = types.BoolNull()
		}
		if apiModel.IndicesOptions.IgnoreThrottled != nil {
			indicesOptionsTF.IgnoreThrottled = types.BoolValue(*apiModel.IndicesOptions.IgnoreThrottled)
		} else {
			indicesOptionsTF.IgnoreThrottled = types.BoolNull()
		}

		indicesOptionsObj, diag := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"expand_wildcards":   ExpandWildcardsType{SetType: basetypes.SetType{ElemType: types.StringType}},
			"ignore_unavailable": types.BoolType,
			"allow_no_indices":   types.BoolType,
			"ignore_throttled":   types.BoolType,
		}, indicesOptionsTF)
		diags.Append(diag...)
		m.IndicesOptions = indicesOptionsObj
	} else {
		m.IndicesOptions = types.ObjectNull(map[string]attr.Type{
			"expand_wildcards":   ExpandWildcardsType{SetType: basetypes.SetType{ElemType: types.StringType}},
			"ignore_unavailable": types.BoolType,
			"allow_no_indices":   types.BoolType,
			"ignore_throttled":   types.BoolType,
		})
	}

	return diags
}

// toAPICreateModel returns a DatafeedRequest for a create operation.
func (m *Datafeed) toAPICreateModel(ctx context.Context) (elasticsearch.DatafeedRequest, fwdiags.Diagnostics) {
	return m.toDatafeedRequest(ctx)
}

// toAPIUpdateModel returns a DatafeedRequest for an update operation.
// JobID is cleared because update_datafeed does not accept job_id.
func (m *Datafeed) toAPIUpdateModel(ctx context.Context) (elasticsearch.DatafeedRequest, fwdiags.Diagnostics) {
	req, diags := m.toDatafeedRequest(ctx)
	req.JobID = ""
	return req, diags
}
