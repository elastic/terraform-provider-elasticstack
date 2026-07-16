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

package index

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	indexparent "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/aliasutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	// sortKeysExpandedFromNestedBlock are expanded by the Sort list-nested-attribute
	// pre-processor in toIndexSettings() and are skipped in the reflection loop
	// because they have no corresponding flat tfsdk struct field.
	sortKeysExpandedFromNestedBlock = map[string]bool{
		indexparent.SettingSortMissing: true,
		indexparent.SettingSortMode:    true,
	}

	// sortKeysSkippedOnImportHydration are not written into legacy flat attrs during
	// import hydration; nested sort is handled by populateSortFromSettings instead.
	sortKeysSkippedOnImportHydration = map[string]bool{
		indexparent.SettingSortField:   true,
		indexparent.SettingSortOrder:   true,
		indexparent.SettingSortMissing: true,
		indexparent.SettingSortMode:    true,
	}
)

type tfModel struct {
	entitycore.ResourceTimeoutsField
	ID                                 types.String              `tfsdk:"id"`
	ElasticsearchConnection            types.List                `tfsdk:"elasticsearch_connection"`
	Name                               types.String              `tfsdk:"name"`
	ConcreteName                       types.String              `tfsdk:"concrete_name"`
	NumberOfShards                     types.Int64               `tfsdk:"number_of_shards"`
	NumberOfRoutingShards              types.Int64               `tfsdk:"number_of_routing_shards"`
	Codec                              types.String              `tfsdk:"codec"`
	RoutingPartitionSize               types.Int64               `tfsdk:"routing_partition_size"`
	LoadFixedBitsetFiltersEagerly      types.Bool                `tfsdk:"load_fixed_bitset_filters_eagerly"`
	ShardCheckOnStartup                types.String              `tfsdk:"shard_check_on_startup"`
	Sort                               types.List                `tfsdk:"sort"`
	SortField                          types.Set                 `tfsdk:"sort_field"`
	SortOrder                          types.List                `tfsdk:"sort_order"`
	MappingCoerce                      types.Bool                `tfsdk:"mapping_coerce"`
	MappingTotalFieldsLimit            types.Int64               `tfsdk:"mapping_total_fields_limit"`
	NumberOfReplicas                   types.Int64               `tfsdk:"number_of_replicas"`
	AutoExpandReplicas                 types.String              `tfsdk:"auto_expand_replicas"`
	SearchIdleAfter                    types.String              `tfsdk:"search_idle_after"`
	RefreshInterval                    types.String              `tfsdk:"refresh_interval"`
	MaxResultWindow                    types.Int64               `tfsdk:"max_result_window"`
	MaxInnerResultWindow               types.Int64               `tfsdk:"max_inner_result_window"`
	MaxRescoreWindow                   types.Int64               `tfsdk:"max_rescore_window"`
	MaxDocvalueFieldsSearch            types.Int64               `tfsdk:"max_docvalue_fields_search"`
	MaxScriptFields                    types.Int64               `tfsdk:"max_script_fields"`
	MaxNGramDiff                       types.Int64               `tfsdk:"max_ngram_diff"`
	MaxShingleDiff                     types.Int64               `tfsdk:"max_shingle_diff"`
	MaxRefreshListeners                types.Int64               `tfsdk:"max_refresh_listeners"`
	AnalyzeMaxTokenCount               types.Int64               `tfsdk:"analyze_max_token_count"`
	HighlightMaxAnalyzedOffset         types.Int64               `tfsdk:"highlight_max_analyzed_offset"`
	MaxTermsCount                      types.Int64               `tfsdk:"max_terms_count"`
	MaxRegexLength                     types.Int64               `tfsdk:"max_regex_length"`
	QueryDefaultField                  types.Set                 `tfsdk:"query_default_field"`
	RoutingAllocationEnable            types.String              `tfsdk:"routing_allocation_enable"`
	RoutingRebalanceEnable             types.String              `tfsdk:"routing_rebalance_enable"`
	GCDeletes                          types.String              `tfsdk:"gc_deletes"`
	BlocksReadOnly                     types.Bool                `tfsdk:"blocks_read_only"`
	BlocksReadOnlyAllowDelete          types.Bool                `tfsdk:"blocks_read_only_allow_delete"`
	BlocksRead                         types.Bool                `tfsdk:"blocks_read"`
	BlocksWrite                        types.Bool                `tfsdk:"blocks_write"`
	BlocksMetadata                     types.Bool                `tfsdk:"blocks_metadata"`
	DefaultPipeline                    types.String              `tfsdk:"default_pipeline"`
	FinalPipeline                      types.String              `tfsdk:"final_pipeline"`
	UnassignedNodeLeftDelayedTimeout   types.String              `tfsdk:"unassigned_node_left_delayed_timeout"`
	SearchSlowlogThresholdQueryWarn    types.String              `tfsdk:"search_slowlog_threshold_query_warn"`
	SearchSlowlogThresholdQueryInfo    types.String              `tfsdk:"search_slowlog_threshold_query_info"`
	SearchSlowlogThresholdQueryDebug   types.String              `tfsdk:"search_slowlog_threshold_query_debug"`
	SearchSlowlogThresholdQueryTrace   types.String              `tfsdk:"search_slowlog_threshold_query_trace"`
	SearchSlowlogThresholdFetchWarn    types.String              `tfsdk:"search_slowlog_threshold_fetch_warn"`
	SearchSlowlogThresholdFetchInfo    types.String              `tfsdk:"search_slowlog_threshold_fetch_info"`
	SearchSlowlogThresholdFetchDebug   types.String              `tfsdk:"search_slowlog_threshold_fetch_debug"`
	SearchSlowlogThresholdFetchTrace   types.String              `tfsdk:"search_slowlog_threshold_fetch_trace"`
	SearchSlowlogLevel                 types.String              `tfsdk:"search_slowlog_level"`
	IndexingSlowlogThresholdIndexWarn  types.String              `tfsdk:"indexing_slowlog_threshold_index_warn"`
	IndexingSlowlogThresholdIndexInfo  types.String              `tfsdk:"indexing_slowlog_threshold_index_info"`
	IndexingSlowlogThresholdIndexDebug types.String              `tfsdk:"indexing_slowlog_threshold_index_debug"`
	IndexingSlowlogThresholdIndexTrace types.String              `tfsdk:"indexing_slowlog_threshold_index_trace"`
	IndexingSlowlogLevel               types.String              `tfsdk:"indexing_slowlog_level"`
	IndexingSlowlogSource              types.String              `tfsdk:"indexing_slowlog_source"`
	AnalysisAnalyzer                   jsontypes.Normalized      `tfsdk:"analysis_analyzer"`
	AnalysisTokenizer                  jsontypes.Normalized      `tfsdk:"analysis_tokenizer"`
	AnalysisCharFilter                 jsontypes.Normalized      `tfsdk:"analysis_char_filter"`
	AnalysisFilter                     jsontypes.Normalized      `tfsdk:"analysis_filter"`
	AnalysisNormalizer                 jsontypes.Normalized      `tfsdk:"analysis_normalizer"`
	Alias                              types.Set                 `tfsdk:"alias"`
	Mappings                           indexparent.MappingsValue `tfsdk:"mappings"`
	SettingsRaw                        jsontypes.Normalized      `tfsdk:"settings_raw"`
	DeletionProtection                 types.Bool                `tfsdk:"deletion_protection"`
	UseExisting                        types.Bool                `tfsdk:"use_existing"`
	WaitForActiveShards                types.String              `tfsdk:"wait_for_active_shards"`
	MasterTimeout                      customtypes.Duration      `tfsdk:"master_timeout"`
	Timeout                            customtypes.Duration      `tfsdk:"timeout"`
	Settings                           types.List                `tfsdk:"settings"`
}

