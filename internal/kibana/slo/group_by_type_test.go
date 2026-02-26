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

package slo

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func mustGroupBy(t *testing.T, elems ...string) GroupByValue {
	t.Helper()

	values := make([]attr.Value, len(elems))
	for i, e := range elems {
		values[i] = types.StringValue(e)
	}

	v, diags := NewGroupByValue(values)
	require.False(t, diags.HasError(), "failed to create GroupByValue: %v", diags)
	return v
}

func TestGroupBy_ListSemanticEquals(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name   string
		left   GroupByValue
		right  GroupByValue
		expect bool
	}{
		{
			name:   "empty equals star",
			left:   mustGroupBy(t /* empty */),
			right:  mustGroupBy(t, "*"),
			expect: true,
		},
		{
			name:   "star equals empty",
			left:   mustGroupBy(t, "*"),
			right:  mustGroupBy(t /* empty */),
			expect: true,
		},
		{
			name:   "empty equals empty",
			left:   mustGroupBy(t /* empty */),
			right:  mustGroupBy(t /* empty */),
			expect: true,
		},
		{
			name:   "star equals star",
			left:   mustGroupBy(t, "*"),
			right:  mustGroupBy(t, "*"),
			expect: true,
		},
		{
			name:   "empty not equal foo",
			left:   mustGroupBy(t /* empty */),
			right:  mustGroupBy(t, "foo"),
			expect: false,
		},
		{
			name:   "star not equal foo",
			left:   mustGroupBy(t, "*"),
			right:  mustGroupBy(t, "foo"),
			expect: false,
		},
		{
			name:   "null equals null",
			left:   NewGroupByNull(),
			right:  NewGroupByNull(),
			expect: true,
		},
		{
			name:   "unknown equals unknown",
			left:   NewGroupByUnknown(),
			right:  NewGroupByUnknown(),
			expect: true,
		},
		{
			name:   "null not equal star",
			left:   NewGroupByNull(),
			right:  mustGroupBy(t, "*"),
			expect: false,
		},
		{
			name:   "unknown not equal star",
			left:   NewGroupByUnknown(),
			right:  mustGroupBy(t, "*"),
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, diags := tt.left.ListSemanticEquals(ctx, tt.right)
			require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
			require.Equal(t, tt.expect, got)
		})
	}
}
