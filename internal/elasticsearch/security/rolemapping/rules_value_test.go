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

package rolemapping

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/stretchr/testify/require"
)

func TestNormalizedRulesValue_StringSemanticEquals(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name      string
		a         string
		b         string
		wantEqual bool
	}{
		{
			name:      "single-element array vs plain string",
			a:         `{"field":{"groups":["project1"]}}`,
			b:         `{"field":{"groups":"project1"}}`,
			wantEqual: true,
		},
		{
			name:      "both array form",
			a:         `{"field":{"groups":["project1"]}}`,
			b:         `{"field":{"groups":["project1"]}}`,
			wantEqual: true,
		},
		{
			name:      "both string form",
			a:         `{"field":{"groups":"project1"}}`,
			b:         `{"field":{"groups":"project1"}}`,
			wantEqual: true,
		},
		{
			name:      "multi-element array vs different value",
			a:         `{"field":{"groups":["a","b"]}}`,
			b:         `{"field":{"groups":["a"]}}`,
			wantEqual: false,
		},
		{
			name:      "whitespace difference",
			a:         `{"field":{"groups":"x"}}`,
			b:         `{ "field": { "groups": "x" } }`,
			wantEqual: true,
		},
		{
			name:      "nested any with single-element array vs string",
			a:         `{"any":[{"field":{"groups":["x"]}}]}`,
			b:         `{"any":[{"field":{"groups":"x"}}]}`,
			wantEqual: true,
		},
		{
			name:      "nested all with single-element array vs string",
			a:         `{"all":[{"field":{"groups":["x"]}}]}`,
			b:         `{"all":[{"field":{"groups":"x"}}]}`,
			wantEqual: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := NewNormalizedRulesValue(tt.a)
			b := NewNormalizedRulesValue(tt.b)

			eq, diags := a.StringSemanticEquals(ctx, b)
			require.False(t, diags.HasError(), diags)
			require.Equal(t, tt.wantEqual, eq)

			rev, revDiags := b.StringSemanticEquals(ctx, a)
			require.False(t, revDiags.HasError(), revDiags)
			require.Equal(t, tt.wantEqual, rev)
		})
	}
}

func TestNormalizedRulesValue_StringSemanticEquals_null(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	nullA := NewNormalizedRulesNull()
	nullB := NewNormalizedRulesNull()
	known := NewNormalizedRulesValue(`{"field":{"groups":"x"}}`)

	eq, diags := nullA.StringSemanticEquals(ctx, nullB)
	require.False(t, diags.HasError())
	require.True(t, eq)

	eq, diags = nullA.StringSemanticEquals(ctx, known)
	require.False(t, diags.HasError())
	require.False(t, eq)
}

func TestNormalizedRulesValue_StringSemanticEquals_unknown(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	unknownA := NewNormalizedRulesUnknown()
	unknownB := NewNormalizedRulesUnknown()
	known := NewNormalizedRulesValue(`{"field":{"groups":"x"}}`)
	null := NewNormalizedRulesNull()

	eq, diags := unknownA.StringSemanticEquals(ctx, unknownB)
	require.False(t, diags.HasError())
	require.True(t, eq)

	eq, diags = known.StringSemanticEquals(ctx, unknownA)
	require.False(t, diags.HasError())
	require.False(t, eq)

	eq, diags = null.StringSemanticEquals(ctx, unknownA)
	require.False(t, diags.HasError())
	require.False(t, eq)
}

func TestNormalizedRulesValue_StringSemanticEquals_parseErrorFallback(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	invalid := NewNormalizedRulesValue(`not json`)
	other := NewNormalizedRulesValue(`also not json`)

	eq, diags := invalid.StringSemanticEquals(ctx, other)
	require.True(t, diags.HasError())
	require.False(t, eq)
}

func TestNormalizedRulesValue_StringSemanticEquals_nonRulesTypeFallback(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	rules := NewNormalizedRulesValue(`{"field":{"groups":"x"}}`)
	plain := jsontypes.NewNormalizedValue(`{"field":{"groups":"x"}}`)

	eq, diags := rules.StringSemanticEquals(ctx, plain)
	require.False(t, diags.HasError())
	require.True(t, eq)
}
