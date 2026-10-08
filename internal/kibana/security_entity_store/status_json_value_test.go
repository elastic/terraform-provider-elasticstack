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

package security_entity_store

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusJSONType_ValueFromTerraform(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	nullVal, err := StatusJSONType{}.ValueFromTerraform(ctx, tftypes.NewValue(tftypes.String, nil))
	require.NoError(t, err)
	assert.True(t, nullVal.IsNull(), "null input must produce a null StatusJSONValue, got %T", nullVal)

	unknownVal, err := StatusJSONType{}.ValueFromTerraform(ctx, tftypes.NewValue(tftypes.String, tftypes.UnknownValue))
	require.NoError(t, err)
	assert.True(t, unknownVal.IsUnknown(), "unknown input must produce an unknown StatusJSONValue, got %T", unknownVal)

	knownVal, err := StatusJSONType{}.ValueFromTerraform(ctx, tftypes.NewValue(tftypes.String, `{"status":"running"}`))
	require.NoError(t, err)
	assert.IsType(t, StatusJSONValue{}, knownVal)
}

func TestStatusJSONValue_ValueFromString(t *testing.T) {
	t.Parallel()

	val, diags := StatusJSONType{}.ValueFromString(context.Background(), basetypes.NewStringValue(`{"status":"running"}`))
	require.False(t, diags.HasError(), "%v", diags)
	require.IsType(t, StatusJSONValue{}, val)
	assert.JSONEq(t, `{"status":"running"}`, val.(StatusJSONValue).ValueString())
	assert.False(t, val.(StatusJSONValue).IsNull())
	assert.False(t, val.(StatusJSONValue).IsUnknown())
}

func TestStatusJSONValue_RawConstructorUntouched(t *testing.T) {
	t.Parallel()

	// The raw API JSON is stored byte-identical, including unusual
	// whitespace, key order, and escaped characters.
	raw := `{"engines":[ {"type":"use\u0072"} ],  "status":"running"}`
	assert.Equal(t, raw, NewStatusJSONValue(raw).ValueString())
	assert.True(t, NewStatusJSONNull().IsNull())
	assert.True(t, NewStatusJSONUnknown().IsUnknown())
}

func TestStatusJSONValue_NullUnknownSemantics(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	known := NewStatusJSONValue(`{"status":"running"}`)

	tests := []struct {
		name  string
		a     StatusJSONValue
		b     StatusJSONValue
		equal bool
	}{
		{"null vs null", NewStatusJSONNull(), NewStatusJSONNull(), true},
		{"null vs known", NewStatusJSONNull(), known, false},
		{"known vs null", known, NewStatusJSONNull(), false},
		{"unknown vs unknown", NewStatusJSONUnknown(), NewStatusJSONUnknown(), true},
		{"unknown vs null", NewStatusJSONUnknown(), NewStatusJSONNull(), false},
		{"known vs unknown", known, NewStatusJSONUnknown(), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			eq, diags := tc.a.StringSemanticEquals(ctx, tc.b)
			require.False(t, diags.HasError(), "%v", diags)
			assert.Equal(t, tc.equal, eq)
		})
	}
}

func TestStatusJSONValue_EngineOrderPermutations(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	engines := map[string]string{
		"generic": `{"type":"generic","indexPattern":".entities-generic-v1","status":"running"}`,
		"host":    `{"type":"host","indexPattern":".entities-host-v1","status":"running"}`,
		"user":    `{"type":"user","indexPattern":".entities-user-v1","status":"running"}`,
	}
	permutations := []string{
		"generic host user", "generic user host", "host generic user",
		"host user generic", "user generic host", "user host generic",
	}

	first := NewStatusJSONNull()
	for _, perm := range permutations {
		order := strings.Split(perm, " ")
		rawEngines := make([]string, len(order))
		for i, engineType := range order {
			rawEngines[i] = engines[engineType]
		}
		body := `{"status":"running","engines":[` + strings.Join(rawEngines, ",") + `]}`
		v := NewStatusJSONValue(body)
		if first.IsNull() {
			first = v
			continue
		}
		eq, diags := first.StringSemanticEquals(ctx, v)
		require.False(t, diags.HasError(), "%v", diags)
		assert.True(t, eq, "permutation %q must be semantically equal to the first permutation", perm)
	}
}

