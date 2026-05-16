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

package index_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewIndexMappingsValue_stripsImplicitTypeObject verifies that the typed
// go-elasticsearch client's habit of injecting "type":"object" into every
// ObjectProperty is removed at construction time, so post-apply Read results
// match the plan value derived from the user's config.
func TestNewIndexMappingsValue_stripsImplicitTypeObject(t *testing.T) {
	t.Parallel()

	// Simulate what the typed client produces: every nested object property
	// gets an explicit "type":"object" even though the user never wrote it.
	input := `{
		"properties": {
			"nginx": {
				"type": "object",
				"properties": {
					"access": {
						"type": "object",
						"properties": {
							"bytes": {"type": "long"}
						}
					}
				}
			}
		}
	}`

	v := index.NewMappingsValue(input)
	require.False(t, v.IsNull())
	require.False(t, v.IsUnknown())

	var got map[string]any
	require.NoError(t, json.Unmarshal([]byte(v.ValueString()), &got))

	// "nginx" and "access" are object properties — their implicit "type":"object"
	// should be stripped.
	props := got["properties"].(map[string]any)
	nginx := props["nginx"].(map[string]any)
	assert.NotContains(t, nginx, "type", `"type":"object" should be stripped from implicit object property "nginx"`)

	nginxProps := nginx["properties"].(map[string]any)
	access := nginxProps["access"].(map[string]any)
	assert.NotContains(t, access, "type", `"type":"object" should be stripped from implicit object property "access"`)

	// Leaf field with an explicit non-object type must NOT be stripped.
	accessProps := access["properties"].(map[string]any)
	bytes := accessProps["bytes"].(map[string]any)
	assert.Equal(t, "long", bytes["type"], `explicit "type":"long" must be preserved`)
}

// TestNewIndexMappingsValue_preservesExplicitNested verifies that "type":"nested"
// is never stripped — it is semantically distinct from "type":"object".
func TestNewIndexMappingsValue_preservesExplicitNested(t *testing.T) {
	t.Parallel()

	input := `{"properties":{"tags":{"type":"nested","properties":{"label":{"type":"keyword"}}}}}`
	v := index.NewMappingsValue(input)

	var got map[string]any
	require.NoError(t, json.Unmarshal([]byte(v.ValueString()), &got))

	props := got["properties"].(map[string]any)
	tags := props["tags"].(map[string]any)
	assert.Equal(t, "nested", tags["type"], `"type":"nested" must be preserved`)
}

// TestNewIndexMappingsValue_collapsesMatchArrays verifies that single-element
// string arrays for dynamic-template keys are collapsed to plain strings.
func TestNewIndexMappingsValue_collapsesMatchArrays(t *testing.T) {
	t.Parallel()

	input := `{"dynamic_templates":[{"strings_as_keywords":{"match_mapping_type":["string"],"mapping":{"type":"keyword"}}}]}`
	v := index.NewMappingsValue(input)

	var got map[string]any
	require.NoError(t, json.Unmarshal([]byte(v.ValueString()), &got))

	dts := got["dynamic_templates"].([]any)
	rule := dts[0].(map[string]any)["strings_as_keywords"].(map[string]any)
	assert.Equal(t, "string", rule["match_mapping_type"], "single-element array should be collapsed to plain string")
}

