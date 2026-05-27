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

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeDrift(t *testing.T) {
	tests := []struct {
		name     string
		features curatedFeatures
		api      []string
		expected driftReport
	}{
		{
			name: "detects unknown and removed features",
			features: curatedFeatures{
				Documented: []string{"apm", "discover_v2", "fleetv2"},
				Skip:       []string{"internalFeature"},
			},
			api: []string{"apm", "logs", "internalFeature"},
			expected: driftReport{
				UnknownFeatures: []string{"logs"},
				RemovedFeatures: []string{"discover_v2", "fleetv2"},
			},
		},
		{
			name: "no drift when all api features are documented or skipped",
			features: curatedFeatures{
				Documented: []string{"apm", "logs"},
				Skip:       []string{"internalFeature"},
			},
			api: []string{"logs", "apm", "internalFeature"},
			expected: driftReport{
				UnknownFeatures: []string{},
				RemovedFeatures: []string{},
			},
		},
		{
			name: "deduplicates and ignores blanks",
			features: curatedFeatures{
				Documented: []string{"apm", "apm", " "},
				Skip:       []string{"skipme", "skipme"},
			},
			api: []string{"", " skipme ", "logs", "logs", "apm"},
			expected: driftReport{
				UnknownFeatures: []string{"logs"},
				RemovedFeatures: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, computeDrift(tt.features, tt.api))
		})
	}
}
