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
	"testing"

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	set, diags := flattenElasticsearch(ctx, &es)
	require.False(t, diags.HasError())
	out, diags2 := expandElasticsearch(ctx, set)
	require.False(t, diags2.HasError())
	b1, err := json.Marshal(es)
	require.NoError(t, err)
	b2, err := json.Marshal(out)
	require.NoError(t, err)
	assert.JSONEq(t, string(b1), string(b2))
}

func TestUnitFlattenExpandRemoteIndicesRoundTrip(t *testing.T) {
	ctx := context.Background()
	grant := []string{"sample"}
	fs := map[string][]string{"grant": grant}
	es := kibanaoapi.SecurityRoleES{
		RemoteIndices: &[]kibanaoapi.SecurityRoleESRemoteIndex{{
			Names:         []string{"sample"},
			Clusters:      []string{"test-cluster"},
			Privileges:    []string{"create", "read", "write"},
			FieldSecurity: &fs,
		}},
	}
	set, diags := flattenElasticsearch(ctx, &es)
	require.False(t, diags.HasError())
	out, diags2 := expandElasticsearch(ctx, set)
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
	set, diags := flattenKibana(ctx, kcfg)
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
	set, diags := flattenKibana(ctx, kcfg)
	require.False(t, diags.HasError())
	out, diags2 := expandKibana(ctx, set)
	require.False(t, diags2.HasError())
	b1, err := json.Marshal(kcfg)
	require.NoError(t, err)
	b2, err := json.Marshal(out)
	require.NoError(t, err)
	assert.JSONEq(t, string(b1), string(b2))
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
	set, diags := flattenElasticsearch(ctx, &es)
	require.False(t, diags.HasError())
	out, diags2 := expandElasticsearch(ctx, set)
	require.False(t, diags2.HasError())
	assert.Nil(t, out.Cluster)
	assert.Nil(t, out.RunAs)
	require.NotNil(t, out.Indices)
}

func TestUnitMetadataJSONRoundTrip(t *testing.T) {
	meta := jsontypes.NewNormalizedValue(`{"team":"ops","env":"prod"}`)
	ptr, diags := expandMetadata(meta)
	require.False(t, diags.HasError())
	require.NotNil(t, ptr)
	role := kibanaoapi.SecurityRole{Metadata: ptr}
	norm, d2 := metadataFromAPI(&role)
	require.False(t, d2.HasError())
	assert.JSONEq(t, meta.ValueString(), norm.ValueString())
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
	_, diags := flattenKibana(ctx, kcfg)
	require.True(t, diags.HasError(), "expected diagnostic for malformed kibana.base payload")
}

func mustMarshalJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}
