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

	"github.com/hashicorp/terraform-plugin-framework/attr"
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
			panelConfigValueState{},
			lensConfigStates(nil),
			panelConfigValueState{},
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
			panelConfigValueState{},
			lensConfigStates(nil),
			panelConfigValueState{},
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
			panelConfigValueState{},
			lensConfigStates(nil),
			panelConfigValueState{},
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
			panelConfigValueState{},
			lensConfigStates(map[string]panelConfigValueState{
				"xy_chart_config": {Set: true},
			}),
			panelConfigValueState{},
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
			panelConfigValueState{},
			lensConfigStates(nil),
			panelConfigValueState{},
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
			panelConfigValueState{},
			lensConfigStates(nil),
			panelConfigValueState{},
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
			panelConfigValueState{},
			lensConfigStates(map[string]panelConfigValueState{
				"xy_chart_config": {Set: true},
				"heatmap_config":  {Set: true},
			}),
			panelConfigValueState{},
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
			panelConfigValueState{},
			lensConfigStates(map[string]panelConfigValueState{
				"gauge_config": {Set: true},
			}),
			panelConfigValueState{},
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
			panelConfigValueState{},
			lensConfigStates(nil),
			panelConfigValueState{},
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
			panelConfigValueState{},
			lensConfigStates(nil),
			panelConfigValueState{},
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
			panelConfigValueState{},
			lensConfigStates(nil),
			panelConfigValueState{},
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
			panelConfigValueState{},
			lensConfigStates(nil),
			panelConfigValueState{},
			nil,
		)
		require.False(t, diags.HasError())
	})
}

func Test_panelConfigValidateDiags_SloBurnRate(t *testing.T) {
	t.Run("accepts slo_burn_rate_config", func(t *testing.T) {
		diags := panelConfigValidateDiags(
			"slo_burn_rate",
			panelConfigValueState{},
			panelConfigValueState{},
			panelConfigValueState{Set: true},
			panelConfigValueState{},
			lensConfigStates(nil),
			panelConfigValueState{},
			nil,
		)
		require.False(t, diags.HasError())
	})

	t.Run("rejects missing config", func(t *testing.T) {
		diags := panelConfigValidateDiags(
			"slo_burn_rate",
			panelConfigValueState{},
			panelConfigValueState{},
			panelConfigValueState{},
			panelConfigValueState{},
			lensConfigStates(nil),
			panelConfigValueState{},
			nil,
		)
		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		require.Equal(t, "Missing SLO burn rate panel configuration", diags[0].Summary())
	})

	t.Run("defers when config is unknown", func(t *testing.T) {
		diags := panelConfigValidateDiags(
			"slo_burn_rate",
			panelConfigValueState{},
			panelConfigValueState{},
			panelConfigValueState{Unknown: true},
			panelConfigValueState{},
			lensConfigStates(nil),
			panelConfigValueState{},
			nil,
		)
		require.False(t, diags.HasError())
	})
}

func Test_rangeSliderControlValueListSizeValidator(t *testing.T) {
	panelSchema := getPanelSchema()
	rsAttr, ok := panelSchema.Attributes["range_slider_control_config"].(schema.SingleNestedAttribute)
	require.True(t, ok)

	valueAttr, ok := rsAttr.Attributes["value"].(schema.ListAttribute)
	require.True(t, ok)
	require.NotEmpty(t, valueAttr.Validators)

	testCases := []struct {
		name      string
		value     types.List
		expectErr bool
	}{
		{
			name:      "rejects empty list",
			value:     types.ListValueMust(types.StringType, []attr.Value{}),
			expectErr: true,
		},
		{
			name:      "rejects single element",
			value:     types.ListValueMust(types.StringType, []attr.Value{types.StringValue("10")}),
			expectErr: true,
		},
		{
			name:      "accepts exactly two elements",
			value:     types.ListValueMust(types.StringType, []attr.Value{types.StringValue("10"), types.StringValue("500")}),
			expectErr: false,
		},
		{
			name: "rejects three elements",
			value: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("10"),
				types.StringValue("200"),
				types.StringValue("500"),
			}),
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tfvalidator.ListRequest{
				Path:           path.Root("value"),
				PathExpression: path.MatchRoot("value"),
				ConfigValue:    tc.value,
			}
			resp := tfvalidator.ListResponse{}

			for _, v := range valueAttr.Validators {
				v.ValidateList(context.Background(), req, &resp)
			}

			if tc.expectErr {
				require.True(t, resp.Diagnostics.HasError(), "expected error but got none")
			} else {
				require.False(t, resp.Diagnostics.HasError(), "unexpected error: %s", resp.Diagnostics)
			}
		})
	}
}

