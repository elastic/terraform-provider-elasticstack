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
		old, next   []string
		wantAdded   []string
		wantRemoved []string
	}{
		{
			name:        "add one",
			old:         []string{"a"},
			next:        []string{"a", "b"},
			wantAdded:   []string{"b"},
			wantRemoved: nil,
		},
		{
			name:        "remove one",
			old:         []string{"a", "b"},
			next:        []string{"a"},
			wantAdded:   nil,
			wantRemoved: []string{"b"},
		},
		{
			name:        "add and remove",
			old:         []string{"a", "b"},
			next:        []string{"b", "c"},
			wantAdded:   []string{"c"},
			wantRemoved: []string{"a"},
		},
		{
			name:        "no change",
			old:         []string{"a", "b"},
			next:        []string{"a", "b"},
			wantAdded:   nil,
			wantRemoved: nil,
		},
		{
			name:        "all new",
			old:         []string{"a"},
			next:        []string{"b", "c"},
			wantAdded:   []string{"b", "c"},
			wantRemoved: []string{"a"},
		},
		{
			name:        "all removed",
			old:         []string{"a", "b"},
			next:        []string{},
			wantAdded:   nil,
			wantRemoved: []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAdded, gotRemoved := computeSetDiff(tt.old, tt.next)
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
			name: "basic aliases without target",
			payload: map[string]any{
				"aliases": []any{
					map[string]any{"entity": map[string]any{"id": "a"}},
					map[string]any{"entity": map[string]any{"id": "b"}},
					map[string]any{"entity": map[string]any{"id": "c"}},
				},
			},
			targetID: "t",
			want:     []string{"a", "b", "c"},
		},
		{
			name: "filters out target_id",
			payload: map[string]any{
				"aliases": []any{
					map[string]any{"entity": map[string]any{"id": "t"}},
					map[string]any{"entity": map[string]any{"id": "a"}},
					map[string]any{"entity": map[string]any{"id": "b"}},
				},
			},
			targetID: "t",
			want:     []string{"a", "b"},
		},
		{
			name:     "missing aliases key",
			payload:  map[string]any{"other": []any{"a"}},
			targetID: "t",
			want:     nil,
		},
		{
			name: "ignores malformed alias entries",
			payload: map[string]any{
				"aliases": []any{
					map[string]any{"entity": map[string]any{"id": "a"}},
					map[string]any{"entity": "not-a-map"},
					42,
					map[string]any{"no_entity": true},
				},
			},
			targetID: "t",
			want:     []string{"a"},
		},
		{
			name: "empty aliases",
			payload: map[string]any{
				"aliases": []any{},
				"target":  map[string]any{"entity": map[string]any{"id": "t"}},
			},
			targetID: "t",
			want:     nil,
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