type settingsTfSet struct {
	Setting types.Set `tfsdk:"setting"`
}

type settingTfModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type sortEntryModel struct {
	Field   types.String `tfsdk:"field"`
	Order   types.String `tfsdk:"order"`
	Missing types.String `tfsdk:"missing"`
	Mode    types.String `tfsdk:"mode"`
}

func (model *tfModel) populateFromAPI(ctx context.Context, indexName string, apiModel estypes.IndexState, hydrateAll bool) diag.Diagnostics {
	// Always set the concrete name to the actual index name from Elasticsearch.
	model.ConcreteName = types.StringValue(indexName)

	// Preserve the configured name if it is already in state.
	// Only backfill name from the concrete index name when state has no name
	// (e.g. after an import where no date math expression is known).
	if model.Name.IsNull() || model.Name.IsUnknown() || model.Name.ValueString() == "" {
		model.Name = types.StringValue(indexName)
	}

	modelAliases, diags := aliasutil.AliasesFromAPI(ctx, apiModel.Aliases, aliasElementType(ctx))
	if diags.HasError() {
		return diags
	}

	model.Alias = modelAliases

	if apiModel.Mappings != nil {
		mappingBytes, err := json.Marshal(apiModel.Mappings)
		if err != nil {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic("failed to marshal index mappings", err.Error()),
			}
		}
		model.Mappings = indexparent.NewMappingsValue(string(mappingBytes))
	}

	diags = setSettingsFromAPI(model, apiModel)
	if diags.HasError() {
		return diags
	}

	// Populate sort settings from the API response based on the current state shape.
	if model.Sort.IsNull() || model.Sort.IsUnknown() {
		// Legacy path: only populate SortField/SortOrder when they were already
		// configured (known and non-null) in state. Populating them for resources
		// that never set these attributes (e.g. index templates with sort settings)
		// would introduce perpetual diffs.
		if (!model.SortField.IsNull() && !model.SortField.IsUnknown()) ||
			(!model.SortOrder.IsNull() && !model.SortOrder.IsUnknown()) {
			if legDiags := populateLegacySortFromSettings(ctx, model); legDiags.HasError() {
				return legDiags
			}
		}
	} else {
		// New path: populate Sort from ES response.
		if sortDiags := populateSortFromSettings(ctx, model); sortDiags.HasError() {
			return sortDiags
		}
	}

	if hydrateAll {
		if hydrateDiags := hydrateAllSettingsFromRaw(ctx, model); hydrateDiags.HasError() {
			return hydrateDiags
		}
		populateOperationalDefaults(model)
	}

	return nil
}

