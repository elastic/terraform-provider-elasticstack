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

package alias

import (
	"context"
	"testing"

	esTypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readIndexModelWithRouting(name, routing string) readIndexModel {
	return readIndexModel{
		Name:            types.StringValue(name),
		ConcreteIndices: types.SetNull(types.StringType),
		Filter:          jsontypes.NewNormalizedNull(),
		IndexRouting:    types.StringNull(),
		IsHidden:        types.BoolValue(false),
		Routing:         types.StringValue(routing),
		SearchRouting:   types.StringNull(),
	}
}

func mustReadIndexSet(ctx context.Context, t *testing.T, indices ...readIndexModel) types.Set {
	t.Helper()

	value, diags := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: getReadIndexAttrTypes(ctx)}, indices)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	return value
}

func TestReadIndexModel_ConcreteIndices(t *testing.T) {
	t.Parallel()

	concreteIndices := types.SetValueMust(
		types.StringType,
		[]attr.Value{types.StringValue("traces-apm-default")},
	)

	model := readIndexModel{
		ConcreteIndices: concreteIndices,
	}

	require.Equal(t, concreteIndices, model.ConcreteIndices)
}

func TestTfModel_PopulateFromAPI_UsesReadIndexAttributeTypes(t *testing.T) {
	t.Parallel()

	model := tfModel{}
	diags := model.populateFromAPI(context.Background(), "logs", map[string]esTypes.AliasDefinition{
		"logs-1": {},
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
}

func TestIndexConfig_Equals(t *testing.T) {
	tests := []struct {
		name     string
		a        IndexConfig
		b        IndexConfig
		expected bool
	}{
		{
			name: "identical configs",
			a: IndexConfig{
				Name:          "test-index",
				IsWriteIndex:  true,
				Filter:        map[string]any{"user": "admin", "status": "active"},
				IndexRouting:  "1",
				IsHidden:      false,
				Routing:       "2",
				SearchRouting: "3",
			},
			b: IndexConfig{
				Name:          "test-index",
				IsWriteIndex:  true,
				Filter:        map[string]any{"user": "admin", "status": "active"},
				IndexRouting:  "1",
				IsHidden:      false,
				Routing:       "2",
				SearchRouting: "3",
			},
			expected: true,
		},
		{
			name: "different name",
			a: IndexConfig{
				Name:         "test-index-1",
				IsWriteIndex: true,
			},
			b: IndexConfig{
				Name:         "test-index-2",
				IsWriteIndex: true,
			},
			expected: false,
		},
		{
			name: "different IsWriteIndex",
			a: IndexConfig{
				Name:         "test-index",
				IsWriteIndex: true,
			},
			b: IndexConfig{
				Name:         "test-index",
				IsWriteIndex: false,
			},
			expected: false,
		},
		{
			name: "different IndexRouting",
			a: IndexConfig{
				Name:         "test-index",
				IndexRouting: "1",
			},
			b: IndexConfig{
				Name:         "test-index",
				IndexRouting: "2",
			},
			expected: false,
		},
		{
			name: "different IsHidden",
			a: IndexConfig{
				Name:     "test-index",
				IsHidden: true,
			},
			b: IndexConfig{
				Name:     "test-index",
				IsHidden: false,
			},
			expected: false,
		},
		{
			name: "different Routing",
			a: IndexConfig{
				Name:    "test-index",
				Routing: "route-1",
			},
			b: IndexConfig{
				Name:    "test-index",
				Routing: "route-2",
			},
			expected: false,
		},
		{
			name: "different SearchRouting",
			a: IndexConfig{
				Name:          "test-index",
				SearchRouting: "search-1",
			},
			b: IndexConfig{
				Name:          "test-index",
				SearchRouting: "search-2",
			},
			expected: false,
		},
		{
			name: "different Filter - different values",
			a: IndexConfig{
				Name:   "test-index",
				Filter: map[string]any{"user": "admin"},
			},
			b: IndexConfig{
				Name:   "test-index",
				Filter: map[string]any{"user": "guest"},
			},
			expected: false,
		},
		{
			name: "different Filter - different keys",
			a: IndexConfig{
				Name:   "test-index",
				Filter: map[string]any{"user": "admin"},
			},
			b: IndexConfig{
				Name:   "test-index",
				Filter: map[string]any{"role": "admin"},
			},
			expected: false,
		},
		{
			name: "one nil Filter, one non-nil",
			a: IndexConfig{
				Name:   "test-index",
				Filter: map[string]any{"term": "value"},
			},
			b: IndexConfig{
				Name:   "test-index",
				Filter: nil,
			},
			expected: false,
		},
		{
			name: "both nil Filters",
			a: IndexConfig{
				Name:   "test-index",
				Filter: nil,
			},
			b: IndexConfig{
				Name:   "test-index",
				Filter: nil,
			},
			expected: true,
		},
		{
			name: "both empty Filters",
			a: IndexConfig{
				Name:   "test-index",
				Filter: map[string]any{},
			},
			b: IndexConfig{
				Name:   "test-index",
				Filter: map[string]any{},
			},
			expected: true,
		},
		{
			name: "complex nested Filter match",
			a: IndexConfig{
				Name:   "test-index",
				Filter: map[string]any{"environment": "prod", "tier": "premium"},
			},
			b: IndexConfig{
				Name:   "test-index",
				Filter: map[string]any{"environment": "prod", "tier": "premium"},
			},
			expected: true,
		},
		{
			name: "all empty string fields",
			a: IndexConfig{
				Name:          "test-index",
				IndexRouting:  "",
				Routing:       "",
				SearchRouting: "",
			},
			b: IndexConfig{
				Name:          "test-index",
				IndexRouting:  "",
				Routing:       "",
				SearchRouting: "",
			},
			expected: true,
		},
		{
			name: "empty string vs populated string",
			a: IndexConfig{
				Name:    "test-index",
				Routing: "",
			},
			b: IndexConfig{
				Name:    "test-index",
				Routing: "route",
			},
			expected: false,
		},
		{
			name: "multiple fields different",
			a: IndexConfig{
				Name:          "test-index-1",
				IsWriteIndex:  true,
				IndexRouting:  "1",
				SearchRouting: "search-1",
			},
			b: IndexConfig{
				Name:          "test-index-2",
				IsWriteIndex:  false,
				IndexRouting:  "2",
				SearchRouting: "search-2",
			},
			expected: false,
		},
		{
			name: "fully populated identical configs",
			a: IndexConfig{
				Name:          "production-index",
				IsWriteIndex:  true,
				Filter:        map[string]any{"environment": "prod"},
				IndexRouting:  "prod-route",
				IsHidden:      true,
				Routing:       "main-route",
				SearchRouting: "search-route",
			},
			b: IndexConfig{
				Name:          "production-index",
				IsWriteIndex:  true,
				Filter:        map[string]any{"environment": "prod"},
				IndexRouting:  "prod-route",
				IsHidden:      true,
				Routing:       "main-route",
				SearchRouting: "search-route",
			},
			expected: true,
		},
		{
			name: "Filter with nested maps",
			a: IndexConfig{
				Name: "test-index",
				Filter: map[string]any{
					"term": map[string]any{"user": "admin"},
				},
			},
			b: IndexConfig{
				Name: "test-index",
				Filter: map[string]any{
					"term": map[string]any{"user": "admin"},
				},
			},
			expected: true,
		},
		{
			name: "Filter with slices",
			a: IndexConfig{
				Name: "test-index",
				Filter: map[string]any{
					"bool": map[string]any{
						"must": []any{
							map[string]any{"term": map[string]any{"status": "active"}},
						},
					},
				},
			},
			b: IndexConfig{
				Name: "test-index",
				Filter: map[string]any{
					"bool": map[string]any{
						"must": []any{
							map[string]any{"term": map[string]any{"status": "active"}},
						},
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip tests that would panic due to maps.Equal limitations
			// if tt.name == "Filter with nested maps - demonstrates maps.Equal limitation" ||
			// 	tt.name == "Filter with slices - demonstrates maps.Equal panic" {
			// 	t.Skip("This test demonstrates the limitation of maps.Equal with uncomparable types - it would panic")
			// }
			result := tt.a.Equals(tt.b)
			assert.Equal(t, tt.expected, result, "Equals() returned unexpected result")
		})
	}
}

func TestTfModel_Validate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	indexAttrTypes := getIndexAttrTypes(context.Background())
	readIndexAttrTypes := getReadIndexAttrTypes(context.Background())
	readIndexObjectType := types.ObjectType{AttrTypes: readIndexAttrTypes}

	indexModelForName := func(name types.String) readIndexModel {
		return readIndexModel{
			Name:            name,
			ConcreteIndices: types.SetValueMust(types.StringType, nil),
			Filter:          jsontypes.NewNormalizedNull(),
			IndexRouting:    types.StringNull(),
			IsHidden:        types.BoolValue(false),
			Routing:         types.StringNull(),
			SearchRouting:   types.StringNull(),
		}
	}

	mustIndexObject := func(t *testing.T, name types.String) types.Object {
		t.Helper()
		obj, diags := types.ObjectValueFrom(ctx, indexAttrTypes, struct {
			Name          types.String         `tfsdk:"name"`
			Filter        jsontypes.Normalized `tfsdk:"filter"`
			IndexRouting  types.String         `tfsdk:"index_routing"`
			IsHidden      types.Bool           `tfsdk:"is_hidden"`
			Routing       types.String         `tfsdk:"routing"`
			SearchRouting types.String         `tfsdk:"search_routing"`
		}{
			Name:          name,
			Filter:        jsontypes.NewNormalizedNull(),
			IndexRouting:  types.StringNull(),
			IsHidden:      types.BoolValue(false),
			Routing:       types.StringNull(),
			SearchRouting: types.StringNull(),
		})
		require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
		return obj
	}

	mustIndexSet := func(t *testing.T, names ...types.String) types.Set {
		t.Helper()
		indices := make([]readIndexModel, 0, len(names))
		for _, name := range names {
			indices = append(indices, indexModelForName(name))
		}
		setVal, diags := types.SetValueFrom(ctx, readIndexObjectType, indices)
		require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
		return setVal
	}

	t.Run("returns_no_error_when_write_index_is_null", func(t *testing.T) {
		t.Parallel()
		m := tfModel{
			WriteIndex:  types.ObjectNull(indexAttrTypes),
			ReadIndices: mustIndexSet(t, types.StringValue("r1")),
		}
		diags := m.Validate(ctx)
		require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	})

	t.Run("returns_no_error_when_write_index_is_unknown", func(t *testing.T) {
		t.Parallel()
		m := tfModel{
			WriteIndex:  types.ObjectUnknown(indexAttrTypes),
			ReadIndices: mustIndexSet(t, types.StringValue("r1")),
		}
		diags := m.Validate(ctx)
		require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	})

	t.Run("returns_no_error_when_read_indices_is_null", func(t *testing.T) {
		t.Parallel()
		m := tfModel{
			WriteIndex:  mustIndexObject(t, types.StringValue("w1")),
			ReadIndices: types.SetNull(readIndexObjectType),
		}
		diags := m.Validate(ctx)
		require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	})

	t.Run("returns_no_error_when_read_indices_is_unknown", func(t *testing.T) {
		t.Parallel()
		m := tfModel{
			WriteIndex:  mustIndexObject(t, types.StringValue("w1")),
			ReadIndices: types.SetUnknown(readIndexObjectType),
		}
		diags := m.Validate(ctx)
		require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	})

	t.Run("returns_error_when_write_index_name_is_in_read_indices", func(t *testing.T) {
		t.Parallel()
		m := tfModel{
			WriteIndex:  mustIndexObject(t, types.StringValue("idx")),
			ReadIndices: mustIndexSet(t, types.StringValue("idx"), types.StringValue("idx2")),
		}
		diags := m.Validate(ctx)
		require.True(t, diags.HasError())
		require.Contains(t, diags.Errors()[0].Summary(), "Invalid Configuration")
		require.Contains(t, diags.Errors()[0].Detail(), "cannot be both a write index and a read index")
	})

	t.Run("returns_no_error_when_write_and_read_names_are_distinct", func(t *testing.T) {
		t.Parallel()
		m := tfModel{
			WriteIndex:  mustIndexObject(t, types.StringValue("w1")),
			ReadIndices: mustIndexSet(t, types.StringValue("r1"), types.StringValue("r2")),
		}
		diags := m.Validate(ctx)
		require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	})

	t.Run("skips_validation_when_write_index_name_is_unknown", func(t *testing.T) {
		t.Parallel()
		m := tfModel{
			WriteIndex:  mustIndexObject(t, types.StringUnknown()),
			ReadIndices: mustIndexSet(t, types.StringValue("idx")),
		}
		diags := m.Validate(ctx)
		require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	})

	t.Run("ignores_read_indices_with_unknown_name", func(t *testing.T) {
		t.Parallel()
		m := tfModel{
			WriteIndex:  mustIndexObject(t, types.StringValue("w1")),
			ReadIndices: mustIndexSet(t, types.StringUnknown(), types.StringValue("r1")),
		}
		diags := m.Validate(ctx)
		require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	})
}

func TestTfModel_ResolveAliasConfigs_DeduplicatesOverlappingIdenticalSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := tfModel{ReadIndices: mustReadIndexSet(ctx, t,
		readIndexModelWithRouting("logs-*", "shared-route"),
		readIndexModelWithRouting("logs-current", "shared-route"),
	)}
	configs, configDiags := model.resolveAliasConfigs(ctx, func(_ context.Context, expression string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		switch expression {
		case "logs-*":
			return elasticsearch.ResolvedIndexTargets{Names: []string{"logs-1", "logs-2"}}, nil
		case "logs-current":
			return elasticsearch.ResolvedIndexTargets{Names: []string{"logs-2"}}, nil
		default:
			require.FailNowf(t, "unexpected expression", "%q", expression)
			return elasticsearch.ResolvedIndexTargets{}, nil
		}
	})

	require.False(t, configDiags.HasError(), "unexpected diagnostics: %v", configDiags.Errors())
	require.ElementsMatch(t, []IndexConfig{
		{Name: "logs-1", Routing: "shared-route"},
		{Name: "logs-2", Routing: "shared-route"},
	}, configs)
}

func TestTfModel_ResolveAliasConfigs_RejectsOverlappingConflictingSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := tfModel{ReadIndices: mustReadIndexSet(ctx, t,
		readIndexModelWithRouting("logs-*", "route-a"),
		readIndexModelWithRouting("logs-current", "route-b"),
	)}
	_, configDiags := model.resolveAliasConfigs(ctx, func(_ context.Context, expression string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		require.Contains(t, []string{"logs-*", "logs-current"}, expression)
		return elasticsearch.ResolvedIndexTargets{Names: []string{"logs-1"}}, nil
	})

	require.True(t, configDiags.HasError())
	require.Contains(t, configDiags.Errors()[0].Summary(), "Invalid Configuration")
	require.Contains(t, configDiags.Errors()[0].Detail(), "logs-1")
	require.Contains(t, configDiags.Errors()[0].Detail(), "logs-*")
	require.Contains(t, configDiags.Errors()[0].Detail(), "logs-current")
}

func TestTfModel_ResolveAliasConfigs_RejectsMixedKindsAcrossExpressions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := tfModel{ReadIndices: mustReadIndexSet(ctx, t,
		readIndexModelWithRouting("logs-*", ""),
		readIndexModelWithRouting("metrics-*", ""),
	)}
	_, configDiags := model.resolveAliasConfigs(ctx, func(_ context.Context, expression string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		switch expression {
		case "logs-*":
			return elasticsearch.ResolvedIndexTargets{
				Kind:  elasticsearch.RegularIndexTarget,
				Names: []string{"logs-1"},
			}, nil
		case "metrics-*":
			return elasticsearch.ResolvedIndexTargets{
				Kind:  elasticsearch.DataStreamTarget,
				Names: []string{"metrics-default"},
			}, nil
		default:
			require.FailNowf(t, "unexpected expression", "%q", expression)
			return elasticsearch.ResolvedIndexTargets{}, nil
		}
	})

	require.True(t, configDiags.HasError())
	require.Contains(t, configDiags.Errors()[0].Detail(), "both regular indices and data streams")
}

