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

package calendar

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalendarTFModel_toAPICreateModel(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		model          CalendarTFModel
		expectedAPI    *CalendarCreateAPIModel
		expectedJobIDs []string
	}{
		{
			name: "with description and job_ids",
			model: CalendarTFModel{
				Description: types.StringValue("test calendar"),
				JobIDs:      mustStringSet(ctx, t, []string{"job1", "job2"}),
			},
			expectedAPI: &CalendarCreateAPIModel{
				Description: "test calendar",
				JobIDs:      []string{"job1", "job2"},
			},
		},
		{
			name: "with description only",
			model: CalendarTFModel{
				Description: types.StringValue("just a description"),
				JobIDs:      types.SetNull(types.StringType),
			},
			expectedAPI: &CalendarCreateAPIModel{
				Description: "just a description",
				JobIDs:      nil,
			},
		},
		{
			name: "with null description and null job_ids",
			model: CalendarTFModel{
				Description: types.StringNull(),
				JobIDs:      types.SetNull(types.StringType),
			},
			expectedAPI: &CalendarCreateAPIModel{
				Description: "",
				JobIDs:      nil,
			},
		},
		{
			name: "with empty job_ids set",
			model: CalendarTFModel{
				Description: types.StringValue("empty jobs"),
				JobIDs:      mustStringSet(ctx, t, []string{}),
			},
			expectedAPI: &CalendarCreateAPIModel{
				Description: "empty jobs",
				JobIDs:      []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, diags := tt.model.toAPICreateModel(ctx)
			require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
			assert.Equal(t, tt.expectedAPI.Description, result.Description)
			assert.Equal(t, tt.expectedAPI.JobIDs, result.JobIDs)
		})
	}
}

func TestCalendarTFModel_fromAPIModel(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name               string
		initialJobIDs      types.Set
		apiModel           *CalendarAPIModel
		expectedCalendarID string
		expectedDesc       types.String
		expectJobIDsNull   bool
		expectedJobIDs     []string
	}{
		{
			name:          "full API response",
			initialJobIDs: mustStringSet(ctx, t, []string{"old-job"}),
			apiModel: &CalendarAPIModel{
				CalendarID:  "my-calendar",
				Description: "A test calendar",
				JobIDs:      []string{"job1", "job2"},
			},
			expectedCalendarID: "my-calendar",
			expectedDesc:       types.StringValue("A test calendar"),
			expectedJobIDs:     []string{"job1", "job2"},
		},
		{
			name:          "empty description from API becomes null",
			initialJobIDs: types.SetNull(types.StringType),
			apiModel: &CalendarAPIModel{
				CalendarID:  "my-calendar",
				Description: "",
				JobIDs:      []string{},
			},
			expectedCalendarID: "my-calendar",
			expectedDesc:       types.StringNull(),
			expectJobIDsNull:   true,
		},
		{
			name:          "empty job_ids from API with non-null TF state becomes empty set",
			initialJobIDs: mustStringSet(ctx, t, []string{}),
			apiModel: &CalendarAPIModel{
				CalendarID: "my-calendar",
				JobIDs:     []string{},
			},
			expectedCalendarID: "my-calendar",
			expectedDesc:       types.StringNull(),
			expectedJobIDs:     []string{},
		},
		{
			name:          "empty job_ids from API with null TF state stays null",
			initialJobIDs: types.SetNull(types.StringType),
			apiModel: &CalendarAPIModel{
				CalendarID: "my-calendar",
				JobIDs:     []string{},
			},
			expectedCalendarID: "my-calendar",
			expectedDesc:       types.StringNull(),
			expectJobIDsNull:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &CalendarTFModel{
				JobIDs: tt.initialJobIDs,
			}

			diags := model.fromAPIModel(ctx, tt.apiModel)
			require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

			assert.Equal(t, tt.expectedCalendarID, model.CalendarID.ValueString())
			assert.Equal(t, tt.expectedDesc, model.Description)

			if tt.expectJobIDsNull {
				assert.True(t, model.JobIDs.IsNull(), "expected JobIDs to be null")
			} else {
				var jobIDs []string
				diags = model.JobIDs.ElementsAs(ctx, &jobIDs, false)
				require.False(t, diags.HasError())
				assert.ElementsMatch(t, tt.expectedJobIDs, jobIDs)
			}
		})
	}
}

func mustStringSet(ctx context.Context, t *testing.T, vals []string) types.Set {
	t.Helper()
	s, diags := types.SetValueFrom(ctx, types.StringType, vals)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	return s
}
