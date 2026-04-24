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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/stretchr/testify/assert"
)

// TestFilterFieldAttrs pins the behaviour that resolves
// https://github.com/elastic/terraform-provider-elasticstack/issues/1287:
// Kibana Discover auto-populates `field_attrs.<field>.count` after a data
// view is used in the UI. The provider used to surface those server-side
// entries as drift, and because `field_attrs` carries `RequiresReplace`,
// the entire resource (and its `id`) would be rebuilt on the next apply —
// breaking every dashboard that referenced the prior data-view id.
//
// The filter drops entries whose `custom_label` is nil unconditionally, so
// upgrading the provider also heals state that was polluted by older
// versions that wrote count-only entries into it.
func TestFilterFieldAttrs(t *testing.T) {
	t.Parallel()

	str := func(s string) *string { return &s }
	i64 := func(i int) *int { return &i }

	tests := []struct {
		name     string
		api      map[string]kbapi.DataViewsFieldattrs
		wantKeys []string
	}{
		{
			name: "empty API returns empty",
			api:  map[string]kbapi.DataViewsFieldattrs{},
		},
		{
			name: "nil API returns nil",
			api:  nil,
		},
		{
			name: "server-only count-only entries are dropped",
			api: map[string]kbapi.DataViewsFieldattrs{
				"host.hostname": {Count: i64(5)},
				"event.action":  {Count: i64(12)},
			},
		},
		{
			name: "entries with custom_label are kept, count-only siblings are dropped",
			api: map[string]kbapi.DataViewsFieldattrs{
				"message":       {CustomLabel: str("Log Message"), Count: i64(42)},
				"host.hostname": {Count: i64(5)},
			},
			wantKeys: []string{"message"},
		},
		{
			name: "count-only entries are dropped even when they are already in polluted state (self-heals)",
			// Simulates a refresh after upgrading from an older provider
			// version that had already pushed these into Terraform state.
			api: map[string]kbapi.DataViewsFieldattrs{
				"field.a": {Count: i64(1)},
				"field.b": {Count: i64(2)},
				"field.c": {Count: i64(3)},
			},
		},
		{
			name: "custom_label with nil count is kept (user configured it)",
			api: map[string]kbapi.DataViewsFieldattrs{
				"only.label": {CustomLabel: str("Labelled Only")},
			},
			wantKeys: []string{"only.label"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := filterFieldAttrs(tc.api)

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