func setSettingsFromAPI(model *tfModel, apiModel estypes.IndexState) diag.Diagnostics {
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

func (model tfModel) toAPIModel(ctx context.Context) (models.Index, diag.Diagnostics) {
	var diags diag.Diagnostics

	apiModel := models.Index{
		Name:     model.Name.ValueString(),
		Settings: map[string]any{},
	}

	if typeutils.IsKnown(model.Alias) {
		apiModel.Aliases = map[string]models.IndexAlias{}

		var planAliases []aliasutil.AliasModel
		diags.Append(model.Alias.ElementsAs(ctx, &planAliases, true)...)
		if diags.HasError() {
			return models.Index{}, diags
		}

		for _, planAlias := range planAliases {
			apiAlias, diags := aliasToAPIModel(planAlias)
			if diags.HasError() {
				return models.Index{}, diags
			}

			apiModel.Aliases[apiAlias.Name] = apiAlias
		}
	}

	settings, diags := model.toIndexSettings(ctx)
	if diags.HasError() {
		return models.Index{}, diags
	}

	apiModel.Settings = settings

	if typeutils.IsKnown(model.Mappings) {
		diags.Append(model.Mappings.Unmarshal(&apiModel.Mappings)...)
		if diags.HasError() {
			return models.Index{}, diags
		}
	}

	return apiModel, diags
}

func (model tfModel) toPutIndexParams(isServerless bool) models.PutIndexParams {
	// The string values are validated as durations as part of schema validation
	masterTimeout, _ := model.MasterTimeout.Parse()
	timeout, _ := model.Timeout.Parse()

	params := models.PutIndexParams{
		Timeout: timeout,
	}

	if !isServerless {
		params.MasterTimeout = masterTimeout
		params.WaitForActiveShards = model.WaitForActiveShards.ValueString()
	}

	return params
}

// GetID satisfies [entitycore.ElasticsearchResourceModel].
func (model tfModel) GetID() types.String { return model.ID }

// GetResourceID satisfies [entitycore.ElasticsearchResourceModel].
// Returns the configured index name (which may be a date-math expression)
// used as the write identity on create.
func (model tfModel) GetResourceID() types.String { return model.Name }

// GetElasticsearchConnection satisfies [entitycore.ElasticsearchResourceModel].
func (model tfModel) GetElasticsearchConnection() types.List { return model.ElasticsearchConnection }

func (model tfModel) getCompositeID() (*clients.CompositeID, diag.Diagnostics) {
	compID, compIDDiags := clients.CompositeIDFromStr(model.ID.ValueString())
	if compIDDiags.HasError() {
		return nil, compIDDiags
	}

	return compID, nil
}

func (model tfModel) toIndexSettings(ctx context.Context) (map[string]any, diag.Diagnostics) {
	settings := map[string]any{}
	modelType := reflect.TypeFor[tfModel]()

	// Pre-process sort ListNestedAttribute: expand to flat settings keys.
	if typeutils.IsKnown(model.Sort) {
		var sortEntries []sortEntryModel
		if diags := model.Sort.ElementsAs(ctx, &sortEntries, false); diags.HasError() {
			return map[string]any{}, diags
		}

		if len(sortEntries) > 0 {
			sortFields := make([]string, len(sortEntries))
			sortOrders := make([]string, len(sortEntries))
			sortMissing := make([]string, len(sortEntries))
			sortModes := make([]string, len(sortEntries))

			allMissingNull := true
			allModeNull := true

			for i, entry := range sortEntries {
				sortFields[i] = entry.Field.ValueString()

				if entry.Order.IsNull() || entry.Order.IsUnknown() {
					sortOrders[i] = sortOrderAsc
				} else {
					sortOrders[i] = entry.Order.ValueString()
				}

				if !entry.Missing.IsNull() && !entry.Missing.IsUnknown() {
					sortMissing[i] = entry.Missing.ValueString()
					allMissingNull = false
				}
				// else: sortMissing[i] stays "" (empty placeholder for positional alignment)

				if !entry.Mode.IsNull() && !entry.Mode.IsUnknown() {
					sortModes[i] = entry.Mode.ValueString()
					allModeNull = false
				}
				// else: sortModes[i] stays "" (empty placeholder for positional alignment)
			}

			settings[indexparent.SettingSortField] = sortFields
			settings[indexparent.SettingSortOrder] = sortOrders

			if !allMissingNull {
				settings[indexparent.SettingSortMissing] = sortMissing
			}
			if !allModeNull {
				settings[indexparent.SettingSortMode] = sortModes
			}
		}
	}

	for _, key := range indexparent.AllSettingsKeys {
		// sort.missing and sort.mode are only populated by the pre-processor above.
		// They have no flat tfsdk struct field to reflect on, so skip them here.
		if sortKeysExpandedFromNestedBlock[key] {
			continue
		}

		tfFieldKey := typeutils.ConvertSettingsKeyToTFFieldKey(key)
		value, ok := model.getFieldValueByTagValue(tfFieldKey, modelType)
		if !ok {
			return map[string]any{}, diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"failed to find setting value",
					fmt.Sprintf("expected setting with key %s", tfFieldKey),
				),
			}
		}

		if !value.IsNull() && !value.IsUnknown() {
			var settingsValue any
			switch a := value.(type) {
			case types.String:
				settingsValue = a.ValueString()
			case types.Bool:
				settingsValue = a.ValueBool()
			case types.Int64:
				settingsValue = a.ValueInt64()
			case types.List:
				elemType := a.ElementType(ctx)
				if elemType != types.StringType {
					return map[string]any{}, diag.Diagnostics{
						diag.NewErrorDiagnostic(
							"expected list of string",
							fmt.Sprintf("expected list element type to be string but got %s", elemType),
						),
					}
				}

				elems := []string{}
				if diags := a.ElementsAs(ctx, &elems, true); diags.HasError() {
					return map[string]any{}, diags
				}

				settingsValue = elems
			case types.Set:
				elemType := a.ElementType(ctx)
				if elemType != types.StringType {
					return map[string]any{}, diag.Diagnostics{
						diag.NewErrorDiagnostic(
							"expected set of string",
							fmt.Sprintf("expected set element type to be string but got %s", elemType),
						),
					}
				}

				elems := []string{}
				if diags := a.ElementsAs(ctx, &elems, true); diags.HasError() {
					return map[string]any{}, diags
				}

				settingsValue = elems
			default:
				return map[string]any{}, diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"unknown value type",
						fmt.Sprintf("unknown index setting value type %s", a.Type(ctx)),
					),
				}
			}

			settings[key] = settingsValue
		}
	}

	analysis := map[string]any{}
	for name, property := range analysisNormalizedFields(model) {
		if typeutils.IsKnown(property) {
			var parsedValue map[string]any
			if diags := property.Unmarshal(&parsedValue); diags.HasError() {
				return map[string]any{}, diags
			}

			analysis[name] = parsedValue
		}
	}

	if len(analysis) > 0 {
		settings["analysis"] = analysis
	}

	var settingSet []settingsTfSet
	if typeutils.IsKnown(model.Settings) {
		if diags := model.Settings.ElementsAs(ctx, &settingSet, true); diags.HasError() {
			return map[string]any{}, diags
		}
	}

	if len(settingSet) == 1 {
		var rawSettings []settingTfModel
		if diags := settingSet[0].Setting.ElementsAs(ctx, &rawSettings, true); diags.HasError() {
			return map[string]any{}, diags
		}

		for _, setting := range rawSettings {
			name := setting.Name.ValueString()
			if _, ok := settings[name]; ok {
				return map[string]any{}, diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"duplicate setting definition",
						fmt.Sprintf("setting [%s] is both explicitly defined and included in the deprecated raw settings blocks. Please remove it from `settings` to avoid unexpected settings", name),
					),
				}
			}

			settings[name] = setting.Value.ValueString()
		}
	}

	return settings, nil
}

