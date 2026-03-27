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

package dashboard

import (
	"context"
	"maps"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	tfvalidator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func lensConfigStates(overrides map[string]panelConfigValueState) map[string]panelConfigValueState {
	states := make(map[string]panelConfigValueState, len(lensPanelConfigNames))
	for _, name := range lensPanelConfigNames {
		states[name] = panelConfigValueState{}
	}
	maps.Copy(states, overrides)
	return states
}

func Test_panelConfigValidateDiags_markdown(t *testing.T) {
	t.Run("accepts markdown_config", func(t *testing.T) {
		diags := panelConfigValidateDiags(
			"markdown",
			panelConfigValueState{Set: true},
			panelConfigValueState{},
			panelConfigValueState{},
			lensConfigStates(nil),
			nil,
		)
		require.False(t, diags.HasError())
	})

	t.Run("accepts config_json fallback", func(t *testing.T) {
		diags := panelConfigValidateDiags(
			"markdown",
			panelConfigValueState{},
			panelConfigValueState{Set: true},
			panelConfigValueState{},
			lensConfigStates(nil),
			nil,
		)
		require.False(t, diags.HasError())
	})

	t.Run("rejects missing config", func(t *testing.T) {
		diags := panelConfigValidateDiags(
			"markdown",
			panelConfigValueState{},
			panelConfigValueState{},
			panelConfigValueState{},
			lensConfigStates(nil),
			nil,
		)
		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		require.Equal(t, "Missing markdown panel configuration", diags[0].Summary())
	})
}

func Test_panelConfigValidateDiags_lens(t *testing.T) {
	t.Run("accepts one typed config", func(t *testing.T) {
		diags := panelConfigValidateDiags(
			"lens",
			panelConfigValueState{},
			panelConfigValueState{},
			panelConfigValueState{},
			lensConfigStates(map[string]panelConfigValueState{
				"xy_chart_config": {Set: true},
			}),
			nil,
		)
		require.False(t, diags.HasError())
	})

	t.Run("accepts config_json fallback", func(t *testing.T) {
		diags := panelConfigValidateDiags(
			"lens",
			panelConfigValueState{},
			panelConfigValueState{Set: true},
			panelConfigValueState{},
			lensConfigStates(nil),
			nil,
		)
		require.False(t, diags.HasError())
	})

	t.Run("rejects missing config", func(t *testing.T) {
		diags := panelConfigValidateDiags(
			"lens",
			panelConfigValueState{},
			panelConfigValueState{},
			panelConfigValueState{},
			lensConfigStates(nil),
			nil,
		)
		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		require.Equal(t, "Missing lens panel configuration", diags[0].Summary())
	})

	t.Run("rejects multiple typed configs", func(t *testing.T) {
		diags := panelConfigValidateDiags(
			"lens",
			panelConfigValueState{},
			panelConfigValueState{},
			panelConfigValueState{},
			lensConfigStates(map[string]panelConfigValueState{
				"xy_chart_config": {Set: true},
				"heatmap_config":  {Set: true},
			}),
			nil,
		)
		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		require.Equal(t, "Invalid lens panel configuration", diags[0].Summary())
	})

	t.Run("rejects typed config plus config_json", func(t *testing.T) {
		diags := panelConfigValidateDiags(
			"lens",
			panelConfigValueState{},
			panelConfigValueState{Set: true},
			panelConfigValueState{},
			lensConfigStates(map[string]panelConfigValueState{
				"gauge_config": {Set: true},
			}),
			nil,
		)
		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		require.Equal(t, "Invalid lens panel configuration", diags[0].Summary())
	})

	t.Run("defers when config_json is unknown", func(t *testing.T) {
		diags := panelConfigValidateDiags(
			"lens",
			panelConfigValueState{},
			panelConfigValueState{Unknown: true},
			panelConfigValueState{},
			lensConfigStates(nil),
			nil,
		)
		require.False(t, diags.HasError())
	})
}

func Test_panelConfigValidateDiags_timeSlider(t *testing.T) {
	t.Run("accepts no config blocks", func(t *testing.T) {
		diags := panelConfigValidateDiags(
			"time_slider_control",
			panelConfigValueState{},
			panelConfigValueState{},
			panelConfigValueState{},
			lensConfigStates(nil),
			nil,
		)
		require.False(t, diags.HasError())
	})

	t.Run("does not emit diagnostic for practitioner-authored config_json", func(t *testing.T) {
		// Schema validation on `config_json` (type allowlist) produces the single plan-time error;
		// this object validator intentionally does not duplicate it.
		diags := panelConfigValidateDiags(
			"time_slider_control",
			panelConfigValueState{},
			panelConfigValueState{Set: true},
			panelConfigValueState{},
			lensConfigStates(nil),
			nil,
		)
		require.False(t, diags.HasError())
	})

	t.Run("accepts time_slider when config_json is unknown", func(t *testing.T) {
		diags := panelConfigValidateDiags(
			"time_slider_control",
			panelConfigValueState{},
			panelConfigValueState{Unknown: true},
			panelConfigValueState{},
			lensConfigStates(nil),
			nil,
		)
		require.False(t, diags.HasError())
	})
}

func Test_panelConfigValidateDiags_sloErrorBudget(t *testing.T) {
	t.Run("accepts slo_error_budget_config", func(t *testing.T) {
		diags := panelConfigValidateDiags(
			"slo_error_budget",
			panelConfigValueState{},
			panelConfigValueState{},
			panelConfigValueState{Set: true},
			lensConfigStates(nil),
			nil,
		)
		require.False(t, diags.HasError())
	})

	t.Run("defers when slo_error_budget_config is unknown", func(t *testing.T) {
		diags := panelConfigValidateDiags(
			"slo_error_budget",
			panelConfigValueState{},
			panelConfigValueState{},
			panelConfigValueState{Unknown: true},
			lensConfigStates(nil),
			nil,
		)
		require.False(t, diags.HasError())
	})

	t.Run("rejects missing slo_error_budget_config", func(t *testing.T) {
		diags := panelConfigValidateDiags(
			"slo_error_budget",
			panelConfigValueState{},
			panelConfigValueState{},
			panelConfigValueState{},
			lensConfigStates(nil),
			nil,
		)
		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		require.Equal(t, "Missing slo_error_budget panel configuration", diags[0].Summary())
	})
}

func Test_timeSliderControlPercentageValidators(t *testing.T) {
	panelSchema := getPanelSchema()
	timeSliderAttr, ok := panelSchema.Attributes["time_slider_control_config"].(schema.SingleNestedAttribute)
	require.True(t, ok)

	testCases := []struct {
		name      string
		attrName  string
		value     float64
		expectErr bool
	}{
		{
			name:      "start percentage above upper bound",
			attrName:  "start_percentage_of_time_range",
			value:     1.5,
			expectErr: true,
		},
		{
			name:      "end percentage below lower bound",
			attrName:  "end_percentage_of_time_range",
			value:     -0.1,
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			attr, ok := timeSliderAttr.Attributes[tc.attrName].(schema.Float32Attribute)
			require.True(t, ok)
			require.NotEmpty(t, attr.Validators)

			req := tfvalidator.Float32Request{
				Path:           path.Root(tc.attrName),
				PathExpression: path.MatchRoot(tc.attrName),
				ConfigValue:    types.Float32Value(float32(tc.value)),
			}
			resp := tfvalidator.Float32Response{}

			for _, v := range attr.Validators {
				v.ValidateFloat32(context.Background(), req, &resp)
			}

			if tc.expectErr {
				require.True(t, resp.Diagnostics.HasError())
				require.Len(t, resp.Diagnostics, 1)
				require.Contains(t, resp.Diagnostics[0].Detail(), "between 0.000000 and 1.000000")
				return
			}

			require.False(t, resp.Diagnostics.HasError())
		})
	}
}
