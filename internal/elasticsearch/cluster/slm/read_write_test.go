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

package slm

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestReadSlmIndicesRoundtrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		name         string
		apiIndices   []string
		stateIndices types.List
		wantNull     bool
		wantEmpty    bool
		wantElements []string
	}{
		{
			name:         "api returns indices, state is null",
			apiIndices:   []string{"idx1", "idx2"},
			stateIndices: types.ListNull(types.StringType),
			wantElements: []string{"idx1", "idx2"},
		},
		{
			name:         "api returns indices, state has values",
			apiIndices:   []string{"idx1", "idx2"},
			stateIndices: mustList(ctx, t, []string{"a", "b"}),
			wantElements: []string{"idx1", "idx2"},
		},
		{
			name:         "api omits indices, state is null → null",
			apiIndices:   nil,
			stateIndices: types.ListNull(types.StringType),
			wantNull:     true,
		},
		{
			name:         "api omits indices, state is empty → empty",
			apiIndices:   nil,
			stateIndices: mustList(ctx, t, []string{}),
			wantEmpty:    true,
		},
		{
			name:         "api omits indices, state has elements → empty",
			apiIndices:   nil,
			stateIndices: mustList(ctx, t, []string{"a"}),
			wantEmpty:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			slm := &elasticsearch.SlmPolicy{
				Repository: "repo",
				Schedule:   "0 30 1 * * ?",
				Name:       "snap",
				Config: &elasticsearch.SlmConfig{
					Indices: tc.apiIndices,
				},
			}

			state := Data{Indices: tc.stateIndices}
			data, diags := mapSlmToData(ctx, slm, "test", state)
			require.False(t, diags.HasError(), "unexpected error: %s", diags)

			if tc.wantNull {
				require.True(t, data.Indices.IsNull(), "expected null indices")
				return
			}
			if tc.wantEmpty {
				require.False(t, data.Indices.IsNull(), "expected non-null indices")
				require.Empty(t, data.Indices.Elements(), "expected empty list")
				return
			}
			var got []string
			require.False(t, data.Indices.ElementsAs(ctx, &got, false).HasError())
			require.Equal(t, tc.wantElements, got)
		})
	}
}

func TestReadSlmFeatureStatesRoundtrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		name               string
		apiFeatureStates   []string
		stateFeatureStates types.Set
		wantNull           bool
		wantEmpty          bool
		wantElements       []string
	}{
		{
			name:               "api returns feature states, state is null",
			apiFeatureStates:   []string{"kibana"},
			stateFeatureStates: types.SetNull(types.StringType),
			wantElements:       []string{"kibana"},
		},
		{
			name:               "api omits feature states, state is null → null",
			apiFeatureStates:   nil,
			stateFeatureStates: types.SetNull(types.StringType),
			wantNull:           true,
		},
		{
			name:               "api omits feature states, state is empty → empty",
			apiFeatureStates:   nil,
			stateFeatureStates: mustSet(ctx, t, []string{}),
			wantEmpty:          true,
		},
		{
			name:               "api omits feature states, state has elements → empty",
			apiFeatureStates:   nil,
			stateFeatureStates: mustSet(ctx, t, []string{"kibana"}),
			wantEmpty:          true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			slm := &elasticsearch.SlmPolicy{
				Repository: "repo",
				Schedule:   "0 30 1 * * ?",
				Name:       "snap",
				Config: &elasticsearch.SlmConfig{
					FeatureStates: tc.apiFeatureStates,
				},
			}

			state := Data{FeatureStates: tc.stateFeatureStates}
			data, diags := mapSlmToData(ctx, slm, "test", state)
			require.False(t, diags.HasError())

			if tc.wantNull {
				require.True(t, data.FeatureStates.IsNull())
				return
			}
			if tc.wantEmpty {
				require.False(t, data.FeatureStates.IsNull())
				require.Empty(t, data.FeatureStates.Elements())
				return
			}
			var got []string
			require.False(t, data.FeatureStates.ElementsAs(ctx, &got, false).HasError())
			require.Equal(t, tc.wantElements, got)
		})
	}
}

func TestReadSlmMetadataRoundtrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		name       string
		apiMeta    map[string]json.RawMessage
		wantMetaJS string
		wantNull   bool
	}{
		{
			name:       "metadata with string and number",
			apiMeta:    map[string]json.RawMessage{"team": json.RawMessage(`"search"`), "retention": json.RawMessage(`30`)},
			wantMetaJS: `{"retention":30,"team":"search"}`,
		},
		{
			name:       "metadata with nested object",
			apiMeta:    map[string]json.RawMessage{"tags": json.RawMessage(`["a","b"]`), "info": json.RawMessage(`{"key":"val"}`)},
			wantMetaJS: `{"info":{"key":"val"},"tags":["a","b"]}`,
		},
		{
			name:     "no metadata",
			apiMeta:  nil,
			wantNull: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			slm := &elasticsearch.SlmPolicy{
				Repository: "repo",
				Schedule:   "0 30 1 * * ?",
				Name:       "snap",
				Config: &elasticsearch.SlmConfig{
					Metadata: tc.apiMeta,
				},
			}

			state := Data{}
			data, diags := mapSlmToData(ctx, slm, "test", state)
			require.False(t, diags.HasError())

			if tc.wantNull {
				require.True(t, data.Metadata.IsNull())
				return
			}
			require.False(t, data.Metadata.IsNull())
			var got map[string]any
			require.NoError(t, json.Unmarshal([]byte(data.Metadata.ValueString()), &got))
			var want map[string]any
			require.NoError(t, json.Unmarshal([]byte(tc.wantMetaJS), &want))
			require.Equal(t, want, got)
		})
	}
}

func TestReadSlmDefaultsWhenConfigNil(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	slm := &elasticsearch.SlmPolicy{
		Repository: "repo",
		Schedule:   "0 30 1 * * ?",
		Name:       "snap",
		Config:     nil,
	}

	state := Data{}
	data, diags := mapSlmToData(ctx, slm, "test", state)
	require.False(t, diags.HasError())

	require.Equal(t, defaultExpandWildcards, data.ExpandWildcards.ValueString())
	require.True(t, data.IncludeGlobalState.ValueBool())
	require.False(t, data.IgnoreUnavailable.ValueBool())
	require.False(t, data.Partial.ValueBool())
	require.True(t, data.Indices.IsNull())
	require.True(t, data.FeatureStates.IsNull())
	require.True(t, data.Metadata.IsNull())
}

func mustList(ctx context.Context, t *testing.T, elems []string) types.List {
	l, diags := types.ListValueFrom(ctx, types.StringType, elems)
	require.False(t, diags.HasError())
	return l
}

func mustSet(ctx context.Context, t *testing.T, elems []string) types.Set {
	s, diags := types.SetValueFrom(ctx, types.StringType, elems)
	require.False(t, diags.HasError())
	return s
}
