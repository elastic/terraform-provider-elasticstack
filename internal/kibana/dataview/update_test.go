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

package dataview

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestBuildFieldAttrsMetadataDelta(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		planFA  map[string]fieldAttrModel
		stateFA map[string]fieldAttrModel
		want    map[string]any
	}{
		{
			name:    "both empty produces empty delta",
			planFA:  map[string]fieldAttrModel{},
			stateFA: map[string]fieldAttrModel{},
			want:    map[string]any{},
		},
		{
			name:    "nil maps produce empty delta",
			planFA:  nil,
			stateFA: nil,
			want:    map[string]any{},
		},
		{
			name: "added field with custom_label only",
			planFA: map[string]fieldAttrModel{
				"host.hostname": {CustomLabel: types.StringValue("Host"), Count: types.Int64Null()},
			},
			stateFA: map[string]fieldAttrModel{},
			want: map[string]any{
				"host.hostname": map[string]any{"customLabel": "Host"},
			},
		},
		{
			name: "added field with custom_label and count",
			planFA: map[string]fieldAttrModel{
				"f1": {CustomLabel: types.StringValue("F1"), Count: types.Int64Value(7)},
			},
			stateFA: map[string]fieldAttrModel{},
			want: map[string]any{
				"f1": map[string]any{"customLabel": "F1", "count": int64(7)},
			},
		},
		{
			name: "unchanged field omitted from delta",
			planFA: map[string]fieldAttrModel{
				"f1": {CustomLabel: types.StringValue("Same"), Count: types.Int64Null()},
			},
			stateFA: map[string]fieldAttrModel{
				"f1": {CustomLabel: types.StringValue("Same"), Count: types.Int64Null()},
			},
			want: map[string]any{},
		},
		{
			name: "changed custom_label produces full payload",
			planFA: map[string]fieldAttrModel{
				"f1": {CustomLabel: types.StringValue("New"), Count: types.Int64Null()},
			},
			stateFA: map[string]fieldAttrModel{
				"f1": {CustomLabel: types.StringValue("Old"), Count: types.Int64Null()},
			},
			want: map[string]any{
				"f1": map[string]any{"customLabel": "New"},
			},
		},
		{
			name:   "removed field clears metadata with explicit null keys",
			planFA: map[string]fieldAttrModel{},
			stateFA: map[string]fieldAttrModel{
				"f1": {CustomLabel: types.StringValue("Bye"), Count: types.Int64Null()},
			},
			want: map[string]any{
				"f1": map[string]any{"customLabel": nil, "count": nil},
			},
		},
		{
			name: "mixed add change remove keeps unchanged out",
			planFA: map[string]fieldAttrModel{
				"added":     {CustomLabel: types.StringValue("A"), Count: types.Int64Null()},
				"changed":   {CustomLabel: types.StringValue("C2"), Count: types.Int64Null()},
				"unchanged": {CustomLabel: types.StringValue("U"), Count: types.Int64Null()},
			},
			stateFA: map[string]fieldAttrModel{
				"changed":   {CustomLabel: types.StringValue("C1"), Count: types.Int64Null()},
				"unchanged": {CustomLabel: types.StringValue("U"), Count: types.Int64Null()},
				"removed":   {CustomLabel: types.StringValue("R"), Count: types.Int64Null()},
			},
			want: map[string]any{
				"added":   map[string]any{"customLabel": "A"},
				"changed": map[string]any{"customLabel": "C2"},
				"removed": map[string]any{"customLabel": nil, "count": nil},
			},
		},
		{
			name: "count-only payload when label is null",
			planFA: map[string]fieldAttrModel{
				"f1": {CustomLabel: types.StringNull(), Count: types.Int64Value(3)},
			},
			stateFA: map[string]fieldAttrModel{},
			want: map[string]any{
				"f1": map[string]any{"count": int64(3)},
			},
		},
		{
			name: "explicit count change is detected",
			planFA: map[string]fieldAttrModel{
				"f1": {CustomLabel: types.StringValue("L"), Count: types.Int64Value(2)},
			},
			stateFA: map[string]fieldAttrModel{
				"f1": {CustomLabel: types.StringValue("L"), Count: types.Int64Value(1)},
			},
			want: map[string]any{
				"f1": map[string]any{"customLabel": "L", "count": int64(2)},
			},
		},
		{
			name: "all-null plan entry is skipped (no ambiguous empty payload)",
			planFA: map[string]fieldAttrModel{
				"f1": {CustomLabel: types.StringNull(), Count: types.Int64Null()},
			},
			stateFA: map[string]fieldAttrModel{},
			want:    map[string]any{},
		},
		{
			name:   "server-only count-only state entry is not cleared on removal",
			planFA: map[string]fieldAttrModel{},
			stateFA: map[string]fieldAttrModel{
				"host.hostname": {CustomLabel: types.StringNull(), Count: types.Int64Value(5)},
			},
			want: map[string]any{},
		},
		{
			name: "server-only count entry adjacent to user removal is not cleared",
			planFA: map[string]fieldAttrModel{
				"keep": {CustomLabel: types.StringValue("Keep"), Count: types.Int64Null()},
			},
			stateFA: map[string]fieldAttrModel{
				"keep":          {CustomLabel: types.StringValue("Keep"), Count: types.Int64Null()},
				"removed":       {CustomLabel: types.StringValue("Bye"), Count: types.Int64Null()},
				"host.hostname": {CustomLabel: types.StringNull(), Count: types.Int64Value(5)},
			},
			want: map[string]any{
				"removed": map[string]any{"customLabel": nil, "count": nil},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildFieldAttrsMetadataDelta(tt.planFA, tt.stateFA)
			require.Equal(t, tt.want, got)
		})
	}
}
