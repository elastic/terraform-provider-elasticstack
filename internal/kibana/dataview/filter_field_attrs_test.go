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
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFilterFieldAttrs pins the behaviour that resolves
// https://github.com/elastic/terraform-provider-elasticstack/issues/1287:
// Kibana Discover auto-populates `field_attrs.<field>.count` after a data
// view is used in the UI. The provider used to surface those server-side
// entries as drift, and because `field_attrs` carries `RequiresReplace`,
// the entire resource (and its `id`) would be rebuilt on the next apply —
// breaking every dashboard that referenced the prior data-view id.
func TestFilterFieldAttrs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	str := func(s string) *string { return &s }
	i64 := func(i int) *int { return &i }

	elemType := getFieldAttrElemType()

	buildPriorState := func(names ...string) types.Map {
		if len(names) == 0 {
			return types.MapNull(elemType)
		}
		var diags diag.Diagnostics
		m := make(map[string]fieldAttrModel, len(names))
		for _, n := range names {
			m[n] = fieldAttrModel{
				CustomLabel: types.StringValue("user-configured:" + n),
				Count:       types.Int64Null(),
			}
		}
		result := typeutils.MapValueFrom(ctx, m, elemType, path.Root("data_view").AtName("field_attrs"), &diags)
		require.False(t, diags.HasError(), "test setup: %v", diags)
		return result
	}

	tests := []struct {
		name      string
		api       map[string]kbapi.DataViewsFieldattrs
		prior     types.Map
		wantKeys  []string
	}{
		{
			name:     "empty API returns empty",
			api:      map[string]kbapi.DataViewsFieldattrs{},
			prior:    types.MapNull(elemType),
			wantKeys: nil,
		},
		{
			name:     "nil API returns nil",
			api:      nil,
			prior:    types.MapNull(elemType),
			wantKeys: nil,
		},
		{
			name: "server-only count-only entries dropped when not in prior state",
			api: map[string]kbapi.DataViewsFieldattrs{
				"host.hostname": {Count: i64(5)},
				"event.action":  {Count: i64(12)},
			},
			prior:    types.MapNull(elemType),
			wantKeys: nil,
		},
		{
			name: "entries with custom_label kept even when also count-populated",
			api: map[string]kbapi.DataViewsFieldattrs{
				"message":       {CustomLabel: str("Log Message"), Count: i64(42)},
				"host.hostname": {Count: i64(5)}, // server-only, no state
			},
			prior:    types.MapNull(elemType),
			wantKeys: []string{"message"},
		},
		{
			name: "server entries kept when their key is already in prior state",
			api: map[string]kbapi.DataViewsFieldattrs{
				"message":       {CustomLabel: str("Log Message"), Count: i64(42)},
				"host.hostname": {Count: i64(5)}, // count-only but tracked previously
			},
			prior:    buildPriorState("host.hostname"),
			wantKeys: []string{"message", "host.hostname"},
		},
		{
			name: "all-server entries with empty prior state → empty output",
			api: map[string]kbapi.DataViewsFieldattrs{
				"field.a": {Count: i64(1)},
				"field.b": {Count: i64(2)},
				"field.c": {Count: i64(3)},
			},
			prior:    buildPriorState(),
			wantKeys: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := filterFieldAttrs(tc.api, tc.prior)

			if len(tc.wantKeys) == 0 {
				assert.Empty(t, got, "expected filter to drop all entries, got %v", got)
				return
			}
			gotKeys := make([]string, 0, len(got))
			for k := range got {
				gotKeys = append(gotKeys, k)
			}
			assert.ElementsMatch(t, tc.wantKeys, gotKeys,
				"expected filtered keys %v, got %v", tc.wantKeys, gotKeys)
		})
	}
}
