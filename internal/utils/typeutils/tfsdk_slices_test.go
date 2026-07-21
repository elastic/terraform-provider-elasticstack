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

package typeutils_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestValueStringSlice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []types.String
		want  []string
	}{
		{name: "nil slice", input: nil, want: nil},
		{name: "empty slice", input: []types.String{}, want: nil},
		{name: "one element", input: []types.String{types.StringValue("a")}, want: []string{"a"}},
		{name: "multiple elements", input: []types.String{types.StringValue("x"), types.StringValue("y"), types.StringValue("z")}, want: []string{"x", "y", "z"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := typeutils.ValueStringSlice(tt.input)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestStringSliceValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []string
		want  []types.String
	}{
		{name: "nil slice", input: nil, want: nil},
		{name: "empty slice", input: []string{}, want: nil},
		{name: "one element", input: []string{"a"}, want: []types.String{types.StringValue("a")}},
		{name: "multiple elements", input: []string{"x", "y", "z"}, want: []types.String{types.StringValue("x"), types.StringValue("y"), types.StringValue("z")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := typeutils.StringSliceValue(tt.input)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestStringsToListMust(t *testing.T) {
	t.Parallel()

	emptyList := types.ListValueMust(types.StringType, []attr.Value{})
	fullList := types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("a"),
		types.StringValue("b"),
	})

	tests := []struct {
		name  string
		input []string
		want  types.List
	}{
		{name: "nil slice", input: nil, want: emptyList},
		{name: "empty slice", input: []string{}, want: emptyList},
		{name: "multiple elements", input: []string{"a", "b"}, want: fullList},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := typeutils.StringsToListMust(tt.input)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestListToStringsMust(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input types.List
		want  []string
	}{
		{name: "null list", input: types.ListNull(types.StringType), want: nil},
		{name: "unknown list", input: types.ListUnknown(types.StringType), want: nil},
		{name: "empty list", input: types.ListValueMust(types.StringType, []attr.Value{}), want: []string{}},
		{
			name: "multiple elements",
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("x"),
				types.StringValue("y"),
			}),
			want: []string{"x", "y"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := typeutils.ListToStringsMust(tt.input)
			require.Equal(t, tt.want, got)
		})
	}
}
