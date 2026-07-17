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

package security_role

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func nullEsHint() types.Object {
	return types.ObjectNull(elasticsearchResourceAttrTypes())
}

func nullKibanaHint() types.Set {
	return types.SetNull(kibanaBlockObjectType())
}

func TestUnitFlattenExpandIndexFieldSecurityRoundTrip(t *testing.T) {
	ctx := context.Background()
	grant := []string{"field1", "field2"}
	except := []string{"secret"}
	fs := map[string][]string{"grant": grant, "except": except}
	q := `{"match_all":{}}`
	es := kibanaoapi.SecurityRoleES{
		Indices: &[]kibanaoapi.SecurityRoleESIndex{{
			Names:         []string{"logs-*"},
			Privileges:    []string{"read", "write"},
			Query:         &q,
			FieldSecurity: &fs,
		}},
	}
	obj, diags := flattenElasticsearchObject(ctx, &es, nullEsHint())
	require.False(t, diags.HasError())
	out, diags2 := expandElasticsearch(ctx, obj)
	require.False(t, diags2.HasError())
	b1, err := json.Marshal(es)
	require.NoError(t, err)
	b2, err := json.Marshal(out)
	require.NoError(t, err)
	assert.JSONEq(t, string(b1), string(b2))
}

func TestUnitFlattenRemoteIndicesAPIFalseWithNullHint(t *testing.T) {
	ctx := context.Background()
	apiFalse := false
	es := kibanaoapi.SecurityRoleES{
		RemoteIndices: &[]kibanaoapi.SecurityRoleESRemoteIndex{{
			Names:                  []string{"sample"},
			Clusters:               []string{"test-cluster"},
			Privileges:             []string{"read"},
			AllowRestrictedIndices: &apiFalse,
		}},
	}
	obj, diags := flattenElasticsearchObject(ctx, &es, nullEsHint())
	require.False(t, diags.HasError())
	remoteSet := obj.Attributes()[attrRemoteIndices].(types.Set)
	require.Equal(t, types.BoolValue(false), remoteSet.Elements()[0].(types.Object).Attributes()[attrAllowRestrictedIndices])
}

func TestUnitFlattenRemoteIndicesHintNullPreservesNullDespiteAPIFalse(t *testing.T) {
	ctx := context.Background()
	apiFalse := false
	es := kibanaoapi.SecurityRoleES{
		RemoteIndices: &[]kibanaoapi.SecurityRoleESRemoteIndex{{
			Names:                  []string{"sample"},
			Clusters:               []string{"test-cluster"},
			Privileges:             []string{"read"},
			AllowRestrictedIndices: &apiFalse,
		}},
	}
	clustersSet := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test-cluster")})
	namesSet := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("sample")})
	privSet := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("read")})
	hintEntry, d := types.ObjectValue(esRemoteIndexResourceAttrTypes(), map[string]attr.Value{
		attrAllowRestrictedIndices: types.BoolNull(),
		attrClusters:               clustersSet,
		attrNames:                  namesSet,
		attrPrivileges:             privSet,
		attrQuery:                  jsontypes.NewNormalizedNull(),
		attrFieldSecurity:          types.ObjectNull(fieldSecurityAttrTypes()),
	})
	require.False(t, d.HasError())
	hint, d := types.ObjectValue(elasticsearchResourceAttrTypes(), map[string]attr.Value{
		attrCluster:       types.SetNull(types.StringType),
		attrRunAs:         types.SetNull(types.StringType),
		attrIndices:       types.SetNull(types.ObjectType{AttrTypes: esIndexResourceAttrTypes()}),
		attrRemoteIndices: types.SetValueMust(types.ObjectType{AttrTypes: esRemoteIndexResourceAttrTypes()}, []attr.Value{hintEntry}),
	})
	require.False(t, d.HasError())
	obj, diags := flattenElasticsearchObject(ctx, &es, hint)
	require.False(t, diags.HasError())
	remoteSet := obj.Attributes()[attrRemoteIndices].(types.Set)
	require.Equal(t, types.BoolNull(), remoteSet.Elements()[0].(types.Object).Attributes()[attrAllowRestrictedIndices])
}

func TestUnitFlattenExpandRemoteIndicesRoundTrip(t *testing.T) {
	ctx := context.Background()
	grant := []string{"sample"}
	fs := map[string][]string{"grant": grant}
	allowRestricted := true
	es := kibanaoapi.SecurityRoleES{
		RemoteIndices: &[]kibanaoapi.SecurityRoleESRemoteIndex{{
			Names:                  []string{"sample"},
			Clusters:               []string{"test-cluster"},
			Privileges:             []string{"create", "read", "write"},
			AllowRestrictedIndices: &allowRestricted,
			FieldSecurity:          &fs,
		}},
	}
	obj, diags := flattenElasticsearchObject(ctx, &es, nullEsHint())
	require.False(t, diags.HasError())
	out, diags2 := expandElasticsearch(ctx, obj)
	require.False(t, diags2.HasError())
	b1, err := json.Marshal(es)
	require.NoError(t, err)
	b2, err := json.Marshal(out)
	require.NoError(t, err)
	assert.JSONEq(t, string(b1), string(b2))
}

