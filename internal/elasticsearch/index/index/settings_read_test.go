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
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func Test_hydrateAllSettingsFromRaw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name     string
		raw      string
		assertFn func(t *testing.T, model *tfModel)
	}{
		{
			name: "int64 string conversion",
			raw:  `{"index.number_of_replicas":"2"}`,
			assertFn: func(t *testing.T, model *tfModel) {
				require.Equal(t, int64(2), model.NumberOfReplicas.ValueInt64())
			},
		},
		{
			name: "bool string conversion",
			raw:  `{"index.blocks.read":"true"}`,
			assertFn: func(t *testing.T, model *tfModel) {
				require.True(t, model.BlocksRead.ValueBool())
			},
		},
		{
			name: "string passthrough",
			raw:  `{"index.refresh_interval":"30s"}`,
			assertFn: func(t *testing.T, model *tfModel) {
				require.Equal(t, "30s", model.RefreshInterval.ValueString())
			},
		},
		{
			name: "flat analysis analyzer keys",
			raw: `{
				"index.analysis.analyzer.import_test.filter":["lowercase"],
				"index.analysis.analyzer.import_test.tokenizer":"standard",
				"index.analysis.analyzer.import_test.type":"custom"
			}`,
			assertFn: func(t *testing.T, model *tfModel) {
				require.JSONEq(t,
					`{"import_test":{"filter":["lowercase"],"tokenizer":"standard","type":"custom"}}`,
					model.AnalysisAnalyzer.ValueString(),
				)
			},
		},
		{
			name: "nested index.analysis object",
			raw: `{
				"index.analysis":{
					"analyzer":{"nested":{"type":"custom","tokenizer":"keyword"}}
				}
			}`,
			assertFn: func(t *testing.T, model *tfModel) {
				require.JSONEq(t,
					`{"nested":{"tokenizer":"keyword","type":"custom"}}`,
					model.AnalysisAnalyzer.ValueString(),
				)
			},
		},
		{
			name: "query default field array",
			raw:  `{"index.query.default_field":["field1","field2"]}`,
			assertFn: func(t *testing.T, model *tfModel) {
				elems := make([]string, 0, 2)
				diags := model.QueryDefaultField.ElementsAs(ctx, &elems, false)
				require.Empty(t, diags)
				require.ElementsMatch(t, []string{"field1", "field2"}, elems)
			},
		},
		{
			name: "query default field JSON string array",
			raw:  `{"index.query.default_field":"[\"field1\"]"}`,
			assertFn: func(t *testing.T, model *tfModel) {
				elems := make([]string, 0, 1)
				diags := model.QueryDefaultField.ElementsAs(ctx, &elems, false)
				require.Empty(t, diags)
				require.Equal(t, []string{"field1"}, elems)
			},
		},
		{
			name: "missing keys are no-op",
			raw:  `{"index.refresh_interval":"10s"}`,
			assertFn: func(t *testing.T, model *tfModel) {
				require.True(t, model.NumberOfReplicas.IsNull())
			},
		},
		{
			name: "invalid int64 conversion is skipped",
			raw:  `{"index.number_of_replicas":"not-a-number"}`,
			assertFn: func(t *testing.T, model *tfModel) {
				require.True(t, model.NumberOfReplicas.IsNull())
			},
		},
		{
			name: "sort field keys are not hydrated into legacy attrs",
			raw: `{
				"index.sort.field":["sort_key"],
				"index.sort.order":["asc"]
			}`,
			assertFn: func(t *testing.T, model *tfModel) {
				require.True(t, model.SortField.IsNull() || model.SortField.IsUnknown())
				require.True(t, model.SortOrder.IsNull() || model.SortOrder.IsUnknown())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			model := &tfModel{
				SettingsRaw: jsontypes.NewNormalizedValue(tt.raw),
			}
			diags := hydrateAllSettingsFromRaw(ctx, model)
			require.Empty(t, diags)
			tt.assertFn(t, model)
		})
	}
}

func Test_hydrateAnalysisFromFlatSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := &tfModel{
		SettingsRaw: jsontypes.NewNormalizedValue(`{
			"index.analysis.tokenizer.ngram.type": "ngram",
			"index.analysis.tokenizer.ngram.min_gram": "3",
			"index.analysis.tokenizer.ngram.max_gram": "5",
			"index.analysis.filter.lower.type": "lowercase",
			"index.analysis.filter.boolish.type": "stop",
			"index.analysis.filter.boolish.ignore_case": "true"
		}`),
	}
	diags := hydrateAllSettingsFromRaw(ctx, model)
	require.Empty(t, diags)
	require.JSONEq(t,
		`{"ngram":{"type":"ngram","min_gram":3,"max_gram":5}}`,
		model.AnalysisTokenizer.ValueString(),
	)
	require.JSONEq(t,
		`{"boolish":{"type":"stop","ignore_case":true},"lower":{"type":"lowercase"}}`,
		model.AnalysisFilter.ValueString(),
	)
}

func Test_pruneImportHydratedPlanFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("config null prunes plan field", func(t *testing.T) {
		plan := &tfModel{
			NumberOfReplicas: types.Int64Value(1),
			RefreshInterval:  types.StringValue("30s"),
		}
		config := &tfModel{
			RefreshInterval: types.StringValue("30s"),
		}
		pruneImportHydratedPlanFields(ctx, plan, config)
		require.True(t, plan.NumberOfReplicas.IsNull())
		require.Equal(t, "30s", plan.RefreshInterval.ValueString())
	})

	t.Run("config set preserves plan field", func(t *testing.T) {
		plan := &tfModel{
			NumberOfReplicas: types.Int64Value(1),
		}
		config := &tfModel{
			NumberOfReplicas: types.Int64Value(1),
		}
		pruneImportHydratedPlanFields(ctx, plan, config)
		require.Equal(t, int64(1), plan.NumberOfReplicas.ValueInt64())
	})

	t.Run("config unknown leaves plan field", func(t *testing.T) {
		plan := &tfModel{
			BlocksRead: types.BoolValue(true),
		}
		config := &tfModel{
			BlocksRead: types.BoolUnknown(),
		}
		pruneImportHydratedPlanFields(ctx, plan, config)
		require.True(t, plan.BlocksRead.ValueBool())
	})

	t.Run("analysis config null prunes normalized field", func(t *testing.T) {
		plan := &tfModel{
			AnalysisAnalyzer: jsontypes.NewNormalizedValue(`{"a":{"type":"custom"}}`),
		}
		config := &tfModel{
			AnalysisAnalyzer: jsontypes.NewNormalizedNull(),
		}
		pruneImportHydratedPlanFields(ctx, plan, config)
		require.True(t, plan.AnalysisAnalyzer.IsNull())
	})
}