func (model tfModel) getFieldValueByTagValue(tagName string, t reflect.Type) (attr.Value, bool) {
	return indexparent.GetFieldValueByTagValue(reflect.ValueOf(model), t, tagName)
}

// analysisNormalizedFields returns index.analysis sub-keys mapped to model fields.
func analysisNormalizedFields(model tfModel) map[string]jsontypes.Normalized {
	return map[string]jsontypes.Normalized{
		"analyzer":    model.AnalysisAnalyzer,
		"tokenizer":   model.AnalysisTokenizer,
		"char_filter": model.AnalysisCharFilter,
		attrFilter:    model.AnalysisFilter,
		"normalizer":  model.AnalysisNormalizer,
	}
}

// analysisNormalizedFieldTargets returns writable targets for import hydration.
func analysisNormalizedFieldTargets(model *tfModel) map[string]*jsontypes.Normalized {
	return map[string]*jsontypes.Normalized{
		"analyzer":    &model.AnalysisAnalyzer,
		"tokenizer":   &model.AnalysisTokenizer,
		"char_filter": &model.AnalysisCharFilter,
		attrFilter:    &model.AnalysisFilter,
		"normalizer":  &model.AnalysisNormalizer,
	}
}

func aliasToAPIModel(model aliasutil.AliasModel) (models.IndexAlias, diag.Diagnostics) {
	apiModel := models.IndexAlias{
		Name:          model.Name.ValueString(),
		IndexRouting:  model.IndexRouting.ValueString(),
		IsHidden:      model.IsHidden.ValueBool(),
		IsWriteIndex:  model.IsWriteIndex.ValueBool(),
		Routing:       model.Routing.ValueString(),
		SearchRouting: model.SearchRouting.ValueString(),
	}

	if typeutils.IsKnown(model.Filter) {
		if diags := model.Filter.Unmarshal(&apiModel.Filter); diags.HasError() {
			return models.IndexAlias{}, diags
		}
	}

	return apiModel, nil
}

