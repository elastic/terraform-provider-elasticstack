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
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func Test_isSemanticallyEquivalentMissing(t *testing.T) {
	tests := []struct {
		name     string
		planned  types.String
		existing string
		want     bool
	}{
		{"null equals absent", types.StringNull(), "", true},
		{"null equals _last", types.StringNull(), "_last", true},
		{"_last equals null existing", types.StringValue("_last"), "", true},
		{"_last equals _last", types.StringValue("_last"), "_last", true},
		{"_first differs from null", types.StringValue("_first"), "", false},
		{"_first differs from _last", types.StringValue("_first"), "_last", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSemanticallyEquivalentMissing(tt.planned, tt.existing)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_isSemanticallyEquivalentMode(t *testing.T) {
	tests := []struct {
		name     string
		planned  types.String
		existing string
		order    types.String
		want     bool
	}{
		{"null with asc equals absent", types.StringNull(), "", types.StringValue("asc"), true},
		{"null with asc equals min", types.StringNull(), "min", types.StringValue("asc"), true},
		{"min with asc equals min", types.StringValue("min"), "min", types.StringValue("asc"), true},
		{"min with asc equals absent", types.StringValue("min"), "", types.StringValue("asc"), true},
		{"null with desc equals absent", types.StringNull(), "", types.StringValue("desc"), true},
		{"null with desc equals max", types.StringNull(), "max", types.StringValue("desc"), true},
		{"max with desc equals max", types.StringValue("max"), "max", types.StringValue("desc"), true},
		{"max with asc differs from absent", types.StringValue("max"), "", types.StringValue("asc"), false},
		{"min with desc differs from absent", types.StringValue("min"), "", types.StringValue("desc"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSemanticallyEquivalentMode(tt.planned, tt.existing, tt.order)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_extractSortSetting(t *testing.T) {
	tests := []struct {
		name     string
		settings map[string]any
		key      string
		want     []string
	}{
		{
			name:     "bare key string slice",
			settings: map[string]any{"sort.field": []any{"date", "id"}},
			key:      "sort.field",
			want:     []string{"date", "id"},
		},
		{
			name:     "prefixed key",
			settings: map[string]any{"index.sort.field": []any{"date"}},
			key:      "sort.field",
			want:     []string{"date"},
		},
		{
			name:     "missing key returns nil",
			settings: map[string]any{"other": "value"},
			key:      "sort.missing",
			want:     nil,
		},
		{
			name:     "single string value",
			settings: map[string]any{"sort.field": "date"},
			key:      "sort.field",
			want:     []string{"date"},
		},
		{
			name:     "sort.missing as string slice",
			settings: map[string]any{"sort.missing": []any{"_last", "_first"}},
			key:      "sort.missing",
			want:     []string{"_last", "_first"},
		},
		{
			name:     "sort.mode as string slice",
			settings: map[string]any{"sort.mode": []any{"min", "max"}},
			key:      "sort.mode",
			want:     []string{"min", "max"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractSortSetting(tt.settings, tt.key)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_sortPrivateState_MarshalRoundTrip(t *testing.T) {
	ps := sortPrivateState{
		Fields:  []string{"date", "id"},
		Orders:  []string{"desc", "asc"},
		Missing: []string{"_last"},
		Mode:    []string{"max"},
	}

	data, err := json.Marshal(ps)
	require.NoError(t, err)

	var decoded sortPrivateState
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	require.Equal(t, ps, decoded)
}
