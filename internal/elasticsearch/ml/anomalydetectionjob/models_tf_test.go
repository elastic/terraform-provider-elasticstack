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

package anomalydetectionjob

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestIsEmptyJSONObject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value jsontypes.Normalized
		want  bool
	}{
		{
			name:  "null",
			value: jsontypes.NewNormalizedNull(),
			want:  false,
		},
		{
			name:  "unknown",
			value: jsontypes.NewNormalizedUnknown(),
			want:  false,
		},
		{
			name:  "empty object",
			value: jsontypes.NewNormalizedValue("{}"),
			want:  true,
		},
		{
			name:  "empty object with whitespace",
			value: jsontypes.NewNormalizedValue("{ }"),
			want:  true,
		},
		{
			name:  "non-empty object",
			value: jsontypes.NewNormalizedValue(`{"a":1}`),
			want:  false,
		},
		{
			name:  "invalid JSON",
			value: jsontypes.NewNormalizedValue("not-json"),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, isEmptyJSONObject(tt.value))
		})
	}
}

func TestFromAPIModel_customSettingsHandsOff(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name           string
		prior          jsontypes.Normalized
		api            map[string]any
		wantNull       bool
		wantValue      string
		wantJSONEquals string
	}{
		{
			name:     "null prior stays null when API has values",
			prior:    jsontypes.NewNormalizedNull(),
			api:      map[string]any{"created_by": "advanced-wizard"},
			wantNull: true,
		},
		{
			name:     "null prior stays null when API is nil",
			prior:    jsontypes.NewNormalizedNull(),
			api:      nil,
			wantNull: true,
		},
		{
			name:      "empty-object prior stays empty when API has values",
			prior:     jsontypes.NewNormalizedValue("{}"),
			api:       map[string]any{"created_by": "advanced-wizard"},
			wantValue: "{}",
		},
		{
			name:      "empty-object prior stays empty when API is nil",
			prior:     jsontypes.NewNormalizedValue("{}"),
			api:       nil,
			wantValue: "{}",
		},
		{
			name:      "whitespace empty-object prior normalizes to {}",
			prior:     jsontypes.NewNormalizedValue("{ }"),
			api:       map[string]any{"created_by": "advanced-wizard"},
			wantValue: "{}",
		},
		{
			name:           "owned object refreshes from API",
			prior:          jsontypes.NewNormalizedValue(`{"department":"ops"}`),
			api:            map[string]any{"department": "ops", "created_by": "advanced-wizard"},
			wantJSONEquals: `{"created_by":"advanced-wizard","department":"ops"}`,
		},
		{
			name:      "owned object falls back to empty when API is nil",
			prior:     jsontypes.NewNormalizedValue(`{"department":"ops"}`),
			api:       nil,
			wantValue: "{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			plan := TFModel{
				CustomSettings: tt.prior,
				Groups:         types.SetNull(types.StringType),
			}
			apiModel := &APIModel{
				JobID:          "job-unit-test",
				JobType:        "anomaly_detector",
				CustomSettings: tt.api,
			}

			diags := plan.fromAPIModel(ctx, apiModel)
			require.False(t, diags.HasError(), "%v", diags)

			if tt.wantNull {
				require.True(t, plan.CustomSettings.IsNull())
				return
			}
			require.False(t, plan.CustomSettings.IsNull())
			if tt.wantValue != "" {
				require.Equal(t, tt.wantValue, plan.CustomSettings.ValueString())
			}
			if tt.wantJSONEquals != "" {
				require.JSONEq(t, tt.wantJSONEquals, plan.CustomSettings.ValueString())
			}
		})
	}
}