func TestTfModel_ResolveAliasConfigs_RejectsResolvedWriteIndexCollision(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	writeIndex, diags := types.ObjectValueFrom(ctx, getIndexAttrTypes(ctx), indexModel{
		Name:          types.StringValue("logs-current"),
		Filter:        jsontypes.NewNormalizedNull(),
		IndexRouting:  types.StringNull(),
		IsHidden:      types.BoolValue(false),
		Routing:       types.StringNull(),
		SearchRouting: types.StringNull(),
	})
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())

	model := tfModel{
		WriteIndex:  writeIndex,
		ReadIndices: mustReadIndexSet(ctx, t, readIndexModelWithRouting("logs-*", "")),
	}
	_, configDiags := model.resolveAliasConfigs(ctx, func(_ context.Context, expression string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		require.Equal(t, "logs-*", expression)
		return elasticsearch.ResolvedIndexTargets{Names: []string{"logs-current"}}, nil
	})

	require.True(t, configDiags.HasError())
	require.Contains(t, configDiags.Errors()[0].Summary(), "Invalid Configuration")
	require.Contains(t, configDiags.Errors()[0].Detail(), "logs-*")
	require.Contains(t, configDiags.Errors()[0].Detail(), "logs-current")
}

func TestTfModel_ResolveAliasConfigs_RejectsEmptyReadNameBeforeResolution(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := tfModel{
		ReadIndices: mustReadIndexSet(ctx, t, readIndexModelWithRouting("", "")),
	}
	resolveCalls := 0
	_, configDiags := model.resolveAliasConfigs(ctx, func(_ context.Context, _ string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		resolveCalls++
		return elasticsearch.ResolvedIndexTargets{}, nil
	})

	require.True(t, configDiags.HasError())
	require.Contains(t, configDiags.Errors()[0].Summary(), "Invalid Configuration")
	require.Zero(t, resolveCalls)
}

