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
	"github.com/stretchr/testify/require"
)

func TestTfModel_PopulateReadState_RetainsConfiguredExpressionWithAllAttachedTargets(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := tfModel{
		ReadIndices: mustReadIndexSet(ctx, t, readIndexModelWithRouting("traces-apm*", "configured-route")),
	}
	apiRoute := "api-route"
	diags := model.populateReadState(ctx, "traces", map[string]esTypes.AliasDefinition{
		"traces-apm-default":     {Routing: &apiRoute},
		"traces-apm.rum-default": {Routing: &apiRoute},
	}, func(_ context.Context, expression string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		require.Equal(t, "traces-apm*", expression)
		return elasticsearch.ResolvedIndexTargets{Names: []string{
			"traces-apm-default",
			"traces-apm.rum-default",
		}}, nil
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())

	var readIndices []readIndexModel
	diags = model.ReadIndices.ElementsAs(ctx, &readIndices, false)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Len(t, readIndices, 1)
	require.Equal(t, "traces-apm*", readIndices[0].Name.ValueString())
	require.Equal(t, "configured-route", readIndices[0].Routing.ValueString())
	require.Equal(t, types.SetValueMust(types.StringType, []attr.Value{
		types.StringValue("traces-apm-default"),
		types.StringValue("traces-apm.rum-default"),
	}), readIndices[0].ConcreteIndices)
}

func TestReadAliasIntoModelWithResolution_RecordsOnlyAttachedResolvedTargets(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := tfModel{
		ReadIndices: mustReadIndexSet(ctx, t, readIndexModelWithRouting("traces-apm*", "")),
	}
	diags := readAliasIntoModelWithResolution(ctx, "traces", map[string]esTypes.IndexAliases{
		"traces-apm-default": {
			Aliases: map[string]esTypes.AliasDefinition{"traces": {}},
		},
	}, &model, func(_ context.Context, expression string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		require.Equal(t, "traces-apm*", expression)
		return elasticsearch.ResolvedIndexTargets{Names: []string{
			"traces-apm-default",
			"traces-apm.rum-default",
		}}, nil
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	var readIndices []readIndexModel
	diags = model.ReadIndices.ElementsAs(ctx, &readIndices, false)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Len(t, readIndices, 1)
	require.Equal(t, types.SetValueMust(types.StringType, []attr.Value{
		types.StringValue("traces-apm-default"),
	}), readIndices[0].ConcreteIndices)
}

func TestTfModel_PopulateReadState_RetainsConfiguredExpressionWithNoAttachedTargets(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := tfModel{
		ReadIndices: mustReadIndexSet(ctx, t, readIndexModelWithRouting("traces-apm*", "")),
	}
	diags := model.populateReadState(ctx, "traces", map[string]esTypes.AliasDefinition{
		"unrelated": {},
	}, func(_ context.Context, expression string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		require.Equal(t, "traces-apm*", expression)
		return elasticsearch.ResolvedIndexTargets{}, nil
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	var readIndices []readIndexModel
	diags = model.ReadIndices.ElementsAs(ctx, &readIndices, false)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Len(t, readIndices, 2)
	requireReadIndexConcreteIndices(ctx, t, readIndices, "traces-apm*")
	requireReadIndexConcreteIndices(ctx, t, readIndices, "unrelated", "unrelated")
}

func TestTfModel_IsVirtualState_RetainsEmptyConfiguredMembership(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := tfModel{
		WriteIndex: types.ObjectNull(getIndexAttrTypes(ctx)),
		ReadIndices: mustReadIndexSet(ctx, t, readIndexModel{
			Name:            types.StringValue("traces-apm*"),
			ConcreteIndices: types.SetValueMust(types.StringType, nil),
			Filter:          jsontypes.NewNormalizedNull(),
			IndexRouting:    types.StringNull(),
			IsHidden:        types.BoolValue(false),
			Routing:         types.StringNull(),
			SearchRouting:   types.StringNull(),
		}),
	}

	require.True(t, model.isVirtualState(ctx))
}

func TestReadAliasIntoModelWithResolution_RetainsVirtualStateWhenNewTargetMatches(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := tfModel{
		WriteIndex: types.ObjectNull(getIndexAttrTypes(ctx)),
		ReadIndices: mustReadIndexSet(ctx, t, readIndexModel{
			Name:            types.StringValue("traces-apm*"),
			ConcreteIndices: types.SetValueMust(types.StringType, nil),
			Filter:          jsontypes.NewNormalizedNull(),
			IndexRouting:    types.StringNull(),
			IsHidden:        types.BoolValue(false),
			Routing:         types.StringNull(),
			SearchRouting:   types.StringNull(),
		}),
	}
	resolveCalls := 0

	diags := readAliasIntoModelWithResolution(ctx, "traces", map[string]esTypes.IndexAliases{}, &model, func(context.Context, string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		resolveCalls++
		return elasticsearch.ResolvedIndexTargets{Names: []string{"traces-apm-default"}}, nil
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Zero(t, resolveCalls)
	requireReadModelConcreteIndices(ctx, t, model, "traces-apm*")
}

func TestReadAliasIntoModelWithResolution_RetainsVirtualStateAfterUpdateToNoMatches(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := tfModel{
		WriteIndex:  types.ObjectNull(getIndexAttrTypes(ctx)),
		ReadIndices: mustReadIndexSet(ctx, t, virtualReadIndexModel("logs-*")),
	}

	diags := readAliasIntoModelWithResolution(ctx, "logs", map[string]esTypes.IndexAliases{}, &model, func(context.Context, string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		require.FailNow(t, "resolver should not be called for a desired-empty post-update read")
		return elasticsearch.ResolvedIndexTargets{}, nil
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	requireReadModelConcreteIndices(ctx, t, model, "logs-*")
}

func TestReadAliasIntoModelWithResolution_RetainsDesiredEmptyPostUpdateThenRefreshesCleanly(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := tfModel{
		WriteIndex: types.ObjectNull(getIndexAttrTypes(ctx)),
		ReadIndices: mustReadIndexSet(ctx, t, readIndexModel{
			Name:            types.StringValue("logs-*"),
			ConcreteIndices: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("logs-1")}),
			Filter:          jsontypes.NewNormalizedNull(),
			IndexRouting:    types.StringNull(),
			IsHidden:        types.BoolValue(false),
			Routing:         types.StringValue("configured-route"),
			SearchRouting:   types.StringNull(),
		}),
	}

	diags := model.markDesiredEmptyAfterWrite(ctx)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.True(t, model.desiredEmptyAfterWrite)
	requireReadModelConcreteIndices(ctx, t, model, "logs-*")

	diags = readAliasIntoModelWithResolution(ctx, "logs", map[string]esTypes.IndexAliases{}, &model, func(context.Context, string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		require.FailNow(t, "resolver should not be called for a desired-empty post-update read")
		return elasticsearch.ResolvedIndexTargets{}, nil
	})
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	requireReadModelConcreteIndices(ctx, t, model, "logs-*")

	model.desiredEmptyAfterWrite = false
	diags = readAliasIntoModelWithResolution(ctx, "logs", map[string]esTypes.IndexAliases{}, &model, func(context.Context, string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		require.FailNow(t, "resolver should not be called for a virtual-state refresh")
		return elasticsearch.ResolvedIndexTargets{}, nil
	})
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	requireReadModelConcreteIndices(ctx, t, model, "logs-*")
}

func TestTfModel_PopulateReadState_RecordsOverlappingCoverageInEachExpression(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := tfModel{
		ReadIndices: mustReadIndexSet(ctx, t,
			readIndexModelWithRouting("logs-*", "shared-route"),
			readIndexModelWithRouting("logs-1", "shared-route"),
		),
	}
	diags := model.populateReadState(ctx, "logs", map[string]esTypes.AliasDefinition{
		"logs-1": {Routing: new("api-route")},
	}, func(_ context.Context, expression string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		require.Contains(t, []string{"logs-*", "logs-1"}, expression)
		return elasticsearch.ResolvedIndexTargets{Names: []string{"logs-1"}}, nil
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	requireReadModelConcreteIndices(ctx, t, model, "logs-*", "logs-1")
	requireReadModelConcreteIndices(ctx, t, model, "logs-1", "logs-1")
	var readIndices []readIndexModel
	diags = model.ReadIndices.ElementsAs(ctx, &readIndices, false)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Len(t, readIndices, 2)
}

func TestTfModel_PopulateReadState_AddsUncoveredMemberAsSingleton(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := tfModel{
		ReadIndices: mustReadIndexSet(ctx, t, readIndexModelWithRouting("logs-configured", "")),
	}
	diags := model.populateReadState(ctx, "logs", map[string]esTypes.AliasDefinition{
		"logs-configured": {},
		"logs-drifted":    {},
	}, func(_ context.Context, expression string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		require.Equal(t, "logs-configured", expression)
		return elasticsearch.ResolvedIndexTargets{Names: []string{"logs-configured"}}, nil
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	requireReadModelConcreteIndices(ctx, t, model, "logs-configured", "logs-configured")
	requireReadModelConcreteIndices(ctx, t, model, "logs-drifted", "logs-drifted")
}

func TestTfModel_PopulateReadState_PreservesConfiguredReadSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	configured := readIndexModel{
		Name:            types.StringValue("logs-*"),
		ConcreteIndices: types.SetNull(types.StringType),
		Filter:          jsontypes.NewNormalizedValue(`{"term":{"environment":"prod"}}`),
		IndexRouting:    types.StringValue("configured-index-route"),
		IsHidden:        types.BoolValue(true),
		Routing:         types.StringValue("configured-route"),
		SearchRouting:   types.StringValue("configured-search-route"),
	}
	model := tfModel{ReadIndices: mustReadIndexSet(ctx, t, configured)}
	diags := model.populateReadState(ctx, "logs", map[string]esTypes.AliasDefinition{
		"logs-1": {
			IndexRouting:  new("api-index-route"),
			IsHidden:      new(false),
			Routing:       new("api-route"),
			SearchRouting: new("api-search-route"),
		},
	}, func(_ context.Context, expression string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		require.Equal(t, "logs-*", expression)
		return elasticsearch.ResolvedIndexTargets{Names: []string{"logs-1"}}, nil
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	var readIndices []readIndexModel
	diags = model.ReadIndices.ElementsAs(ctx, &readIndices, false)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Len(t, readIndices, 1)
	require.Equal(t, configured.Filter, readIndices[0].Filter)
	require.Equal(t, configured.IndexRouting, readIndices[0].IndexRouting)
	require.Equal(t, configured.IsHidden, readIndices[0].IsHidden)
	require.Equal(t, configured.Routing, readIndices[0].Routing)
	require.Equal(t, configured.SearchRouting, readIndices[0].SearchRouting)
	requireReadIndexConcreteIndices(ctx, t, readIndices, "logs-*", "logs-1")
}

func TestReadAliasIntoModelWithResolution_RemovesRealNotFoundDrift(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := tfModel{
		WriteIndex: types.ObjectNull(getIndexAttrTypes(ctx)),
		ReadIndices: mustReadIndexSet(ctx, t, readIndexModel{
			Name:            types.StringValue("logs-*"),
			ConcreteIndices: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("logs-1")}),
			Filter:          jsontypes.NewNormalizedNull(),
			IndexRouting:    types.StringNull(),
			IsHidden:        types.BoolValue(false),
			Routing:         types.StringNull(),
			SearchRouting:   types.StringNull(),
		}),
	}

	diags := readAliasIntoModelWithResolution(ctx, "logs", map[string]esTypes.IndexAliases{}, &model, func(context.Context, string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics) {
		require.FailNow(t, "resolver should not be called for missing alias drift")
		return elasticsearch.ResolvedIndexTargets{}, nil
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.True(t, model.WriteIndex.IsNull())
	require.True(t, model.ReadIndices.IsNull())
}

func virtualReadIndexModel(name string) readIndexModel {
	return readIndexModel{
		Name:            types.StringValue(name),
		ConcreteIndices: types.SetValueMust(types.StringType, nil),
		Filter:          jsontypes.NewNormalizedNull(),
		IndexRouting:    types.StringNull(),
		IsHidden:        types.BoolValue(false),
		Routing:         types.StringNull(),
		SearchRouting:   types.StringNull(),
	}
}

func requireReadIndexConcreteIndices(ctx context.Context, t *testing.T, indices []readIndexModel, name string, concreteIndices ...string) {
	t.Helper()

	for _, index := range indices {
		if index.Name.ValueString() != name {
			continue
		}
		var actual []string
		diags := index.ConcreteIndices.ElementsAs(ctx, &actual, false)
		require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
		require.ElementsMatch(t, concreteIndices, actual)
		return
	}

	require.Failf(t, "read index not found", "name %q", name)
}

func requireReadModelConcreteIndices(ctx context.Context, t *testing.T, model tfModel, name string, concreteIndices ...string) {
	t.Helper()

	var readIndices []readIndexModel
	diags := model.ReadIndices.ElementsAs(ctx, &readIndices, false)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	requireReadIndexConcreteIndices(ctx, t, readIndices, name, concreteIndices...)
}
