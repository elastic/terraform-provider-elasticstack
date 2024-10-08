package index

import (
	"context"
	"fmt"
	"testing"

	fuzz "github.com/google/gofuzz"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func Test_tfModel_toAPIModel(t *testing.T) {
	validAliases, diags := basetypes.NewSetValueFrom(
		context.Background(),
		aliasElementType(),
		[]aliasTfModel{
			{Name: basetypes.NewStringValue("alias-0")},
			{
				Name:          basetypes.NewStringValue("alias-1"),
				IndexRouting:  basetypes.NewStringValue("fast"),
				IsHidden:      basetypes.NewBoolValue(false),
				IsWriteIndex:  basetypes.NewBoolValue(true),
				Routing:       basetypes.NewStringValue("slow"),
				SearchRouting: basetypes.NewStringValue("just_right"),
				Filter:        jsontypes.NewNormalizedValue(`{"a": "b"}`),
			},
		},
	)
	require.Empty(t, diags)

	validSetting, diags := basetypes.NewSetValueFrom(
		context.Background(),
		settingElementType(),
		[]settingTfModel{
			{Name: basetypes.NewStringValue("number_of_replicas"), Value: basetypes.NewStringValue("5")},
		},
	)
	require.Empty(t, diags)

	validSettingsBlock, diags := basetypes.NewObjectValue(
		map[string]attr.Type{"setting": basetypes.SetType{ElemType: settingElementType()}},
		map[string]attr.Value{
			"setting": validSetting,
		},
	)
	require.Empty(t, diags)

	validSettings, diags := basetypes.NewListValue(
		settingsElementType(),
		[]attr.Value{validSettingsBlock},
	)
	require.Empty(t, diags)

	tests := []struct {
		name             string
		model            tfModel
		expectedApiModel models.Index
		hasError         bool
		expectedDiags    diag.Diagnostics
	}{
		{
			name: "should not populate aliases if null",
			model: tfModel{
				Name:     basetypes.NewStringValue("index-name"),
				Alias:    basetypes.NewSetNull(basetypes.ObjectType{}),
				Settings: basetypes.NewListNull(basetypes.ObjectType{}),
			},
			expectedApiModel: models.Index{
				Name:     "index-name",
				Settings: map[string]interface{}{},
			},
		},
		{
			name: "should not populate aliases if unknown",
			model: tfModel{
				Name:     basetypes.NewStringValue("index-name"),
				Alias:    basetypes.NewSetUnknown(basetypes.ObjectType{}),
				Settings: basetypes.NewListNull(basetypes.ObjectType{}),
			},
			expectedApiModel: models.Index{
				Name:     "index-name",
				Settings: map[string]interface{}{},
			},
		},
		{
			name: "should populate aliases if provided",
			model: tfModel{
				Name:     basetypes.NewStringValue("index-name"),
				Alias:    validAliases,
				Settings: basetypes.NewListNull(basetypes.ObjectType{}),
			},
			expectedApiModel: models.Index{
				Name:     "index-name",
				Settings: map[string]interface{}{},
				Aliases: map[string]models.IndexAlias{
					"alias-0": {Name: "alias-0"},
					"alias-1": {
						Name:          "alias-1",
						IndexRouting:  "fast",
						IsHidden:      false,
						IsWriteIndex:  true,
						Routing:       "slow",
						SearchRouting: "just_right",
						Filter: map[string]interface{}{
							"a": "b",
						},
					},
				},
			},
		},
		{
			name: "should not populate mappings if null",
			model: tfModel{
				Name:     basetypes.NewStringValue("index-name"),
				Mappings: jsontypes.NewNormalizedNull(),
				Settings: basetypes.NewListNull(basetypes.ObjectType{}),
			},
			expectedApiModel: models.Index{
				Name:     "index-name",
				Settings: map[string]interface{}{},
			},
		},
		{
			name: "should not populate mappings if unknown",
			model: tfModel{
				Name:     basetypes.NewStringValue("index-name"),
				Mappings: jsontypes.NewNormalizedUnknown(),
				Settings: basetypes.NewListNull(basetypes.ObjectType{}),
			},
			expectedApiModel: models.Index{
				Name:     "index-name",
				Settings: map[string]interface{}{},
			},
		},
		{
			name: "should unmarshall mappings if provided",
			model: tfModel{
				Name:     basetypes.NewStringValue("index-name"),
				Mappings: jsontypes.NewNormalizedValue(`{"a": "b"}`),
				Settings: basetypes.NewListNull(basetypes.ObjectType{}),
			},
			expectedApiModel: models.Index{
				Name:     "index-name",
				Settings: map[string]interface{}{},
				Mappings: map[string]interface{}{
					"a": "b",
				},
			},
		},
		{
			name: "should fail to parse a set of non-strings for sort field",
			model: tfModel{
				Name:      basetypes.NewStringValue("index-name"),
				SortField: basetypes.NewSetValueMust(basetypes.Int64Type{}, []attr.Value{basetypes.NewInt64Value(1)}),
				Settings:  basetypes.NewListNull(basetypes.ObjectType{}),
			},
			hasError: true,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"expected set of string",
					"expected set element type to be string but got basetypes.Int64Type",
				),
			},
		},
		{
			name: "should fail to parse a list of non-string for sort order",
			model: tfModel{
				Name:      basetypes.NewStringValue("index-name"),
				SortOrder: basetypes.NewListValueMust(basetypes.Int64Type{}, []attr.Value{basetypes.NewInt64Value(1)}),
				Settings:  basetypes.NewListNull(basetypes.ObjectType{}),
			},
			hasError: true,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"expected list of string",
					"expected list element type to be string but got basetypes.Int64Type",
				),
			},
		},
		{
			name: "should build settings map from model",
			model: tfModel{
				Name:                               basetypes.NewStringValue("index-name"),
				NumberOfShards:                     basetypes.NewInt64Value(3),
				NumberOfRoutingShards:              basetypes.NewInt64Value(5),
				Codec:                              basetypes.NewStringValue("codec"),
				RoutingPartitionSize:               basetypes.NewInt64Value(7),
				LoadFixedBitsetFiltersEagerly:      basetypes.NewBoolValue(true),
				ShardCheckOnStartup:                basetypes.NewStringValue("shard_check_on_startup"),
				SortField:                          basetypes.NewSetValueMust(basetypes.StringType{}, []attr.Value{basetypes.NewStringValue("sort_field")}),
				SortOrder:                          basetypes.NewListValueMust(basetypes.StringType{}, []attr.Value{basetypes.NewStringValue("sort_order")}),
				MappingCoerce:                      basetypes.NewBoolValue(false),
				NumberOfReplicas:                   basetypes.NewInt64Value(9),
				AutoExpandReplicas:                 basetypes.NewStringValue("auto_expand_replicas"),
				RefreshInterval:                    basetypes.NewStringValue("refresh_interval"),
				SearchIdleAfter:                    basetypes.NewStringValue("search.idle.after"),
				MaxResultWindow:                    basetypes.NewInt64Value(11),
				MaxInnerResultWindow:               basetypes.NewInt64Value(13),
				MaxRescoreWindow:                   basetypes.NewInt64Value(15),
				MaxDocvalueFieldsSearch:            basetypes.NewInt64Value(17),
				MaxScriptFields:                    basetypes.NewInt64Value(19),
				MaxNGramDiff:                       basetypes.NewInt64Value(21),
				MaxShingleDiff:                     basetypes.NewInt64Value(23),
				BlocksReadOnly:                     basetypes.NewBoolValue(true),
				BlocksReadOnlyAllowDelete:          basetypes.NewBoolValue(true),
				BlocksRead:                         basetypes.NewBoolValue(true),
				BlocksWrite:                        basetypes.NewBoolValue(true),
				BlocksMetadata:                     basetypes.NewBoolValue(true),
				MaxRefreshListeners:                basetypes.NewInt64Value(25),
				AnalyzeMaxTokenCount:               basetypes.NewInt64Value(27),
				HighlightMaxAnalyzedOffset:         basetypes.NewInt64Value(29),
				MaxTermsCount:                      basetypes.NewInt64Value(31),
				MaxRegexLength:                     basetypes.NewInt64Value(33),
				QueryDefaultField:                  basetypes.NewSetValueMust(basetypes.StringType{}, []attr.Value{basetypes.NewStringValue("query.default_field")}),
				RoutingAllocationEnable:            basetypes.NewStringValue("routing.allocation.enable"),
				RoutingRebalanceEnable:             basetypes.NewStringValue("routing.rebalance.enable"),
				GCDeletes:                          basetypes.NewStringValue("gc_deletes"),
				DefaultPipeline:                    basetypes.NewStringValue("default_pipeline"),
				FinalPipeline:                      basetypes.NewStringValue("final_pipeline"),
				UnassignedNodeLeftDelayedTimeout:   basetypes.NewStringValue("unassigned.node_left.delayed_timeout"),
				SearchSlowlogThresholdQueryWarn:    basetypes.NewStringValue("warn"),
				SearchSlowlogThresholdQueryInfo:    basetypes.NewStringValue("info"),
				SearchSlowlogThresholdQueryDebug:   basetypes.NewStringValue("debug"),
				SearchSlowlogThresholdQueryTrace:   basetypes.NewStringValue("trace"),
				SearchSlowlogThresholdFetchWarn:    basetypes.NewStringValue("warn"),
				SearchSlowlogThresholdFetchInfo:    basetypes.NewStringValue("info"),
				SearchSlowlogThresholdFetchDebug:   basetypes.NewStringValue("debug"),
				SearchSlowlogThresholdFetchTrace:   basetypes.NewStringValue("trace"),
				SearchSlowlogLevel:                 basetypes.NewStringValue("level"),
				IndexingSlowlogThresholdIndexWarn:  basetypes.NewStringValue("warn"),
				IndexingSlowlogThresholdIndexInfo:  basetypes.NewStringValue("info"),
				IndexingSlowlogThresholdIndexDebug: basetypes.NewStringValue("debug"),
				IndexingSlowlogThresholdIndexTrace: basetypes.NewStringValue("trace"),
				IndexingSlowlogLevel:               basetypes.NewStringValue("level"),
				IndexingSlowlogSource:              basetypes.NewStringValue("source"),
				Settings:                           basetypes.NewListNull(basetypes.ObjectType{}),
			},
			expectedApiModel: models.Index{
				Name: "index-name",
				Settings: map[string]interface{}{
					"number_of_shards":                       int64(3),
					"number_of_routing_shards":               int64(5),
					"codec":                                  "codec",
					"routing_partition_size":                 int64(7),
					"load_fixed_bitset_filters_eagerly":      true,
					"shard.check_on_startup":                 "shard_check_on_startup",
					"sort.field":                             []string{"sort_field"},
					"sort.order":                             []string{"sort_order"},
					"mapping.coerce":                         false,
					"number_of_replicas":                     int64(9),
					"auto_expand_replicas":                   "auto_expand_replicas",
					"refresh_interval":                       "refresh_interval",
					"search.idle.after":                      "search.idle.after",
					"max_result_window":                      int64(11),
					"max_inner_result_window":                int64(13),
					"max_rescore_window":                     int64(15),
					"max_docvalue_fields_search":             int64(17),
					"max_script_fields":                      int64(19),
					"max_ngram_diff":                         int64(21),
					"max_shingle_diff":                       int64(23),
					"blocks.read_only":                       true,
					"blocks.read_only_allow_delete":          true,
					"blocks.read":                            true,
					"blocks.write":                           true,
					"blocks.metadata":                        true,
					"max_refresh_listeners":                  int64(25),
					"analyze.max_token_count":                int64(27),
					"highlight.max_analyzed_offset":          int64(29),
					"max_terms_count":                        int64(31),
					"max_regex_length":                       int64(33),
					"query.default_field":                    []string{"query.default_field"},
					"routing.allocation.enable":              "routing.allocation.enable",
					"routing.rebalance.enable":               "routing.rebalance.enable",
					"gc_deletes":                             "gc_deletes",
					"default_pipeline":                       "default_pipeline",
					"final_pipeline":                         "final_pipeline",
					"unassigned.node_left.delayed_timeout":   "unassigned.node_left.delayed_timeout",
					"search.slowlog.threshold.query.warn":    "warn",
					"search.slowlog.threshold.query.info":    "info",
					"search.slowlog.threshold.query.debug":   "debug",
					"search.slowlog.threshold.query.trace":   "trace",
					"search.slowlog.threshold.fetch.warn":    "warn",
					"search.slowlog.threshold.fetch.info":    "info",
					"search.slowlog.threshold.fetch.debug":   "debug",
					"search.slowlog.threshold.fetch.trace":   "trace",
					"search.slowlog.level":                   "level",
					"indexing.slowlog.threshold.index.warn":  "warn",
					"indexing.slowlog.threshold.index.info":  "info",
					"indexing.slowlog.threshold.index.debug": "debug",
					"indexing.slowlog.threshold.index.trace": "trace",
					"indexing.slowlog.level":                 "level",
					"indexing.slowlog.source":                "source",
				},
			},
		},
		{
			name: "should parse arbitrary settings",
			model: tfModel{
				Name:     basetypes.NewStringValue("index-name"),
				Settings: validSettings,
			},
			expectedApiModel: models.Index{
				Name: "index-name",
				Settings: map[string]interface{}{
					"number_of_replicas": "5",
				},
			},
		},
		{
			name: "should fail to parse settings defined in both the type safe attribute, and arbitrary blob",
			model: tfModel{
				Name:             basetypes.NewStringValue("index-name"),
				Settings:         validSettings,
				NumberOfReplicas: basetypes.NewInt64Value(10),
			},
			hasError: true,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"duplicate setting definition",
					"setting [number_of_replicas] is both explicitly defined and included in the deprecated raw settings blocks. Please remove it from `settings` to avoid unexpected settings",
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiModel, diags := tt.model.toAPIModel(context.Background())

			if tt.hasError {
				require.NotEmpty(t, diags)
			} else {
				require.Empty(t, diags)
			}

			if tt.expectedDiags != nil {
				require.Equal(t, tt.expectedDiags, diags)
			}
			require.Equal(t, tt.expectedApiModel, apiModel)
		})
	}
}

func Test_tfModel_toPutIndexParams(t *testing.T) {
	for _, isServerless := range []bool{true, false} {
		t.Run(fmt.Sprintf("isServerless=%t", isServerless), func(t *testing.T) {
			f := fuzz.New()
			var expectedParams models.PutIndexParams
			f.Fuzz(&expectedParams)

			model := tfModel{
				MasterTimeout:       customtypes.NewDurationValue(expectedParams.MasterTimeout.String()),
				Timeout:             customtypes.NewDurationValue(expectedParams.Timeout.String()),
				WaitForActiveShards: basetypes.NewStringValue(expectedParams.WaitForActiveShards),
			}

			flavor := "not_serverless"
			if isServerless {
				flavor = "serverless"
				expectedParams.WaitForActiveShards = ""
			}

			params := model.toPutIndexParams(flavor)
			require.Equal(t, expectedParams, params)
		})
	}
}
