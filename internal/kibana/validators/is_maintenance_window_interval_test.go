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

package validators

import (
	"testing"
)

func TestStringMatchesIntervalFrequencyRegex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		intervalFrequency string
		matched           bool
	}{

		{
			name:              "valid interval/frequency string (2d)",
			intervalFrequency: "2d",
			matched:           true,
		},
		{
			name:              "valid interval/frequency string (5w)",
			intervalFrequency: "5w",
			matched:           true,
		},
		{
			name:              "valid interval/frequency string (3M)",
			intervalFrequency: "3M",
			matched:           true,
		},
		{
			name:              "valid interval/frequency string (1y)",
			intervalFrequency: "1y",
			matched:           true,
		},
		{
			name:              "invalid interval/frequency string (5m)",
			intervalFrequency: "5m",
			matched:           false,
		},
		{
			name:              "invalid interval/frequency string (-1w)",
			intervalFrequency: "-1w",
			matched:           false,
		},
		{
			name:              "invalid interval/frequency string (invalid)",
			intervalFrequency: "invalid",
			matched:           false,
		},
		{
			name:              "invalid interval/frequency empty string",
			intervalFrequency: "  ",
			matched:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, _ := StringMatchesIntervalFrequencyRegex(tt.intervalFrequency)
			if matched != tt.matched {
				t.Errorf("StringMatchesIntervalFrequencyRegex() failed match = %v, want %v", matched, tt.matched)
			}
		})
	}
}