// TestMappingsSemanticallyEqual_scalarVsStringifiedScalar verifies that scalar
// leaf values (bool, number) compare semantically equal to their stringified
// equivalents that Elasticsearch may echo back.
func TestMappingsSemanticallyEqual_scalarVsStringifiedScalar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		userJSON string
		apiJSON  string
		want     bool
	}{
		{
			name:     "bool true vs string true",
			userJSON: `{"index":true}`,
			apiJSON:  `{"index":"true"}`,
			want:     true,
		},
		{
			name:     "bool false vs string false",
			userJSON: `{"index":false}`,
			apiJSON:  `{"index":"false"}`,
			want:     true,
		},
		{
			name:     "number vs string number",
			userJSON: `{"boost":42}`,
			apiJSON:  `{"boost":"42"}`,
			want:     true,
		},
		{
			name:     "bool true vs string false - not equal",
			userJSON: `{"index":true}`,
			apiJSON:  `{"index":"false"}`,
			want:     false,
		},
		{
			name:     "number 1 vs string 2 - not equal",
			userJSON: `{"boost":1}`,
			apiJSON:  `{"boost":"2"}`,
			want:     false,
		},
		{
			name:     "two distinct strings - not equal",
			userJSON: `{"format":"strict_date"}`,
			apiJSON:  `{"format":"epoch_millis"}`,
			want:     false,
		},
		{
			name:     "large number vs string - no scientific notation mismatch",
			userJSON: `{"ignore_above":7000000}`,
			apiJSON:  `{"ignore_above":"7000000"}`,
			want:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var userMap, apiMap map[string]any
			require.NoError(t, json.Unmarshal([]byte(tc.userJSON), &userMap))
			require.NoError(t, json.Unmarshal([]byte(tc.apiJSON), &apiMap))

			got := index.MappingsSemanticallyEqual(userMap, apiMap)
			assert.Equal(t, tc.want, got, "MappingsSemanticallyEqual(%s, %s)", tc.userJSON, tc.apiJSON)
		})
	}
}

// TestMappingsSemanticallyEqual_scalarLeafInFieldDef verifies that scalar drift
// inside nested field definitions (e.g. inside "properties") is also handled.
func TestMappingsSemanticallyEqual_scalarLeafInFieldDef(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		userJSON string
		apiJSON  string
		want     bool
	}{
		{
			name:     "bool leaf in field def - equal",
			userJSON: `{"properties":{"myfield":{"type":"keyword","index":true}}}`,
			apiJSON:  `{"properties":{"myfield":{"type":"keyword","index":"true"}}}`,
			want:     true,
		},
		{
			name:     "numeric leaf in field def - equal",
			userJSON: `{"properties":{"myfield":{"type":"float","boost":1.5}}}`,
			apiJSON:  `{"properties":{"myfield":{"type":"float","boost":"1.5"}}}`,
			want:     true,
		},
		{
			name:     "bool leaf in field def - not equal",
			userJSON: `{"properties":{"myfield":{"type":"keyword","index":true}}}`,
			apiJSON:  `{"properties":{"myfield":{"type":"keyword","index":"false"}}}`,
			want:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var userMap, apiMap map[string]any
			require.NoError(t, json.Unmarshal([]byte(tc.userJSON), &userMap))
			require.NoError(t, json.Unmarshal([]byte(tc.apiJSON), &apiMap))

			got := index.MappingsSemanticallyEqual(userMap, apiMap)
			assert.Equal(t, tc.want, got, "MappingsSemanticallyEqual(%s, %s)", tc.userJSON, tc.apiJSON)
		})
	}
}

// TestIndexMappingsValue_SemanticEquals verifies that the typed-client-injected
// "type":"object" entries don't cause spurious drift between a config-derived
// plan value (no "type":"object") and the API-read state value (with "type":"object").
func TestIndexMappingsValue_SemanticEquals(t *testing.T) {
	t.Parallel()

	// Config / plan value: no "type":"object"
	planJSON := `{"properties":{"nginx":{"properties":{"access":{"properties":{"bytes":{"type":"long"}}}}}}}`
	// API read value: typed client adds "type":"object"
	apiJSON := `{"properties":{"nginx":{"type":"object","properties":{"access":{"type":"object","properties":{"bytes":{"type":"long"}}}}}}}`

	plan := index.NewMappingsValue(planJSON)
	api := index.NewMappingsValue(apiJSON)

	// After normalization both should be equal strings (type:object stripped from api).
	assert.Equal(t, plan.ValueString(), api.ValueString(), "normalized forms should be identical")

	// StringSemanticEquals should also return true.
	eq, diags := plan.StringSemanticEquals(context.Background(), api)
	require.False(t, diags.HasError())
	assert.True(t, eq)
}
