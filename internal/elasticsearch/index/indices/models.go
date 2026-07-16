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

package indices

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	indexparent "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/aliasutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	staticSettingsKeys = []string{
		indexparent.SettingNumberOfShards,
		indexparent.SettingNumberOfRoutingShards,
		indexparent.SettingCodec,
		indexparent.SettingRoutingPartitionSize,
		indexparent.SettingLoadFixedBitsetFiltersEagerly,
		indexparent.SettingShardCheckOnStartup,
		indexparent.SettingSortField,
		indexparent.SettingSortOrder,
		indexparent.SettingMappingCoerce,
	}
	dynamicSettingsKeys = []string{
		indexparent.SettingNumberOfReplicas,
		indexparent.SettingAutoExpandReplicas,
		indexparent.SettingRefreshInterval,
		indexparent.SettingSearchIdleAfter,
		indexparent.SettingMaxResultWindow,
		indexparent.SettingMaxInnerResultWindow,
		indexparent.SettingMaxRescoreWindow,
		indexparent.SettingMaxDocvalueFieldsSearch,
		indexparent.SettingMaxScriptFields,
		indexparent.SettingMaxNgramDiff,
		indexparent.SettingMaxShingleDiff,
		"blocks.read_only",
		"blocks.read_only_allow_delete",
		"blocks.read",
		"blocks.write",
		"blocks.metadata",
		indexparent.SettingMaxRefreshListeners,
		"analyze.max_token_count",
		"highlight.max_analyzed_offset",
		indexparent.SettingMaxTermsCount,
		indexparent.SettingMaxRegexLength,
		indexparent.SettingQueryDefaultField,
		indexparent.SettingRoutingAllocationEnable,
		indexparent.SettingRoutingRebalanceEnable,
		indexparent.SettingGCDeletes,
		indexparent.SettingDefaultPipeline,
		indexparent.SettingFinalPipeline,
		indexparent.SettingUnassignedNodeLeftDelayedTimeout,
		"search.slowlog.threshold.query.warn",
		"search.slowlog.threshold.query.info",
		"search.slowlog.threshold.query.debug",
		"search.slowlog.threshold.query.trace",
		"search.slowlog.threshold.fetch.warn",
		"search.slowlog.threshold.fetch.info",
		"search.slowlog.threshold.fetch.debug",
		"search.slowlog.threshold.fetch.trace",
		"search.slowlog.level",
		"indexing.slowlog.threshold.index.warn",
		"indexing.slowlog.threshold.index.info",
		"indexing.slowlog.threshold.index.debug",
		"indexing.slowlog.threshold.index.trace",
		"indexing.slowlog.level",
		"indexing.slowlog.source",
	}
	allSettingsKeys = func() []string {
		all := make([]string, 0, len(staticSettingsKeys)+len(dynamicSettingsKeys))
		all = append(all, staticSettingsKeys...)
		all = append(all, dynamicSettingsKeys...)
		return all
	}()
)

type tfModel struct {
	entitycore.ElasticsearchConnectionField
	ID      types.String `tfsdk:"id"`
	Target  types.String `tfsdk:"target"`
	Indices types.List   `tfsdk:"indices"`
}

