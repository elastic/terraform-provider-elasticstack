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

func TestStringMatchesOnWeekDay(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		onWeekDay string
		matched   bool
	}{
		{
			name:      "valid on_week_day string (+1MO)",
			onWeekDay: "+1MO",
			matched:   true,
		},
		{
			name:      "valid on_week_day string (+2TU)",
			onWeekDay: "+2TU",
			matched:   true,
		},
		{
			name:      "valid on_week_day string (+3WE)",
			onWeekDay: "+3WE",
			matched:   true,
		},
		{
			name:      "valid on_week_day string (+4TH)",
			onWeekDay: "+4TH",
			matched:   true,
		},
		{
			name:      "valid on_week_day string (+5FR)",
			onWeekDay: "+5FR",
			matched:   true,
		},
		{
			name:      "valid on_week_day string (-5SA)",
			onWeekDay: "-5SA",
			matched:   true,
		},
		{
			name:      "valid on_week_day string (-4SU)",
			onWeekDay: "-4SU",
			matched:   true,
		},
		{
			name:      "valid on_week_day string (-3MO)",
			onWeekDay: "-3MO",
			matched:   true,
		},
		{
			name:      "valid on_week_day string (-2TU)",
			onWeekDay: "-2TU",
			matched:   true,
		},
		{
			name:      "valid on_week_day string (-1WE)",
			onWeekDay: "-1WE",
			matched:   true,
		},
		{
			name:      "invalid on_week_day unit (FOOBAR)",
			onWeekDay: "FOOBAR",
			matched:   false,
		},
		{
			name:      "invalid on_week_day string (+9MO)",
			onWeekDay: "+9MO",
			matched:   false,
		},
		{
			name:      "invalid on_week_day string (-7FR)",
			onWeekDay: "-7FR",
			matched:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, _ := StringMatchesOnWeekDayRegex(tt.onWeekDay)
			if matched != tt.matched {
				t.Errorf("StringMatchesOnWeekDayRegex() failed match = %v, want %v", matched, tt.matched)
			}
		})
	}
}
