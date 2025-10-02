package datafeed

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
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
	ExpandWildcards   types.List `tfsdk:"expand_wildcards"`
	IgnoreUnavailable types.Bool `tfsdk:"ignore_unavailable"`
	AllowNoIndices    types.Bool `tfsdk:"allow_no_indices"`
	IgnoreThrottled   types.Bool `tfsdk:"ignore_throttled"`
}

// ToAPIModel converts the Terraform model to an API model for creating/updating
func (m *Datafeed) ToAPIModel(ctx context.Context) (*models.Datafeed, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	apiModel := &models.Datafeed{
		DatafeedId: m.DatafeedID.ValueString(),
		JobId:      m.JobID.ValueString(),
	}

	// Convert indices
	if !m.Indices.IsNull() && !m.Indices.IsUnknown() {
		var indices []string
		diags.Append(m.Indices.ElementsAs(ctx, &indices, false)...)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.Indices = indices
	}

	// Convert query
	if !m.Query.IsNull() && !m.Query.IsUnknown() {
		var query map[string]interface{}
		diags.Append(m.Query.Unmarshal(&query)...)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.Query = query
	}

	// Convert aggregations
	if !m.Aggregations.IsNull() && !m.Aggregations.IsUnknown() {
		var aggregations map[string]interface{}
		diags.Append(m.Aggregations.Unmarshal(&aggregations)...)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.Aggregations = aggregations
	}

	// Convert script_fields
	if !m.ScriptFields.IsNull() && !m.ScriptFields.IsUnknown() {
		var scriptFields map[string]interface{}
		err := json.Unmarshal([]byte(m.ScriptFields.ValueString()), &scriptFields)
		if err != nil {
			diags.AddError("Failed to unmarshal script_fields", err.Error())
			return nil, diags
		}
		apiModel.ScriptFields = scriptFields
	}

	// Convert runtime_mappings
	if !m.RuntimeMappings.IsNull() && !m.RuntimeMappings.IsUnknown() {
		var runtimeMappings map[string]interface{}
		diags.Append(m.RuntimeMappings.Unmarshal(&runtimeMappings)...)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.RuntimeMappings = runtimeMappings
	}

	// Convert scroll_size
	if !m.ScrollSize.IsNull() && !m.ScrollSize.IsUnknown() {
		scrollSize := int(m.ScrollSize.ValueInt64())
		apiModel.ScrollSize = &scrollSize
	}

	// Convert frequency
	if !m.Frequency.IsNull() && !m.Frequency.IsUnknown() {
		frequency := m.Frequency.ValueString()
		apiModel.Frequency = &frequency
	}

	// Convert query_delay
	if !m.QueryDelay.IsNull() && !m.QueryDelay.IsUnknown() {
		queryDelay := m.QueryDelay.ValueString()
		apiModel.QueryDelay = &queryDelay
	}

	// Convert max_empty_searches
	if !m.MaxEmptySearches.IsNull() && !m.MaxEmptySearches.IsUnknown() {
		maxEmptySearches := int(m.MaxEmptySearches.ValueInt64())
		apiModel.MaxEmptySearches = &maxEmptySearches
	}

	// Convert chunking_config
	if !m.ChunkingConfig.IsNull() && !m.ChunkingConfig.IsUnknown() {
		var chunkingConfig ChunkingConfig
		diags.Append(m.ChunkingConfig.As(ctx, &chunkingConfig, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}

		apiChunkingConfig := &models.ChunkingConfig{
			Mode: chunkingConfig.Mode.ValueString(),
		}
		// Only set TimeSpan if mode is "manual" and TimeSpan is provided
		if chunkingConfig.Mode.ValueString() == "manual" && !chunkingConfig.TimeSpan.IsNull() && !chunkingConfig.TimeSpan.IsUnknown() {
			apiChunkingConfig.TimeSpan = chunkingConfig.TimeSpan.ValueString()
		}
		apiModel.ChunkingConfig = apiChunkingConfig
	}

	// Convert delayed_data_check_config
	if !m.DelayedDataCheckConfig.IsNull() && !m.DelayedDataCheckConfig.IsUnknown() {
		var delayedDataCheckConfig DelayedDataCheckConfig
		diags.Append(m.DelayedDataCheckConfig.As(ctx, &delayedDataCheckConfig, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}

		apiDelayedDataCheckConfig := &models.DelayedDataCheckConfig{}
		if !delayedDataCheckConfig.Enabled.IsNull() && !delayedDataCheckConfig.Enabled.IsUnknown() {
			enabled := delayedDataCheckConfig.Enabled.ValueBool()
			apiDelayedDataCheckConfig.Enabled = &enabled
		}
		if !delayedDataCheckConfig.CheckWindow.IsNull() && !delayedDataCheckConfig.CheckWindow.IsUnknown() {
			checkWindow := delayedDataCheckConfig.CheckWindow.ValueString()
			apiDelayedDataCheckConfig.CheckWindow = &checkWindow
		}
		apiModel.DelayedDataCheckConfig = apiDelayedDataCheckConfig
	}

	// Convert indices_options
	if !m.IndicesOptions.IsNull() && !m.IndicesOptions.IsUnknown() {
		var indicesOptions IndicesOptions
		diags.Append(m.IndicesOptions.As(ctx, &indicesOptions, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}

		apiIndicesOptions := &models.IndicesOptions{}
		if !indicesOptions.ExpandWildcards.IsNull() && !indicesOptions.ExpandWildcards.IsUnknown() {
			var expandWildcards []string
			diags.Append(indicesOptions.ExpandWildcards.ElementsAs(ctx, &expandWildcards, false)...)
			if diags.HasError() {
				return nil, diags
			}
			apiIndicesOptions.ExpandWildcards = expandWildcards
		}
		if !indicesOptions.IgnoreUnavailable.IsNull() && !indicesOptions.IgnoreUnavailable.IsUnknown() {
			ignoreUnavailable := indicesOptions.IgnoreUnavailable.ValueBool()
			apiIndicesOptions.IgnoreUnavailable = &ignoreUnavailable
		}
		if !indicesOptions.AllowNoIndices.IsNull() && !indicesOptions.AllowNoIndices.IsUnknown() {
			allowNoIndices := indicesOptions.AllowNoIndices.ValueBool()
			apiIndicesOptions.AllowNoIndices = &allowNoIndices
		}
		if !indicesOptions.IgnoreThrottled.IsNull() && !indicesOptions.IgnoreThrottled.IsUnknown() {
			ignoreThrottled := indicesOptions.IgnoreThrottled.ValueBool()
			apiIndicesOptions.IgnoreThrottled = &ignoreThrottled
		}
		apiModel.IndicesOptions = apiIndicesOptions
	}

	return apiModel, diags
}

// FromAPIModel populates the Terraform model from an API model
func (m *Datafeed) FromAPIModel(ctx context.Context, apiModel *models.Datafeed) fwdiags.Diagnostics {
	var diags fwdiags.Diagnostics

	m.DatafeedID = types.StringValue(apiModel.DatafeedId)
	m.JobID = types.StringValue(apiModel.JobId)

	// Convert indices
	if len(apiModel.Indices) > 0 {
		indicesList, diag := types.ListValueFrom(ctx, types.StringType, apiModel.Indices)
		diags.Append(diag...)
		m.Indices = indicesList
	} else {
		m.Indices = types.ListNull(types.StringType)
	}

	// Convert query
	if apiModel.Query != nil {
		queryJSON, err := json.Marshal(apiModel.Query)
		if err != nil {
			diags.AddError("Failed to marshal query", err.Error())
			return diags
		}
		m.Query = jsontypes.NewNormalizedValue(string(queryJSON))
	} else {
		m.Query = jsontypes.NewNormalizedNull()
	}

	// Convert aggregations
	if apiModel.Aggregations != nil {
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
	if apiModel.ScriptFields != nil {
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
	if apiModel.RuntimeMappings != nil {
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

	// Convert frequency
	if apiModel.Frequency != nil {
		m.Frequency = types.StringValue(*apiModel.Frequency)
	} else {
		m.Frequency = types.StringNull()
	}

	// Convert query_delay
	if apiModel.QueryDelay != nil {
		m.QueryDelay = types.StringValue(*apiModel.QueryDelay)
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
			Mode: types.StringValue(apiModel.ChunkingConfig.Mode),
		}
		// Only set TimeSpan if mode is "manual" and TimeSpan is not empty
		if apiModel.ChunkingConfig.Mode == "manual" && apiModel.ChunkingConfig.TimeSpan != "" {
			chunkingConfigTF.TimeSpan = types.StringValue(apiModel.ChunkingConfig.TimeSpan)
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
	if apiModel.DelayedDataCheckConfig != nil {
		delayedDataCheckConfigTF := DelayedDataCheckConfig{
			Enabled:     types.BoolPointerValue(apiModel.DelayedDataCheckConfig.Enabled),
			CheckWindow: types.StringPointerValue(apiModel.DelayedDataCheckConfig.CheckWindow),
		}

		delayedDataCheckConfigObj, diag := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"enabled":      types.BoolType,
			"check_window": types.StringType,
		}, delayedDataCheckConfigTF)
		diags.Append(diag...)
		m.DelayedDataCheckConfig = delayedDataCheckConfigObj
	} else {
		m.DelayedDataCheckConfig = types.ObjectNull(map[string]attr.Type{
			"enabled":      types.BoolType,
			"check_window": types.StringType,
		})
	}

	// Convert indices_options
	if apiModel.IndicesOptions != nil {
		indicesOptionsTF := IndicesOptions{}
		if len(apiModel.IndicesOptions.ExpandWildcards) > 0 {
			expandWildcardsList, diag := types.ListValueFrom(ctx, types.StringType, apiModel.IndicesOptions.ExpandWildcards)
			diags.Append(diag...)
			indicesOptionsTF.ExpandWildcards = expandWildcardsList
		} else {
			indicesOptionsTF.ExpandWildcards = types.ListNull(types.StringType)
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
			"expand_wildcards":   types.ListType{ElemType: types.StringType},
			"ignore_unavailable": types.BoolType,
			"allow_no_indices":   types.BoolType,
			"ignore_throttled":   types.BoolType,
		}, indicesOptionsTF)
		diags.Append(diag...)
		m.IndicesOptions = indicesOptionsObj
	} else {
		m.IndicesOptions = types.ObjectNull(map[string]attr.Type{
			"expand_wildcards":   types.ListType{ElemType: types.StringType},
			"ignore_unavailable": types.BoolType,
			"allow_no_indices":   types.BoolType,
			"ignore_throttled":   types.BoolType,
		})
	}

	return diags
}

// toAPICreateModel converts the Terraform model to a DatafeedCreateRequest
func (m *Datafeed) toAPICreateModel(ctx context.Context) (*models.DatafeedCreateRequest, fwdiags.Diagnostics) {
	apiModel, diags := m.ToAPIModel(ctx)
	if diags.HasError() {
		return nil, diags
	}

	createRequest := &models.DatafeedCreateRequest{
		JobId:                  apiModel.JobId,
		Indices:                apiModel.Indices,
		Query:                  apiModel.Query,
		Aggregations:           apiModel.Aggregations,
		ScriptFields:           apiModel.ScriptFields,
		RuntimeMappings:        apiModel.RuntimeMappings,
		ScrollSize:             apiModel.ScrollSize,
		ChunkingConfig:         apiModel.ChunkingConfig,
		Frequency:              apiModel.Frequency,
		QueryDelay:             apiModel.QueryDelay,
		DelayedDataCheckConfig: apiModel.DelayedDataCheckConfig,
		MaxEmptySearches:       apiModel.MaxEmptySearches,
		IndicesOptions:         apiModel.IndicesOptions,
	}

	return createRequest, diags
}

// toAPIUpdateModel converts the Terraform model to a DatafeedUpdateRequest
func (m *Datafeed) toAPIUpdateModel(ctx context.Context) (*models.DatafeedUpdateRequest, fwdiags.Diagnostics) {
	apiModel, diags := m.ToAPIModel(ctx)
	if diags.HasError() {
		return nil, diags
	}

	// Create the datafeed update request (note: job_id cannot be updated)
	updateRequest := &models.DatafeedUpdateRequest{
		Indices:                apiModel.Indices,
		Query:                  apiModel.Query,
		Aggregations:           apiModel.Aggregations,
		ScriptFields:           apiModel.ScriptFields,
		RuntimeMappings:        apiModel.RuntimeMappings,
		ScrollSize:             apiModel.ScrollSize,
		ChunkingConfig:         apiModel.ChunkingConfig,
		Frequency:              apiModel.Frequency,
		QueryDelay:             apiModel.QueryDelay,
		DelayedDataCheckConfig: apiModel.DelayedDataCheckConfig,
		MaxEmptySearches:       apiModel.MaxEmptySearches,
		IndicesOptions:         apiModel.IndicesOptions,
	}

	return updateRequest, diags
}