type indexTfModel struct {
	ID                                 types.String         `tfsdk:"id"`
	Name                               types.String         `tfsdk:"name"`
	NumberOfShards                     types.Int64          `tfsdk:"number_of_shards"`
	NumberOfRoutingShards              types.Int64          `tfsdk:"number_of_routing_shards"`
	Codec                              types.String         `tfsdk:"codec"`
	RoutingPartitionSize               types.Int64          `tfsdk:"routing_partition_size"`
	LoadFixedBitsetFiltersEagerly      types.Bool           `tfsdk:"load_fixed_bitset_filters_eagerly"`
	ShardCheckOnStartup                types.String         `tfsdk:"shard_check_on_startup"`
	SortField                          types.Set            `tfsdk:"sort_field"`
	SortOrder                          types.List           `tfsdk:"sort_order"`
	MappingCoerce                      types.Bool           `tfsdk:"mapping_coerce"`
	NumberOfReplicas                   types.Int64          `tfsdk:"number_of_replicas"`
	AutoExpandReplicas                 types.String         `tfsdk:"auto_expand_replicas"`
	SearchIdleAfter                    types.String         `tfsdk:"search_idle_after"`
	RefreshInterval                    types.String         `tfsdk:"refresh_interval"`
	MaxResultWindow                    types.Int64          `tfsdk:"max_result_window"`
	MaxInnerResultWindow               types.Int64          `tfsdk:"max_inner_result_window"`
	MaxRescoreWindow                   types.Int64          `tfsdk:"max_rescore_window"`
	MaxDocvalueFieldsSearch            types.Int64          `tfsdk:"max_docvalue_fields_search"`
	MaxScriptFields                    types.Int64          `tfsdk:"max_script_fields"`
	MaxNGramDiff                       types.Int64          `tfsdk:"max_ngram_diff"`
	MaxShingleDiff                     types.Int64          `tfsdk:"max_shingle_diff"`
	MaxRefreshListeners                types.Int64          `tfsdk:"max_refresh_listeners"`
	AnalyzeMaxTokenCount               types.Int64          `tfsdk:"analyze_max_token_count"`
	HighlightMaxAnalyzedOffset         types.Int64          `tfsdk:"highlight_max_analyzed_offset"`
	MaxTermsCount                      types.Int64          `tfsdk:"max_terms_count"`
	MaxRegexLength                     types.Int64          `tfsdk:"max_regex_length"`
	QueryDefaultField                  types.Set            `tfsdk:"query_default_field"`
	RoutingAllocationEnable            types.String         `tfsdk:"routing_allocation_enable"`
	RoutingRebalanceEnable             types.String         `tfsdk:"routing_rebalance_enable"`
	GCDeletes                          types.String         `tfsdk:"gc_deletes"`
	BlocksReadOnly                     types.Bool           `tfsdk:"blocks_read_only"`
	BlocksReadOnlyAllowDelete          types.Bool           `tfsdk:"blocks_read_only_allow_delete"`
	BlocksRead                         types.Bool           `tfsdk:"blocks_read"`
	BlocksWrite                        types.Bool           `tfsdk:"blocks_write"`
	BlocksMetadata                     types.Bool           `tfsdk:"blocks_metadata"`
	DefaultPipeline                    types.String         `tfsdk:"default_pipeline"`
	FinalPipeline                      types.String         `tfsdk:"final_pipeline"`
	UnassignedNodeLeftDelayedTimeout   types.String         `tfsdk:"unassigned_node_left_delayed_timeout"`
	SearchSlowlogThresholdQueryWarn    types.String         `tfsdk:"search_slowlog_threshold_query_warn"`
	SearchSlowlogThresholdQueryInfo    types.String         `tfsdk:"search_slowlog_threshold_query_info"`
	SearchSlowlogThresholdQueryDebug   types.String         `tfsdk:"search_slowlog_threshold_query_debug"`
	SearchSlowlogThresholdQueryTrace   types.String         `tfsdk:"search_slowlog_threshold_query_trace"`
	SearchSlowlogThresholdFetchWarn    types.String         `tfsdk:"search_slowlog_threshold_fetch_warn"`
	SearchSlowlogThresholdFetchInfo    types.String         `tfsdk:"search_slowlog_threshold_fetch_info"`
	SearchSlowlogThresholdFetchDebug   types.String         `tfsdk:"search_slowlog_threshold_fetch_debug"`
	SearchSlowlogThresholdFetchTrace   types.String         `tfsdk:"search_slowlog_threshold_fetch_trace"`
	SearchSlowlogLevel                 types.String         `tfsdk:"search_slowlog_level"`
	IndexingSlowlogThresholdIndexWarn  types.String         `tfsdk:"indexing_slowlog_threshold_index_warn"`
	IndexingSlowlogThresholdIndexInfo  types.String         `tfsdk:"indexing_slowlog_threshold_index_info"`
	IndexingSlowlogThresholdIndexDebug types.String         `tfsdk:"indexing_slowlog_threshold_index_debug"`
	IndexingSlowlogThresholdIndexTrace types.String         `tfsdk:"indexing_slowlog_threshold_index_trace"`
	IndexingSlowlogLevel               types.String         `tfsdk:"indexing_slowlog_level"`
	IndexingSlowlogSource              types.String         `tfsdk:"indexing_slowlog_source"`
	AnalysisAnalyzer                   jsontypes.Normalized `tfsdk:"analysis_analyzer"`
	AnalysisTokenizer                  jsontypes.Normalized `tfsdk:"analysis_tokenizer"`
	AnalysisCharFilter                 jsontypes.Normalized `tfsdk:"analysis_char_filter"`
	AnalysisFilter                     jsontypes.Normalized `tfsdk:"analysis_filter"`
	AnalysisNormalizer                 jsontypes.Normalized `tfsdk:"analysis_normalizer"`
	DeletionProtection                 types.Bool           `tfsdk:"deletion_protection"`
	WaitForActiveShards                types.String         `tfsdk:"wait_for_active_shards"`
	MasterTimeout                      customtypes.Duration `tfsdk:"master_timeout"`
	Timeout                            customtypes.Duration `tfsdk:"timeout"`
	Mappings                           jsontypes.Normalized `tfsdk:"mappings"`
	SettingsRaw                        jsontypes.Normalized `tfsdk:"settings_raw"`
	Alias                              types.Set            `tfsdk:"alias"`
}

