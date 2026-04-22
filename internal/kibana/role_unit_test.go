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

package kibana

import (
	"encoding/json"
	"testing"

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// expandESForTest calls expandKibanaRoleElasticsearchInto and returns the SecurityRoleES for testing.
func expandESForTest(t *testing.T, esSet *schema.Set, ver *version.Version) kibanaoapi.SecurityRoleES {
	t.Helper()
	var es kibanaoapi.SecurityRoleES
	diags := expandKibanaRoleElasticsearchInto(esSet, ver, &es)
	require.Nil(t, diags)
	return es
}

func TestRoleIndexFieldSecurityRoundTrip(t *testing.T) {
	grantSet := schema.NewSet(schema.HashString, []any{"field1", "field2"})
	exceptSet := schema.NewSet(schema.HashString, []any{"secret"})
	namesSet := schema.NewSet(schema.HashString, []any{"logs-*"})
	privsSet := schema.NewSet(schema.HashString, []any{"read", "write"})

	fieldSecList := []any{
		map[string]any{
			"grant":  grantSet,
			"except": exceptSet,
		},
	}

	indexEntry := map[string]any{
		"names":          namesSet,
		"privileges":     privsSet,
		"query":          `{"match_all":{}}`,
		"field_security": fieldSecList,
	}

	indexSet := schema.NewSet(schema.HashResource(&schema.Resource{
		Schema: map[string]*schema.Schema{},
	}), []any{indexEntry})

	indices := expandIndices(indexSet)
	require.Len(t, indices, 1)

	idx := indices[0]
	require.NotNil(t, idx.FieldSecurity)
	assert.ElementsMatch(t, []string{"field1", "field2"}, (*idx.FieldSecurity)["grant"])
	assert.ElementsMatch(t, []string{"secret"}, (*idx.FieldSecurity)["except"])
	require.NotNil(t, idx.Query)
	assert.JSONEq(t, `{"match_all":{}}`, *idx.Query)

	// Round-trip through JSON into SecurityRoleESIndex
	idxJSON, err := json.Marshal(indices)
	require.NoError(t, err)

	var decoded []kibanaoapi.SecurityRoleESIndex
	require.NoError(t, json.Unmarshal(idxJSON, &decoded))

	flatResult := flattenKibanaRoleIndicesData(&decoded)
	require.Len(t, flatResult, 1)

	flatIdx := flatResult[0].(map[string]any)
	assert.ElementsMatch(t, []string{"logs-*"}, flatIdx["names"])
	assert.ElementsMatch(t, []string{"read", "write"}, flatIdx["privileges"])
	assert.JSONEq(t, `{"match_all":{}}`, flatIdx["query"].(string))

	fsec := flatIdx["field_security"].([]any)[0].(map[string]any)
	assert.ElementsMatch(t, []string{"field1", "field2"}, fsec["grant"])
	assert.ElementsMatch(t, []string{"secret"}, fsec["except"])
}

func TestRoleRemoteIndicesRoundTrip(t *testing.T) {
	namesSet := schema.NewSet(schema.HashString, []any{"sample"})
	clustersSet := schema.NewSet(schema.HashString, []any{"test-cluster"})
	privsSet := schema.NewSet(schema.HashString, []any{"create", "read", "write"})
	grantSet := schema.NewSet(schema.HashString, []any{"sample"})

	fieldSecList := []any{
		map[string]any{
			"grant":  grantSet,
			"except": schema.NewSet(schema.HashString, []any{}),
		},
	}

	remoteEntry := map[string]any{
		"names":          namesSet,
		"clusters":       clustersSet,
		"privileges":     privsSet,
		"query":          "",
		"field_security": fieldSecList,
	}

	remoteSet := schema.NewSet(schema.HashResource(&schema.Resource{Schema: map[string]*schema.Schema{}}), []any{remoteEntry})

	remoteIndices := expandRemoteIndices(remoteSet)
	require.Len(t, remoteIndices, 1)

	ri := remoteIndices[0]
	assert.ElementsMatch(t, []string{"sample"}, ri.Names)
	assert.ElementsMatch(t, []string{"test-cluster"}, ri.Clusters)
	assert.ElementsMatch(t, []string{"create", "read", "write"}, ri.Privileges)
	require.NotNil(t, ri.FieldSecurity)
	assert.ElementsMatch(t, []string{"sample"}, (*ri.FieldSecurity)["grant"])

	// Round-trip through JSON
	riJSON, err := json.Marshal(remoteIndices)
	require.NoError(t, err)

	var decoded []kibanaoapi.SecurityRoleESRemoteIndex
	require.NoError(t, json.Unmarshal(riJSON, &decoded))

	flatResult := flattenKibanaRoleRemoteIndicesData(&decoded)
	require.Len(t, flatResult, 1)

	flatRI := flatResult[0].(map[string]any)
	assert.ElementsMatch(t, []string{"sample"}, flatRI["names"])
	assert.ElementsMatch(t, []string{"test-cluster"}, flatRI["clusters"])
	assert.ElementsMatch(t, []string{"create", "read", "write"}, flatRI["privileges"])

	fsec := flatRI["field_security"].([]any)[0].(map[string]any)
	assert.ElementsMatch(t, []string{"sample"}, fsec["grant"])
}

func TestRoleKibanaBaseRoundTrip(t *testing.T) {
	spacesSet := schema.NewSet(schema.HashString, []any{"default"})
	baseSet := schema.NewSet(schema.HashString, []any{"all"})

	kibanaEntry := map[string]any{
		"base":    baseSet,
		"feature": schema.NewSet(schema.HashString, []any{}),
		"spaces":  spacesSet,
	}

	kibanaSet := schema.NewSet(schema.HashResource(&schema.Resource{Schema: map[string]*schema.Schema{}}), []any{kibanaEntry})

	kibanaPrivs, diags := expandKibanaRoleKibana(kibanaSet)
	require.Nil(t, diags)
	require.Len(t, kibanaPrivs, 1)

	// Verify base is correctly encoded as JSON array
	var base []string
	require.NoError(t, json.Unmarshal(kibanaPrivs[0].Base, &base))
	assert.ElementsMatch(t, []string{"all"}, base)

	// Round-trip flatten
	flatResult := flattenKibanaRoleKibanaData(kibanaPrivs)
	require.Len(t, flatResult, 1)

	flat := flatResult[0].(map[string]any)
	assert.ElementsMatch(t, []string{"all"}, flat["base"])
	assert.ElementsMatch(t, []string{"default"}, flat["spaces"])
}

func TestRoleKibanaFeatureRoundTrip(t *testing.T) {
	spacesSet := schema.NewSet(schema.HashString, []any{"default"})

	featurePrivsSet := schema.NewSet(schema.HashString, []any{"minimal_read", "url_create"})
	featureEntry := map[string]any{
		"name":       "discover",
		"privileges": featurePrivsSet,
	}
	featureSet := schema.NewSet(schema.HashResource(&schema.Resource{Schema: map[string]*schema.Schema{}}), []any{featureEntry})

	kibanaEntry := map[string]any{
		"base":    schema.NewSet(schema.HashString, []any{}),
		"feature": featureSet,
		"spaces":  spacesSet,
	}

	kibanaSet := schema.NewSet(schema.HashResource(&schema.Resource{Schema: map[string]*schema.Schema{}}), []any{kibanaEntry})

	kibanaPrivs, diags := expandKibanaRoleKibana(kibanaSet)
	require.Nil(t, diags)
	require.Len(t, kibanaPrivs, 1)

	// Feature map should have "discover"
	require.NotNil(t, kibanaPrivs[0].Feature)
	featurePrivs, ok := (*kibanaPrivs[0].Feature)["discover"]
	require.True(t, ok)
	assert.ElementsMatch(t, []string{"minimal_read", "url_create"}, featurePrivs)

	// Flatten
	flatResult := flattenKibanaRoleKibanaData(kibanaPrivs)
	require.Len(t, flatResult, 1)

	flat := flatResult[0].(map[string]any)
	assert.Empty(t, flat["base"])

	featureFlat := flattenKibanaRoleKibanaFeatureData(kibanaPrivs[0].Feature)
	assert.Len(t, featureFlat, 1)
	featureMap := featureFlat[0].(map[string]any)
	assert.Equal(t, "discover", featureMap["name"])
	assert.ElementsMatch(t, []string{"minimal_read", "url_create"}, featureMap["privileges"])
}

func TestRoleMetadataRoundTrip(t *testing.T) {
	metaJSON := `{"team":"ops","env":"prod"}`

	metadata, diags := expandKibanaRoleMetadata(metaJSON)
	require.Nil(t, diags)
	require.NotNil(t, metadata)
	assert.Equal(t, "ops", metadata["team"])

	// Simulate what flatten does: marshal back
	out, err := json.Marshal(metadata)
	require.NoError(t, err)
	var roundTripped map[string]any
	require.NoError(t, json.Unmarshal(out, &roundTripped))
	assert.Equal(t, "ops", roundTripped["team"])
	assert.Equal(t, "prod", roundTripped["env"])
}

func TestExpandESConfig_EmptyClusterAndRunAs(t *testing.T) {
	// Verify that empty cluster and run_as are omitted (not set as non-nil pointers)
	clusterSet := schema.NewSet(schema.HashString, []any{})
	runAsSet := schema.NewSet(schema.HashString, []any{})
	namesSet := schema.NewSet(schema.HashString, []any{"my-index"})
	privsSet := schema.NewSet(schema.HashString, []any{"read"})

	indexEntry := map[string]any{
		"names":          namesSet,
		"privileges":     privsSet,
		"query":          "",
		"field_security": []any{},
	}
	indexSet := schema.NewSet(schema.HashResource(&schema.Resource{Schema: map[string]*schema.Schema{}}), []any{indexEntry})

	esMap := map[string]any{
		"cluster":        clusterSet,
		"indices":        indexSet,
		"remote_indices": schema.NewSet(schema.HashString, []any{}),
		"run_as":         runAsSet,
	}
	esSet := schema.NewSet(schema.HashResource(&schema.Resource{Schema: map[string]*schema.Schema{}}), []any{esMap})

	ver := version.Must(version.NewVersion("9.0.0"))
	es := expandESForTest(t, esSet, ver)

	assert.Nil(t, es.Cluster, "empty cluster should be nil")
	assert.Nil(t, es.RunAs, "empty run_as should be nil")
	assert.NotNil(t, es.Indices)
}
