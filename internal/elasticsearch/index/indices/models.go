package indices

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	staticSettingsKeys = []string{
		"number_of_shards",
		"number_of_routing_shards",
		"codec",
		"routing_partition_size",
		"load_fixed_bitset_filters_eagerly",
		"shard.check_on_startup",
		"sort.field",
		"sort.order",
		"mapping.coerce",
	}
	dynamicSettingsKeys = []string{
		"number_of_replicas",
		"auto_expand_replicas",
		"refresh_interval",
		"search.idle.after",
		"max_result_window",
		"max_inner_result_window",
		"max_rescore_window",
		"max_docvalue_fields_search",
		"max_script_fields",
		"max_ngram_diff",
		"max_shingle_diff",
		"blocks.read_only",
		"blocks.read_only_allow_delete",
		"blocks.read",
		"blocks.write",
		"blocks.metadata",
		"max_refresh_listeners",
		"analyze.max_token_count",
		"highlight.max_analyzed_offset",
		"max_terms_count",
		"max_regex_length",
		"query.default_field",
		"routing.allocation.enable",
		"routing.rebalance.enable",
		"gc_deletes",
		"default_pipeline",
		"final_pipeline",
		"unassigned.node_left.delayed_timeout",
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
	allSettingsKeys = []string{}
)

func init() {
	allSettingsKeys = append(allSettingsKeys, staticSettingsKeys...)
	allSettingsKeys = append(allSettingsKeys, dynamicSettingsKeys...)
}

type tfModel struct {
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

type aliasTfModel struct {
	Name          types.String         `tfsdk:"name"`
	Filter        jsontypes.Normalized `tfsdk:"filter"`
	IndexRouting  types.String         `tfsdk:"index_routing"`
	IsHidden      types.Bool           `tfsdk:"is_hidden"`
	IsWriteIndex  types.Bool           `tfsdk:"is_write_index"`
	Routing       types.String         `tfsdk:"routing"`
	SearchRouting types.String         `tfsdk:"search_routing"`
}

func (model *indexTfModel) populateFromAPI(ctx context.Context, indexName string, apiModel models.Index) diag.Diagnostics {
	model.Name = types.StringValue(indexName)
	model.SortField = types.SetValueMust(types.StringType, []attr.Value{})
	model.SortOrder = types.ListValueMust(types.StringType, []attr.Value{})
	model.QueryDefaultField = types.SetValueMust(types.StringType, []attr.Value{})

	modelMappings, diags := mappingsFromAPI(apiModel)
	if diags.HasError() {
		return diags
	}
	modelAliases, diags := aliasesFromAPI(ctx, apiModel)
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

func mappingsFromAPI(apiModel models.Index) (jsontypes.Normalized, diag.Diagnostics) {
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

func aliasesFromAPI(ctx context.Context, apiModel models.Index) (basetypes.SetValue, diag.Diagnostics) {
	aliases := []aliasTfModel{}
	for name, alias := range apiModel.Aliases {
		tfAlias, diags := newAliasModelFromAPI(name, alias)
		if diags.HasError() {
			return basetypes.SetValue{}, diags
		}

		aliases = append(aliases, tfAlias)
	}

	modelAliases, diags := types.SetValueFrom(ctx, aliasElementType(), aliases)
	if diags.HasError() {
		return basetypes.SetValue{}, diags
	}

	return modelAliases, nil
}

func newAliasModelFromAPI(name string, apiModel models.IndexAlias) (aliasTfModel, diag.Diagnostics) {
	tfAlias := aliasTfModel{
		Name:          types.StringValue(name),
		IndexRouting:  types.StringValue(apiModel.IndexRouting),
		IsHidden:      types.BoolValue(apiModel.IsHidden),
		IsWriteIndex:  types.BoolValue(apiModel.IsWriteIndex),
		Routing:       types.StringValue(apiModel.Routing),
		SearchRouting: types.StringValue(apiModel.SearchRouting),
	}

	if apiModel.Filter != nil {
		filterBytes, err := json.Marshal(apiModel.Filter)
		if err != nil {
			return aliasTfModel{}, diag.Diagnostics{
				diag.NewErrorDiagnostic("failed to marshal alias filter", err.Error()),
			}
		}

		tfAlias.Filter = jsontypes.NewNormalizedValue(string(filterBytes))
	}

	return tfAlias, nil
}

func setSettingsFromAPI(ctx context.Context, model *indexTfModel, apiModel models.Index) diag.Diagnostics {
	modelType := reflect.TypeOf(*model)

	for _, key := range allSettingsKeys {
		settingsValue, ok := apiModel.Settings["index."+key]
		var tfValue attr.Value
		if !ok {
			continue
		}

		tfFieldKey := utils.ConvertSettingsKeyToTFFieldKey(key)
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
							fmt.Sprintf("expected setting to be an in but it was a string. Attempted to parse it but got %s", err.Error()),
						),
					}
				}

				settingsValue = int64(settingInt)
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

			elems, ok := settingsValue.([]interface{})
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

			elems, ok := settingsValue.([]interface{})
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

	return nil
}

func (model indexTfModel) getFieldValueByTagValue(tagName string, t reflect.Type) (attr.Value, bool) {
	numField := t.NumField()
	for i := 0; i < numField; i++ {
		field := t.Field(i)
		if field.Tag.Get("tfsdk") == tagName {
			return reflect.ValueOf(model).Field(i).Interface().(attr.Value), true
		}
	}

	return nil, false
}

func (model *indexTfModel) setFieldValueByTagValue(tagName string, t reflect.Type, value attr.Value) bool {
	numField := t.NumField()
	for i := 0; i < numField; i++ {
		field := t.Field(i)
		if field.Tag.Get("tfsdk") == tagName {
			reflect.ValueOf(model).Elem().Field(i).Set(reflect.ValueOf(value))
			return true
		}
	}

	return false
}