func TestTfModel_ResolveAliasConfigs_RejectsUnknownReadNameBeforeResolution(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := tfModel{
		ReadIndices: mustReadIndexSet(ctx, t, readIndexModel{
			Name:            types.StringUnknown(),
			ConcreteIndices: types.SetNull(types.StringType),
			Filter:          jsontypes.NewNormalizedNull(),
			IndexRouting:    types.StringNull(),
			IsHidden:        types.BoolValue(false),
			Routing:         types.StringNull(),
			SearchRouting:   types.StringNull(),
		}),
	}
	resolveCalls := 0
	_, configDiags := model.resolveAliasConfigs(ctx, func(_ context.Context, _ string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		resolveCalls++
		return elasticsearch.ResolvedIndexTargets{}, nil
	})

	require.True(t, configDiags.HasError())
	require.Contains(t, configDiags.Errors()[0].Summary(), "Invalid Configuration")
	require.Zero(t, resolveCalls)
}

func TestTfModel_ResolveAliasConfigs_RejectsUnknownWriteIndexBeforeResolution(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := tfModel{
		WriteIndex: types.ObjectUnknown(getIndexAttrTypes(ctx)),
	}
	resolveCalls := 0
	_, configDiags := model.resolveAliasConfigs(ctx, func(_ context.Context, _ string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		resolveCalls++
		return elasticsearch.ResolvedIndexTargets{}, nil
	})

	require.True(t, configDiags.HasError())
	require.Contains(t, configDiags.Errors()[0].Summary(), "Invalid Configuration")
	require.Zero(t, resolveCalls)
}