func indexStateToModel(state estypes.IndexState) (models.Index, diag.Diagnostics) {
	var model models.Index

	if len(state.Aliases) > 0 {
		model.Aliases = make(map[string]models.IndexAlias, len(state.Aliases))
		for name, alias := range state.Aliases {
			indexAlias := models.IndexAlias{
				Name:          name,
				IndexRouting:  typeutils.Deref(alias.IndexRouting),
				IsHidden:      typeutils.Deref(alias.IsHidden),
				IsWriteIndex:  typeutils.Deref(alias.IsWriteIndex),
				Routing:       typeutils.Deref(alias.Routing),
				SearchRouting: typeutils.Deref(alias.SearchRouting),
			}
			if alias.Filter != nil {
				filterMap, diags := aliasutil.NormalizeAliasFilterAnyToMap(alias.Filter)
				if diags.HasError() {
					return models.Index{}, diags
				}
				indexAlias.Filter = filterMap
			}
			model.Aliases[name] = indexAlias
		}
	}

	if state.Mappings != nil {
		mappingBytes, err := json.Marshal(state.Mappings)
		if err != nil {
			return models.Index{}, diag.Diagnostics{
				diag.NewErrorDiagnostic("failed to marshal index mappings", err.Error()),
			}
		}
		if err := json.Unmarshal(mappingBytes, &model.Mappings); err != nil {
			return models.Index{}, diag.Diagnostics{
				diag.NewErrorDiagnostic("failed to unmarshal index mappings", err.Error()),
			}
		}
	}

	if state.Settings != nil {
		settingsBytes, err := json.Marshal(state.Settings)
		if err != nil {
			return models.Index{}, diag.Diagnostics{
				diag.NewErrorDiagnostic("failed to marshal index settings", err.Error()),
			}
		}
		if err := json.Unmarshal(settingsBytes, &model.Settings); err != nil {
			return models.Index{}, diag.Diagnostics{
				diag.NewErrorDiagnostic("failed to unmarshal index settings", err.Error()),
			}
		}
	}

	return model, nil
}

func modelToIndexState(model models.Index) (estypes.IndexState, diag.Diagnostics) {
	bytes, err := json.Marshal(model)
	if err != nil {
		return estypes.IndexState{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("failed to marshal index model", err.Error()),
		}
	}

	var state estypes.IndexState
	if err := json.Unmarshal(bytes, &state); err != nil {
		return estypes.IndexState{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("failed to unmarshal to typed index state", err.Error()),
		}
	}

	return state, nil
}