func (model *indexTfModel) populateFromAPI(ctx context.Context, indexName string, apiModel estypes.IndexState) diag.Diagnostics {
	model.Name = types.StringValue(indexName)
	model.SortField = types.SetValueMust(types.StringType, []attr.Value{})
	model.SortOrder = types.ListValueMust(types.StringType, []attr.Value{})
	model.QueryDefaultField = types.SetValueMust(types.StringType, []attr.Value{})

	modelMappings, diags := mappingsFromAPI(apiModel)
	if diags.HasError() {
		return diags
	}
	modelAliases, diags := aliasutil.AliasesFromAPI(ctx, apiModel.Aliases, aliasElementType())
	if diags.HasError() {
		return diags
	}

	model.Mappings = modelMappings
	model.Alias = modelAliases

	diags = setSettingsFromAPI(ctx, model, apiModel)
	if diags.HasError() {
		return diags
	}

	return nil
}

func mappingsFromAPI(apiModel estypes.IndexState) (jsontypes.Normalized, diag.Diagnostics) {
	if apiModel.Mappings != nil {
		mappingBytes, err := json.Marshal(apiModel.Mappings)
		if err != nil {
			return jsontypes.NewNormalizedNull(), diag.Diagnostics{
				diag.NewErrorDiagnostic("failed to marshal index mappings", err.Error()),
			}
		}

		return jsontypes.NewNormalizedValue(string(mappingBytes)), nil
	}

	return jsontypes.NewNormalizedNull(), nil
}

