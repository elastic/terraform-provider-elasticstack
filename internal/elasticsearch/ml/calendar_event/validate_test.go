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

package calendar_event

import (
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateEventEndAfterStart(t *testing.T) {
	t1 := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 1, 10, 14, 0, 0, 0, time.UTC)
	same := t1

	tests := []struct {
		name      string
		start     timetypes.RFC3339
		end       timetypes.RFC3339
		wantError bool
	}{
		{
			name:  "end after start",
			start: timetypes.NewRFC3339TimeValue(t1),
			end:   timetypes.NewRFC3339TimeValue(t2),
		},
		{
			name:      "end equal to start",
			start:     timetypes.NewRFC3339TimeValue(t1),
			end:       timetypes.NewRFC3339TimeValue(same),
			wantError: true,
		},
		{
			name:      "end before start",
			start:     timetypes.NewRFC3339TimeValue(t2),
			end:       timetypes.NewRFC3339TimeValue(t1),
			wantError: true,
		},
		{
			name:  "unknown start skipped",
			start: timetypes.NewRFC3339Unknown(),
			end:   timetypes.NewRFC3339TimeValue(t2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := validateEventEndAfterStart(tt.start, tt.end)
			if tt.wantError {
				require.True(t, diags.HasError())
				assert.Contains(t, diags.Errors()[0].Summary(), "Invalid event time range")
			} else {
				assert.False(t, diags.HasError())
			}
		})
	}
}
