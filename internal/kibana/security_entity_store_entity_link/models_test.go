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

package security_entity_store_entity_link

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeSetDiff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		old, new    []string
		wantAdded   []string
		wantRemoved []string
	}{
		{
			name:        "add one",
			old:         []string{"a"},
			new:         []string{"a", "b"},
			wantAdded:   []string{"b"},
			wantRemoved: nil,
		},
		{
			name:        "remove one",
			old:         []string{"a", "b"},
			new:         []string{"a"},
			wantAdded:   nil,
			wantRemoved: []string{"b"},
		},
		{
			name:        "add and remove",
			old:         []string{"a", "b"},
			new:         []string{"b", "c"},
			wantAdded:   []string{"c"},
			wantRemoved: []string{"a"},
		},
		{
			name:        "no change",
			old:         []string{"a", "b"},
			new:         []string{"a", "b"},
			wantAdded:   nil,
			wantRemoved: nil,
		},
		{
			name:        "all new",
			old:         []string{"a"},
			new:         []string{"b", "c"},
			wantAdded:   []string{"b", "c"},
			wantRemoved: []string{"a"},
		},
		{
			name:        "all removed",
			old:         []string{"a", "b"},
			new:         []string{},
			wantAdded:   nil,
			wantRemoved: []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAdded, gotRemoved := computeSetDiff(tt.old, tt.new)
			assert.ElementsMatch(t, tt.wantAdded, gotAdded)
			assert.ElementsMatch(t, tt.wantRemoved, gotRemoved)
		})
	}
}

func TestExtractEntityIDsFromPayload(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		payload  map[string]any
		targetID string
		want     []string
	}{
		{
			name:     "basic array without target",
			payload:  map[string]any{"entity_ids": []any{"a", "b", "c"}},
			targetID: "t",
			want:     []string{"a", "b", "c"},
		},
		{
			name:     "filters out target_id",
			payload:  map[string]any{"entity_ids": []any{"t", "a", "b"}},
			targetID: "t",
			want:     []string{"a", "b"},
		},
		{
			name:     "missing entity_ids key",
			payload:  map[string]any{"other": []any{"a"}},
			targetID: "t",
			want:     nil,
		},
		{
			name:     "ignores non-string elements",
			payload:  map[string]any{"entity_ids": []any{"a", 42, "b"}},
			targetID: "t",
			want:     []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractEntityIDsFromPayload(tt.payload, tt.targetID)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestContainsAll(t *testing.T) {
	t.Parallel()

	assert.True(t, containsAll([]string{"a", "b", "c"}, []string{"a", "b"}))
	assert.True(t, containsAll([]string{"a", "b"}, []string{"a", "b"}))
	assert.False(t, containsAll([]string{"a"}, []string{"a", "b"}))
	assert.True(t, containsAll([]string{"a", "b"}, []string{}))
}