func TestTfModel_ResolveAliasConfigs_RejectsUnknownWriteIndexNameBeforeResolution(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	writeIndex, diags := types.ObjectValueFrom(ctx, getIndexAttrTypes(ctx), indexModel{
		Name:          types.StringUnknown(),
		Filter:        jsontypes.NewNormalizedNull(),
		IndexRouting:  types.StringNull(),
		IsHidden:      types.BoolValue(false),
		Routing:       types.StringNull(),
		SearchRouting: types.StringNull(),
	})
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())

	model := tfModel{WriteIndex: writeIndex}
	resolveCalls := 0
	_, configDiags := model.resolveAliasConfigs(ctx, func(_ context.Context, _ string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		resolveCalls++
		return elasticsearch.ResolvedIndexTargets{}, nil
	})

	require.True(t, configDiags.HasError())
	require.Contains(t, configDiags.Errors()[0].Summary(), "Invalid Configuration")
	require.Zero(t, resolveCalls)
}

func TestTfModel_ResolveAliasConfigs_RejectsKnownWriteIndexSelectorBeforeResolution(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	writeIndex, diags := types.ObjectValueFrom(ctx, getIndexAttrTypes(ctx), indexModel{
		Name:          types.StringValue("logs-*"),
		Filter:        jsontypes.NewNormalizedNull(),
		IndexRouting:  types.StringNull(),
		IsHidden:      types.BoolValue(false),
		Routing:       types.StringNull(),
		SearchRouting: types.StringNull(),
	})
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())

	model := tfModel{WriteIndex: writeIndex}
	resolveCalls := 0
	_, configDiags := model.resolveAliasConfigs(ctx, func(_ context.Context, _ string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		resolveCalls++
		return elasticsearch.ResolvedIndexTargets{}, nil
	})

	require.True(t, configDiags.HasError())
	require.Contains(t, configDiags.Errors()[0].Summary(), "Invalid Configuration")
	require.Zero(t, resolveCalls)
}