func TestUnitFlattenExpandKibanaBaseRoundTrip(t *testing.T) {
	ctx := context.Background()
	spaces := []string{"default"}
	kcfg := []kibanaoapi.SecurityRoleKibana{
		{
			Base:    mustMarshalJSON(t, []string{"all"}),
			Feature: nil,
			Spaces:  &spaces,
		},
	}
	set, diags := flattenKibana(ctx, kcfg, nullKibanaHint())
	require.False(t, diags.HasError())
	out, diags2 := expandKibana(ctx, set)
	require.False(t, diags2.HasError())
	b1, err := json.Marshal(kcfg)
	require.NoError(t, err)
	b2, err := json.Marshal(out)
	require.NoError(t, err)
	assert.JSONEq(t, string(b1), string(b2))
}

func TestUnitFlattenExpandKibanaFeatureRoundTrip(t *testing.T) {
	ctx := context.Background()
	fm := map[string][]string{
		"discover": {"minimal_read", "url_create"},
	}
	spaces := []string{"default"}
	kcfg := []kibanaoapi.SecurityRoleKibana{
		{
			Feature: &fm,
			Spaces:  &spaces,
		},
	}
	set, diags := flattenKibana(ctx, kcfg, nullKibanaHint())
	require.False(t, diags.HasError())
	out, diags2 := expandKibana(ctx, set)
	require.False(t, diags2.HasError())
	b1, err := json.Marshal(kcfg)
	require.NoError(t, err)
	b2, err := json.Marshal(out)
	require.NoError(t, err)
	assert.JSONEq(t, string(b1), string(b2))
}

func TestUnitExpandKibanaRejectsEmptyPrivileges(t *testing.T) {
	ctx := context.Background()
	featureElemType := types.ObjectType{AttrTypes: kibanaFeatureAttrTypes()}
	spacesSet := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("*")})
	emptyFeatureSet := types.SetValueMust(featureElemType, []attr.Value{})
	emptyBaseSet := types.SetValueMust(types.StringType, []attr.Value{})

	kibanaObj := types.ObjectValueMust(kibanaBlockAttrTypes(), map[string]attr.Value{
		attrSpaces:  spacesSet,
		attrBase:    emptyBaseSet,
		attrFeature: emptyFeatureSet,
	})
	kibanaSet := types.SetValueMust(kibanaBlockObjectType(), []attr.Value{kibanaObj})

	_, diags := expandKibana(ctx, kibanaSet)
	require.True(t, diags.HasError(), "expected apply-time privilege validation error")
	found := false
	for _, d := range diags.Errors() {
		if strings.Contains(d.Detail(), "Either one of the `feature` or `base` privileges must be set for kibana role!") {
			found = true
			break
		}
	}
	assert.True(t, found, "expected missing-privilege diagnostic from expandKibana")
}

func TestUnitExpandElasticsearchOmitsEmptyClusterAndRunAs(t *testing.T) {
	ctx := context.Background()
	es := kibanaoapi.SecurityRoleES{
		Indices: &[]kibanaoapi.SecurityRoleESIndex{{
			Names:         []string{"my-index"},
			Privileges:    []string{"read"},
			FieldSecurity: nil,
		}},
	}
	obj, diags := flattenElasticsearchObject(ctx, &es, nullEsHint())
	require.False(t, diags.HasError())
	out, diags2 := expandElasticsearch(ctx, obj)
	require.False(t, diags2.HasError())
	assert.Nil(t, out.Cluster)
	assert.Nil(t, out.RunAs)
	require.NotNil(t, out.Indices)
}

func TestUnitMetadataFromAPI(t *testing.T) {
	t.Run("nil metadata returns null", func(t *testing.T) {
		role := kibanaoapi.SecurityRole{}
		norm, diags := metadataFromAPI(&role)
		require.False(t, diags.HasError())
		assert.True(t, norm.IsNull())
	})

	t.Run("valid metadata returns normalized JSON", func(t *testing.T) {
		m := map[string]any{"team": "ops", "env": "prod"}
		role := kibanaoapi.SecurityRole{Metadata: &m}
		norm, diags := metadataFromAPI(&role)
		require.False(t, diags.HasError())
		assert.JSONEq(t, `{"team":"ops","env":"prod"}`, norm.ValueString())
	})
}

