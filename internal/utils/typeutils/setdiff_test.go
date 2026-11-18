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
	"github.com/stretchr/testify/require"
)

func TestDiffStringSlices(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		before      []string
		after       []string
		wantAdded   []string
		wantRemoved []string
	}{
		{
			name:        "add one",
			before:      []string{"a"},
			after:       []string{"a", "b"},
			wantAdded:   []string{"b"},
			wantRemoved: nil,
		},
		{
			name:        "remove one",
			before:      []string{"a", "b"},
			after:       []string{"a"},
			wantAdded:   nil,
			wantRemoved: []string{"b"},
		},
		{
			name:        "add and remove",
			before:      []string{"a", "b"},
			after:       []string{"b", "c"},
			wantAdded:   []string{"c"},
			wantRemoved: []string{"a"},
		},
		{
			name:        "no change",
			before:      []string{"a", "b"},
			after:       []string{"a", "b"},
			wantAdded:   nil,
			wantRemoved: nil,
		},
		{
			name:        "all new, sorted output",
			before:      []string{"a"},
			after:       []string{"c", "b"},
			wantAdded:   []string{"b", "c"},
			wantRemoved: []string{"a"},
		},
		{
			name:        "all removed",
			before:      []string{"a", "b"},
			after:       []string{},
			wantAdded:   nil,
			wantRemoved: []string{"a", "b"},
		},
		{
			name:        "both empty",
			before:      nil,
			after:       nil,
			wantAdded:   nil,
			wantRemoved: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAdded, gotRemoved := typeutils.DiffStringSlices(tt.before, tt.after)
			require.Equal(t, tt.wantAdded, gotAdded)
			require.Equal(t, tt.wantRemoved, gotRemoved)
		})
	}
}
