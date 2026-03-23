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

func TestStringMatchesAlertingDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration string
		matched  bool
	}{
		{
			name:     "valid Alerting duration string (30d)",
			duration: "30d",
			matched:  true,
		},
		{
			name:     "invalid Alerting duration unit (0s)",
			duration: "0s",
			matched:  false,
		},
		{
			name:     "invalid Alerting duration value (.12y)",
			duration: ".12y",
			matched:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, _ := StringMatchesAlertingDurationRegex(tt.duration)
			if matched != tt.matched {
				t.Errorf("StringMatchesAlertingDurationRegex() failed match = %v, want %v", matched, tt.matched)
			}
		})
	}
}