func TestTfModel_ResolveAliasConfigs_ValidatesLaterInvalidReadNameBeforeResolution(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := tfModel{
		ReadIndices: mustReadIndexSet(ctx, t,
			readIndexModelWithRouting("logs-valid", ""),
			readIndexModelWithRouting("", "")),
	}
	resolveCalls := 0
	_, configDiags := model.resolveAliasConfigs(ctx, func(_ context.Context, _ string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		resolveCalls++
		return elasticsearch.ResolvedIndexTargets{}, nil
	})

	require.True(t, configDiags.HasError())
	require.Contains(t, configDiags.Errors()[0].Summary(), "Invalid Configuration")
	require.Zero(t, resolveCalls)
}

func TestTfModel_ResolveAliasConfigs_RejectsUnknownReadIndicesBeforeResolution(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := tfModel{
		ReadIndices: types.SetUnknown(types.ObjectType{AttrTypes: getReadIndexAttrTypes(ctx)}),
	}
	resolveCalls := 0
	_, configDiags := model.resolveAliasConfigs(ctx, func(_ context.Context, _ string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		resolveCalls++
		return elasticsearch.ResolvedIndexTargets{}, nil
	})

	require.True(t, configDiags.HasError())
	require.Contains(t, configDiags.Errors()[0].Summary(), "Invalid Configuration")
	require.Zero(t, resolveCalls)
}