func Test_panelConfigValidateDiags_sloErrorBudget(t *testing.T) {
	t.Run("accepts slo_error_budget_config", func(t *testing.T) {
		diags := panelConfigValidateDiags(
			"slo_error_budget",
			panelConfigValueState{},
			panelConfigValueState{},
			panelConfigValueState{},
			panelConfigValueState{Set: true},
			lensConfigStates(nil),
			panelConfigValueState{},
			nil,
		)
		require.False(t, diags.HasError())
	})

	t.Run("defers when slo_error_budget_config is unknown", func(t *testing.T) {
		diags := panelConfigValidateDiags(
			"slo_error_budget",
			panelConfigValueState{},
			panelConfigValueState{},
			panelConfigValueState{},
			panelConfigValueState{Unknown: true},
			lensConfigStates(nil),
			panelConfigValueState{},
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
			panelConfigValueState{},
			lensConfigStates(nil),
			panelConfigValueState{},
			nil,
		)
		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		require.Equal(t, "Missing slo_error_budget panel configuration", diags[0].Summary())
	})
}

func Test_getSloErrorBudgetSchema_drilldownsHardcodeAPIConstants(t *testing.T) {
	drilldownsAttr, ok := getSloErrorBudgetSchema()["drilldowns"].(schema.ListNestedAttribute)
	require.True(t, ok)

	attrs := drilldownsAttr.NestedObject.Attributes
	require.Contains(t, attrs, "url")
	require.Contains(t, attrs, "label")
	require.Contains(t, attrs, "encode_url")
	require.Contains(t, attrs, "open_in_new_tab")
	require.NotContains(t, attrs, "trigger")
	require.NotContains(t, attrs, "type")
}

func Test_panelConfigValidateDiags_optionsListControl(t *testing.T) {
	t.Run("accepts no config blocks", func(t *testing.T) {
		diags := panelConfigValidateDiags(
			"options_list_control",
			panelConfigValueState{},
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
			"options_list_control",
			panelConfigValueState{},
			panelConfigValueState{Set: true},
			panelConfigValueState{},
			panelConfigValueState{},
			lensConfigStates(nil),
			nil,
		)
		require.False(t, diags.HasError())
	})

	t.Run("accepts options_list_control when config_json is unknown", func(t *testing.T) {
		diags := panelConfigValidateDiags(
			"options_list_control",
			panelConfigValueState{},
			panelConfigValueState{Unknown: true},
			panelConfigValueState{},
			panelConfigValueState{},
			lensConfigStates(nil),
			nil,
		)
		require.False(t, diags.HasError())
	})
}

func Test_optionsListControlSearchTechniqueValidator(t *testing.T) {
	panelSchema := getPanelSchema()
	olAttr, ok := panelSchema.Attributes["options_list_control_config"].(schema.SingleNestedAttribute)
	require.True(t, ok)

	searchTechAttr, ok := olAttr.Attributes["search_technique"].(schema.StringAttribute)
	require.True(t, ok)
	require.NotEmpty(t, searchTechAttr.Validators)

	testCases := []struct {
		name      string
		value     string
		expectErr bool
	}{
		{name: "accepts prefix", value: "prefix", expectErr: false},
		{name: "accepts wildcard", value: "wildcard", expectErr: false},
		{name: "accepts exact", value: "exact", expectErr: false},
		{name: "rejects fuzzy", value: "fuzzy", expectErr: true},
		{name: "rejects empty string", value: "", expectErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tfvalidator.StringRequest{
				Path:           path.Root("search_technique"),
				PathExpression: path.MatchRoot("search_technique"),
				ConfigValue:    types.StringValue(tc.value),
			}
			resp := tfvalidator.StringResponse{}

			for _, v := range searchTechAttr.Validators {
				v.ValidateString(context.Background(), req, &resp)
			}

			if tc.expectErr {
				require.True(t, resp.Diagnostics.HasError())
			} else {
				require.False(t, resp.Diagnostics.HasError())
			}
		})
	}
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
