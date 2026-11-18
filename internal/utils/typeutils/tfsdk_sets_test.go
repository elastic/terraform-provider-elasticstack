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
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestSetFromAPIStringsPreserveKnownEmpty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	knownEmpty := types.SetValueMust(types.StringType, []attr.Value{})

	tests := []struct {
		name     string
		apiValue *[]string
		dest     types.Set
		want     types.Set
	}{
		{
			name:     "nil apiValue with unknown dest collapses to null",
			apiValue: nil,
			dest:     types.SetUnknown(types.StringType),
			want:     types.SetNull(types.StringType),
		},
		{
			name:     "empty apiValue with unknown dest collapses to null",
			apiValue: &[]string{},
			dest:     types.SetUnknown(types.StringType),
			want:     types.SetNull(types.StringType),
		},
		{
			name:     "nil apiValue with known-empty dest is preserved",
			apiValue: nil,
			dest:     knownEmpty,
			want:     knownEmpty,
		},
		{
			name:     "nil apiValue with null dest stays null",
			apiValue: nil,
			dest:     types.SetNull(types.StringType),
			want:     types.SetNull(types.StringType),
		},
		{
			name:     "non-empty apiValue overrides dest",
			apiValue: &[]string{"a", "b"},
			dest:     types.SetUnknown(types.StringType),
			want:     types.SetValueMust(types.StringType, []attr.Value{types.StringValue("a"), types.StringValue("b")}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, diags := typeutils.SetFromAPIStringsPreserveKnownEmpty(ctx, tt.apiValue, tt.dest)
			require.False(t, diags.HasError())
			require.True(t, tt.want.Equal(got))
		})
	}
}