func TestStatusJSONValue_LogicalDifferencesNotEqual(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	base := `{"status":"running","engines":[{"type":"user","indexPattern":".entities-user-v1","status":"running"}]}`
	tests := []struct {
		name string
		body string
	}{
		{
			name: "engine status differs",
			body: `{"status":"running","engines":[{"type":"user","indexPattern":".entities-user-v1","status":"stopped"}]}`,
		},
		{
			name: "overall status differs",
			body: `{"status":"stopped","engines":[{"type":"user","indexPattern":".entities-user-v1","status":"running"}]}`,
		},
		{
			name: "engine type differs",
			body: `{"status":"running","engines":[{"type":"host","indexPattern":".entities-user-v1","status":"running"}]}`,
		},
		{
			name: "engine index pattern differs",
			body: `{"status":"running","engines":[{"type":"user","indexPattern":".entities-user-v2","status":"running"}]}`,
		},
		{
			name: "extra engine",
			body: `{"status":"running","engines":[{"type":"user","indexPattern":".entities-user-v1","status":"running"},{"type":"host","indexPattern":".entities-host-v1","status":"running"}]}`,
		},
	}
	a := NewStatusJSONValue(base)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			eq, diags := a.StringSemanticEquals(ctx, NewStatusJSONValue(tc.body))
			require.False(t, diags.HasError(), "%v", diags)
			assert.False(t, eq, "body %s must not be semantically equal to the base", tc.body)
		})
	}
}

func TestStatusJSONValue_JSONFormatEquivalence(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name  string
		a     string
		b     string
		equal bool
	}{
		{
			name:  "whitespace and key order equivalent",
			a:     `{"engines":[{"type":"user"}],"status":"running"}`,
			b:     "{\n \"status\":\"running\",\n \"engines\":[{\"type\": \"user\"}]\n}",
			equal: true,
		},
		{
			name:  "unicode escape equivalent",
			a:     `{"engines":[{"type":"use\u0072"}],"status":"running"}`,
			b:     `{"engines":[{"type":"user"}],"status":"running"}`,
			equal: true,
		},
		{
			name:  "non-engine array order remains significant",
			a:     `{"notes":["a","b"],"status":"running"}`,
			b:     `{"notes":["b","a"],"status":"running"}`,
			equal: false,
		},
		{
			// The library decodes with encoding/json UseNumber, so 1e2 and
			// 100 keep distinct literal representations and do NOT compare
			// equal. Pins existing conservative behavior, no change made.
			name:  "number literal representation remains significant (library UseNumber)",
			a:     `{"status":"running","engines":[{"type":"user"}],"n":1e2}`,
			b:     `{"status":"running","engines":[{"type":"user"}],"n":100}`,
			equal: false,
		},
		{
			name:  "non-engine nested array order remains significant",
			a:     `{"engines":[],"extra":{"list":["a","b"]}}`,
			b:     `{"engines":[],"extra":{"list":["b","a"]}}`,
			equal: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			eq, diags := NewStatusJSONValue(tc.a).StringSemanticEquals(ctx, NewStatusJSONValue(tc.b))
			require.False(t, diags.HasError(), "%v", diags)
			assert.Equal(t, tc.equal, eq)
		})
	}
}