func TestUnitFlattenKibanaInvalidBaseReturnsError(t *testing.T) {
	ctx := context.Background()
	spaces := []string{"default"}
	kcfg := []kibanaoapi.SecurityRoleKibana{
		{
			Base:   []byte(`{"unexpected":"shape"}`),
			Spaces: &spaces,
		},
	}
	_, diags := flattenKibana(ctx, kcfg, nullKibanaHint())
	require.True(t, diags.HasError(), "expected diagnostic for malformed kibana.base payload")
}

// TestUnitFlattenKibanaBasePreservesHintRepresentation pins the
// representation chosen for `base` when the API omits it. A user that wrote
// `base = []` must round-trip to an empty set; a user that omitted `base`
// entirely must round-trip to null. The kibana entries are matched against
// the hint by their `spaces` attribute.
func TestUnitFlattenKibanaBasePreservesHintRepresentation(t *testing.T) {
	ctx := context.Background()
	spacesA := []string{"space-a"}
	spacesB := []string{"space-b"}
	feat := map[string][]string{"fleet": {"all"}}
	kcfg := []kibanaoapi.SecurityRoleKibana{
		{Feature: &feat, Spaces: &spacesA},
		{Feature: &feat, Spaces: &spacesB},
	}

	emptyBase := types.SetValueMust(types.StringType, []attr.Value{})
	nullBase := types.SetNull(types.StringType)
	emptyFeature := types.SetNull(types.ObjectType{AttrTypes: kibanaFeatureAttrTypes()})

	spacesASet, d := types.SetValueFrom(ctx, types.StringType, spacesA)
	require.False(t, d.HasError())
	spacesBSet, d := types.SetValueFrom(ctx, types.StringType, spacesB)
	require.False(t, d.HasError())

	entryA, d := types.ObjectValue(kibanaBlockAttrTypes(), map[string]attr.Value{
		"spaces":  spacesASet,
		"base":    emptyBase,
		"feature": emptyFeature,
	})
	require.False(t, d.HasError())
	entryB, d := types.ObjectValue(kibanaBlockAttrTypes(), map[string]attr.Value{
		"spaces":  spacesBSet,
		"base":    nullBase,
		"feature": emptyFeature,
	})
	require.False(t, d.HasError())
	hint, d := types.SetValue(kibanaBlockObjectType(), []attr.Value{entryA, entryB})
	require.False(t, d.HasError())

	set, diags := flattenKibana(ctx, kcfg, hint)
	require.False(t, diags.HasError())

	got := map[string]types.Set{}
	for _, el := range set.Elements() {
		obj := el.(types.Object)
		sp := obj.Attributes()["spaces"].(types.Set)
		got[setStringKey(sp)] = obj.Attributes()["base"].(types.Set)
	}
	assert.False(t, got[setStringKey(spacesASet)].IsNull(), "base for space-a should preserve empty representation")
	assert.Empty(t, got[setStringKey(spacesASet)].Elements())
	assert.True(t, got[setStringKey(spacesBSet)].IsNull(), "base for space-b should preserve null representation")
}

// TestUnitFlattenElasticsearchClusterRunAsPreservesHintRepresentation pins
// representation choice for the optional top-level `cluster` and `run_as`
// sets when the API omits them.
func TestUnitFlattenElasticsearchClusterRunAsPreservesHintRepresentation(t *testing.T) {
	ctx := context.Background()
	es := kibanaoapi.SecurityRoleES{}
	emptySet := types.SetValueMust(types.StringType, []attr.Value{})
	nullSet := types.SetNull(types.StringType)
	nullIndices := types.SetNull(types.ObjectType{AttrTypes: esIndexResourceAttrTypes()})
	nullRemote := types.SetNull(types.ObjectType{AttrTypes: esRemoteIndexResourceAttrTypes()})

	hint, d := types.ObjectValue(elasticsearchResourceAttrTypes(), map[string]attr.Value{
		"cluster":        emptySet,
		"run_as":         nullSet,
		"indices":        nullIndices,
		"remote_indices": nullRemote,
	})
	require.False(t, d.HasError())

	obj, diags := flattenElasticsearchObject(ctx, &es, hint)
	require.False(t, diags.HasError())
	cluster := obj.Attributes()["cluster"].(types.Set)
	runAs := obj.Attributes()["run_as"].(types.Set)
	assert.False(t, cluster.IsNull(), "cluster should preserve empty hint")
	assert.Empty(t, cluster.Elements())
	assert.True(t, runAs.IsNull(), "run_as should preserve null hint")
}

func mustMarshalJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}