func TestTfModel_ModifyPlanReadIndexMembership_MarksUnknownNamesMembershipUnknown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	plan := tfModel{
		ReadIndices: mustReadIndexSet(ctx, t, readIndexModel{
			Name:            types.StringUnknown(),
			ConcreteIndices: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("logs-1")}),
			Filter:          jsontypes.NewNormalizedNull(),
			IndexRouting:    types.StringNull(),
			IsHidden:        types.BoolValue(false),
			Routing:         types.StringNull(),
			SearchRouting:   types.StringNull(),
		}),
	}
	resolveCalls := 0

	diags := plan.modifyPlanReadIndexMembership(ctx, tfModel{}, func(context.Context, string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		resolveCalls++
		return elasticsearch.ResolvedIndexTargets{}, nil
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Zero(t, resolveCalls)

	var readIndices []readIndexModel
	diags = plan.ReadIndices.ElementsAs(ctx, &readIndices, false)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Len(t, readIndices, 1)
	require.True(t, readIndices[0].ConcreteIndices.IsUnknown())
}

func TestTfModel_ModifyPlanReadIndexMembership_PreservesSetElementCorrelation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	state := tfModel{
		ReadIndices: mustReadIndexSet(ctx, t,
			readIndexModel{
				Name:            types.StringValue("logs-*"),
				ConcreteIndices: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("logs-1")}),
				Filter:          jsontypes.NewNormalizedNull(),
				IndexRouting:    types.StringNull(),
				IsHidden:        types.BoolValue(false),
				Routing:         types.StringNull(),
				SearchRouting:   types.StringNull(),
			},
			readIndexModel{
				Name:            types.StringValue("metrics-*"),
				ConcreteIndices: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("metrics-1")}),
				Filter:          jsontypes.NewNormalizedNull(),
				IndexRouting:    types.StringNull(),
				IsHidden:        types.BoolValue(false),
				Routing:         types.StringNull(),
				SearchRouting:   types.StringNull(),
			},
		),
	}
	plan := state

	diags := plan.modifyPlanReadIndexMembership(ctx, state, func(_ context.Context, expression string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		switch expression {
		case "logs-*":
			return elasticsearch.ResolvedIndexTargets{Names: []string{"logs-1", "logs-2"}}, nil
		case "metrics-*":
			return elasticsearch.ResolvedIndexTargets{Names: []string{"metrics-1"}}, nil
		default:
			require.FailNowf(t, "unexpected expression", "%q", expression)
			return elasticsearch.ResolvedIndexTargets{}, nil
		}
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	var readIndices []readIndexModel
	diags = plan.ReadIndices.ElementsAs(ctx, &readIndices, false)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Len(t, readIndices, 2)
	for _, readIndex := range readIndices {
		switch readIndex.Name.ValueString() {
		case "logs-*":
			require.True(t, readIndex.ConcreteIndices.IsUnknown())
		case "metrics-*":
			require.Equal(t, types.SetValueMust(types.StringType, []attr.Value{types.StringValue("metrics-1")}), readIndex.ConcreteIndices)
		default:
			require.FailNowf(t, "unexpected read index", "%q", readIndex.Name.ValueString())
		}
	}
}

func TestBuildAliasActions_SkipsIdenticalConcreteTarget(t *testing.T) {
	t.Parallel()

	config := IndexConfig{Name: "traces-apm-default", Routing: "shared-route"}
	actions := buildAliasActions("traces", map[string]IndexConfig{
		config.Name: config,
	}, []IndexConfig{config})

	require.Empty(t, actions)
}

func TestTfModel_BuildResolvedAliasActions_SkipsAttachedWildcardTargets(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := tfModel{ReadIndices: mustReadIndexSet(ctx, t,
		readIndexModelWithRouting("traces-apm*", "shared-route"),
	)}
	actions, diags := model.buildResolvedAliasActions(ctx, "traces", map[string]IndexConfig{
		"traces-apm-default":     {Name: "traces-apm-default", Routing: "shared-route"},
		"traces-apm.rum-default": {Name: "traces-apm.rum-default", Routing: "shared-route"},
	}, func(_ context.Context, expression string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		require.Equal(t, "traces-apm*", expression)
		return elasticsearch.ResolvedIndexTargets{Names: []string{
			"traces-apm-default",
			"traces-apm.rum-default",
		}}, nil
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Empty(t, actions)
}

func TestCurrentAliasConfigs_UsesConcreteAssociationsForAlias(t *testing.T) {
	t.Parallel()

	configs, diags := currentAliasConfigs("traces", map[string]esTypes.IndexAliases{
		"traces-apm-default": {
			Aliases: map[string]esTypes.AliasDefinition{
				"traces": {},
			},
		},
		"unrelated": {
			Aliases: map[string]esTypes.AliasDefinition{
				"other": {},
			},
		},
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Equal(t, map[string]IndexConfig{
		"traces-apm-default": {Name: "traces-apm-default"},
	}, configs)
}