func TestStatusJSONValue_EnginesNullAndMalformed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name  string
		a     string
		b     string
		equal bool
	}{
		{
			name:  "null engines equal",
			a:     `{"status":"running","engines":null}`,
			b:     `{"engines":null,"status":"running"}`,
			equal: true,
		},
		{
			name:  "null engines not empty array",
			a:     `{"status":"running","engines":null}`,
			b:     `{"status":"running","engines":[]}`,
			equal: false,
		},
		{
			name:  "missing engines key not equal to null engines (library semantics)",
			a:     `{"status":"running"}`,
			b:     `{"status":"running","engines":null}`,
			equal: false,
		},
		{
			name:  "missing type engine order remains significant (fallback)",
			a:     `{"status":"running","engines":[{"indexPattern":"a"},{"indexPattern":"b"}]}`,
			b:     `{"status":"running","engines":[{"indexPattern":"b"},{"indexPattern":"a"}]}`,
			equal: false,
		},
		{
			name:  "non-string type engine order remains significant (fallback)",
			a:     `{"status":"running","engines":[{"type":42},{"type":"user"}]}`,
			b:     `{"status":"running","engines":[{"type":"user"},{"type":42}]}`,
			equal: false,
		},
		{
			name:  "non-object engine element order remains significant (fallback)",
			a:     `{"status":"running","engines":["not-an-object",{"type":"user"}]}`,
			b:     `{"status":"running","engines":[{"type":"user"},"not-an-object"]}`,
			equal: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			eq, diags := NewStatusJSONValue(tc.a).StringSemanticEquals(ctx, NewStatusJSONValue(tc.b))
			require.False(t, diags.HasError(), "%v", diags)
			assert.Equal(t, tc.equal, eq)
		})
	}

	// Malformed JSON falls back to the raw string (no canonicalization), but the
	// embedded jsontypes.Normalized comparison then reports a diagnostic for the
	// unparsable payload rather than guessing equality.
	malformed := `{"status":"running","engines":`
	eq, diags := NewStatusJSONValue(malformed).StringSemanticEquals(ctx, NewStatusJSONValue(malformed))
	assert.False(t, eq)
	assert.True(t, diags.HasError(), "malformed JSON must surface the library semantic equality diagnostic")

	eq, diags = NewStatusJSONValue(malformed).StringSemanticEquals(ctx, NewStatusJSONValue(`{"status":"running","engines":[]}`))
	assert.False(t, eq)
	assert.True(t, diags.HasError(), "malformed JSON must surface the library semantic equality diagnostic")
}

func TestStatusJSONValue_EscapedTypeDuplicateStable(t *testing.T) {
	t.Parallel()

	// Two engines whose "type" values decode to the same string ("user",
	// one spelled with a unicode escape) are sort ties: the stable sort
	// must preserve their original relative order in the canonical copy.
	body := `{"engines":[{"type":"user","marker":"first"},{"type":"use\u0072","marker":"second"}]}`
	canonical := canonicalizeStatusJSONEngines(body)

	var decoded struct {
		Engines []struct {
			Marker string `json:"marker"`
		} `json:"engines"`
	}
	require.NoError(t, json.Unmarshal([]byte(canonical), &decoded))
	require.Len(t, decoded.Engines, 2)
	assert.Equal(t, []string{"first", "second"}, []string{decoded.Engines[0].Marker, decoded.Engines[1].Marker})
}

func TestStatusJSONValue_NullTypeSortsFirst(t *testing.T) {
	t.Parallel()

	// An explicit JSON null "type" decodes into the zero string without
	// error, so it sorts first; a missing or non-string "type" errors and
	// triggers the raw fallback.
	body := `{"engines":[{"type":"zeta"},{"type":null,"marker":"nulled"},{"type":"alpha"}]}`
	canonical := canonicalizeStatusJSONEngines(body)

	var decoded struct {
		Engines []struct {
			Type   *string `json:"type"`
			Marker string  `json:"marker"`
		} `json:"engines"`
	}
	require.NoError(t, json.Unmarshal([]byte(canonical), &decoded))
	require.Len(t, decoded.Engines, 3)
	require.Nil(t, decoded.Engines[0].Type)
	assert.Equal(t, "alpha", *decoded.Engines[1].Type)
	assert.Equal(t, "zeta", *decoded.Engines[2].Type)
}

func TestStatusJSONValue_WrongTypeDiagnostic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	eq, diags := NewStatusJSONValue(`{"status":"running"}`).StringSemanticEquals(ctx, jsontypes.NewNormalizedValue(`{"status":"running"}`))
	assert.False(t, eq)
	assert.True(t, diags.HasError(), "comparing against a jsontypes.Normalized value must raise the unexpected type diagnostic")
}