func setSettingsFromAPI(ctx context.Context, model *indexTfModel, apiModel estypes.IndexState) diag.Diagnostics {
	var settingsMap map[string]any
	if apiModel.Settings != nil {
		settingsBytes, err := json.Marshal(apiModel.Settings)
		if err != nil {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic("failed to marshal index settings", err.Error()),
			}
		}
		if err := json.Unmarshal(settingsBytes, &settingsMap); err != nil {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic("failed to unmarshal index settings", err.Error()),
			}
		}
	}

	modelType := reflect.TypeFor[indexTfModel]()

	for _, key := range allSettingsKeys {
		settingsValue, ok := settingsMap["index."+key]
		var tfValue attr.Value
		if !ok {
			continue
		}

		tfFieldKey := typeutils.ConvertSettingsKeyToTFFieldKey(key)
		value, ok := model.getFieldValueByTagValue(tfFieldKey, modelType)
		if !ok {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"failed to find setting value",
					fmt.Sprintf("expected setting with key %s", tfFieldKey),
				),
			}
		}

		switch a := value.(type) {
		case types.String:
			settingStr, ok := settingsValue.(string)
			if !ok {
				return diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"failed to convert setting to string",
						fmt.Sprintf("expected setting to be a string but got %t", settingsValue),
					)}
			}
			tfValue = basetypes.NewStringValue(settingStr)
		case types.Bool:
			if settingStr, ok := settingsValue.(string); ok {
				settingBool, err := strconv.ParseBool(settingStr)
				if err != nil {
					return diag.Diagnostics{
						diag.NewErrorDiagnostic(
							"failed to convert setting to bool",
							fmt.Sprintf("expected setting to be a bool but it was a string. Attempted to parse it but got %s", err.Error()),
						),
					}
				}

				settingsValue = settingBool
			}

			settingBool, ok := settingsValue.(bool)
			if !ok {
				return diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"failed to convert setting to bool",
						fmt.Sprintf("expected setting to be a bool but got %t", settingsValue),
					)}
			}
			tfValue = basetypes.NewBoolValue(settingBool)
		case types.Int64:
			if settingStr, ok := settingsValue.(string); ok {
				settingInt, err := strconv.Atoi(settingStr)
				if err != nil {
					return diag.Diagnostics{
						diag.NewErrorDiagnostic(
							"failed to convert setting to int",
							fmt.Sprintf("expected setting to be an int but it was a string. Attempted to parse it but got %s", err.Error()),
						),
					}
				}

				settingsValue = int64(settingInt)
			}

			// json.Unmarshal stores numbers as float64 in map[string]any
			if settingFloat, ok := settingsValue.(float64); ok {
				settingsValue = int64(settingFloat)
			}

			settingInt, ok := settingsValue.(int64)
			if !ok {
				return diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"failed to convert setting to int",
						fmt.Sprintf("expected setting to be a int but got %t", settingsValue),
					)}
			}
			tfValue = basetypes.NewInt64Value(settingInt)
		case types.List:
			elemType := a.ElementType(ctx)
			if elemType != types.StringType {
				return diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"expected list of string",
						fmt.Sprintf("expected list element type to be string but got %s", elemType),
					),
				}
			}

			elems, ok := settingsValue.([]any)
			if !ok {
				return diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"failed to convert setting to []string",
						fmt.Sprintf("expected setting to be a []string but got %#v", settingsValue),
					)}
			}

			var diags diag.Diagnostics
			tfValue, diags = basetypes.NewListValueFrom(ctx, basetypes.StringType{}, elems)
			if diags.HasError() {
				return diags
			}
		case types.Set:
			elemType := a.ElementType(ctx)
			if elemType != types.StringType {
				return diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"expected set of string",
						fmt.Sprintf("expected set element type to be string but got %s", elemType),
					),
				}
			}

			elems, ok := settingsValue.([]any)
			if !ok {
				return diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"failed to convert setting to []string",
						fmt.Sprintf("expected setting to be a thing []string but got %#v", settingsValue),
					)}
			}

			var diags diag.Diagnostics
			tfValue, diags = basetypes.NewSetValueFrom(ctx, basetypes.StringType{}, elems)
			if diags.HasError() {
				return diags
			}
		default:
			return diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"unknown value type",
					fmt.Sprintf("unknown index setting value type %s", a.Type(ctx)),
				),
			}
		}

		ok = model.setFieldValueByTagValue(tfFieldKey, modelType, tfValue)
		if !ok {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"failed to find setting value",
					fmt.Sprintf("expected setting with key %s", tfFieldKey),
				),
			}
		}
	}

	if apiModel.Settings != nil {
		settingsBytes, err := json.Marshal(apiModel.Settings)
		if err != nil {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"failed to marshal raw settings",
					err.Error(),
				),
			}
		}

		model.SettingsRaw = jsontypes.NewNormalizedValue(string(settingsBytes))
	} else {
		model.SettingsRaw = jsontypes.NewNormalizedNull()
	}

	return nil
}

func (model indexTfModel) getFieldValueByTagValue(tagName string, t reflect.Type) (attr.Value, bool) {
	return indexparent.GetFieldValueByTagValue(reflect.ValueOf(model), t, tagName)
}

func (model *indexTfModel) setFieldValueByTagValue(tagName string, t reflect.Type, value attr.Value) bool {
	return indexparent.SetFieldValueByTagValue(reflect.ValueOf(model), t, tagName, value)
}
