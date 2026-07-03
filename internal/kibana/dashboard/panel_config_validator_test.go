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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/discoversession"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/image"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/markdown"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/mlanomalyswimlane"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/mlsinglemetricviewer"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/sloalerts"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/sloburnrate"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/sloerrorbudget"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/slooverview"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/visconfig"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

var testAttrPathPanel = path.Root("panel")

// appendValidatePanelDiagnostics mirrors panelConfigValidator.ValidateObject semantics for isolated tests.
func appendValidatePanelDiagnostics(ctx context.Context, panel string, attrs map[string]attr.Value) diag.Diagnostics {
	var diags diag.Diagnostics
	if h := LookupHandler(panel); h != nil {
		diags.Append(h.ValidatePanelConfig(ctx, attrs, testAttrPathPanel)...)
	}
	return diags
}

func Test_markdownHandler_ValidatePanelConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("accepts markdown_config set", func(t *testing.T) {
		diags := markdown.Handler{}.ValidatePanelConfig(ctx, map[string]attr.Value{
			"markdown_config": types.BoolValue(true),
			"config_json":     types.StringNull(),
		}, testAttrPathPanel)
		require.False(t, diags.HasError())
	})

	t.Run("accepts config_json fallback", func(t *testing.T) {
		diags := markdown.Handler{}.ValidatePanelConfig(ctx, map[string]attr.Value{
			"config_json": types.StringValue(`{}`),
		}, testAttrPathPanel)
		require.False(t, diags.HasError())
	})

	t.Run("rejects missing config", func(t *testing.T) {
		diags := markdown.Handler{}.ValidatePanelConfig(ctx, map[string]attr.Value{}, testAttrPathPanel)
		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		require.Equal(t, "Missing markdown panel configuration", diags[0].Summary())
	})

	t.Run("rejects markdown_config combined with panel config_json", func(t *testing.T) {
		diags := markdown.Handler{}.ValidatePanelConfig(ctx, map[string]attr.Value{
			"markdown_config": types.BoolValue(true),
			"config_json":     types.StringValue(`{}`),
		}, testAttrPathPanel)
		require.True(t, diags.HasError())
	})
}

func Test_visHandler_ValidatePanelConfig(t *testing.T) {
	ctx := context.Background()
	h := visconfig.Handler{}

	t.Run("accepts config_json fallback", func(t *testing.T) {
		diags := h.ValidatePanelConfig(ctx, map[string]attr.Value{
			"config_json": types.StringValue(`{}`),
		}, testAttrPathPanel)
		require.False(t, diags.HasError())
	})

	t.Run("accepts vis_config as sole vis configuration", func(t *testing.T) {
		diags := h.ValidatePanelConfig(ctx, map[string]attr.Value{
			"vis_config": types.BoolValue(true),
		}, testAttrPathPanel)
		require.False(t, diags.HasError())
	})

	t.Run("rejects missing config", func(t *testing.T) {
		diags := h.ValidatePanelConfig(ctx, map[string]attr.Value{}, testAttrPathPanel)
		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		require.Equal(t, "Missing vis panel configuration", diags[0].Summary())
	})

	t.Run("rejects vis_config and config_json together", func(t *testing.T) {
		diags := h.ValidatePanelConfig(ctx, map[string]attr.Value{
			"config_json": types.StringValue(`{}`),
			"vis_config":  types.BoolValue(true),
		}, testAttrPathPanel)
		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		require.Equal(t, "Invalid vis panel configuration", diags[0].Summary())
		require.Contains(t, diags[0].Detail(), "vis_config")
	})

	t.Run("defers when config_json is unknown", func(t *testing.T) {
		diags := h.ValidatePanelConfig(ctx, map[string]attr.Value{
			"config_json": types.StringUnknown(),
		}, testAttrPathPanel)
		require.False(t, diags.HasError())
	})

	t.Run("defers when vis_config is unknown", func(t *testing.T) {
		diags := h.ValidatePanelConfig(ctx, map[string]attr.Value{
			"vis_config": types.BoolUnknown(),
		}, testAttrPathPanel)
		require.False(t, diags.HasError())
	})
}

func Test_registryDispatch_markdown_badConfig(t *testing.T) {
	ctx := context.Background()
	diags := appendValidatePanelDiagnostics(ctx, "markdown", map[string]attr.Value{})
	require.True(t, diags.HasError())
	var sawMarkdown bool
	for _, d := range diags {
		if d.Summary() == "Missing markdown panel configuration" {
			sawMarkdown = true
			break
		}
	}
	require.True(t, sawMarkdown)
}

func Test_panelConfigValidateDiags_timeSlider(t *testing.T) {
	ctx := context.Background()

	t.Run("accepts no config blocks", func(t *testing.T) {
		diags := appendValidatePanelDiagnostics(ctx, "time_slider_control", map[string]attr.Value{})
		require.False(t, diags.HasError())
	})

	t.Run("does not emit diagnostic for practitioner-authored config_json", func(t *testing.T) {
		diags := appendValidatePanelDiagnostics(ctx, "time_slider_control", map[string]attr.Value{
			"config_json": types.StringValue(`{}`),
		})
		require.False(t, diags.HasError())
	})

	t.Run("accepts time_slider when config_json is unknown", func(t *testing.T) {
		diags := appendValidatePanelDiagnostics(ctx, "time_slider_control", map[string]attr.Value{
			"config_json": types.StringUnknown(),
		})
		require.False(t, diags.HasError())
	})
}

func Test_mlAnomalySwimlaneHandler_ValidatePanelConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("accepts ml_anomaly_swimlane_config via flat attrs", func(t *testing.T) {
		diags := mlanomalyswimlane.Handler{}.ValidatePanelConfig(ctx, map[string]attr.Value{
			"swimlane_type": types.StringValue("overall"),
			"job_ids":       types.ListValueMust(types.StringType, []attr.Value{types.StringValue("job-a")}),
		}, testAttrPathPanel)
		require.False(t, diags.HasError())
	})

	t.Run("rejects missing config block on full-panel attrs shape", func(t *testing.T) {
		diags := mlanomalyswimlane.Handler{}.ValidatePanelConfig(ctx, map[string]attr.Value{}, testAttrPathPanel)
		require.True(t, diags.HasError())
		require.Equal(t, "Missing ML anomaly swim lane panel configuration", diags[0].Summary())
	})

	t.Run("rejects empty job_ids list", func(t *testing.T) {
		diags := mlanomalyswimlane.Handler{}.ValidatePanelConfig(ctx, map[string]attr.Value{
			"swimlane_type": types.StringValue("overall"),
			"job_ids":       types.ListValueMust(types.StringType, []attr.Value{}),
		}, testAttrPathPanel)
		require.True(t, diags.HasError())
		require.Contains(t, diags[0].Summary(), "Invalid ML anomaly swim lane configuration")
	})

	t.Run("rejects null job_ids list", func(t *testing.T) {
		diags := mlanomalyswimlane.Handler{}.ValidatePanelConfig(ctx, map[string]attr.Value{
			"swimlane_type": types.StringValue("overall"),
			"job_ids":       types.ListNull(types.StringType),
		}, testAttrPathPanel)
		require.True(t, diags.HasError())
		require.Contains(t, diags[0].Summary(), "Invalid ML anomaly swim lane configuration")
	})
}

func Test_mlSingleMetricViewerHandler_ValidatePanelConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("accepts ml_single_metric_viewer_config via flat attrs", func(t *testing.T) {
		diags := mlsinglemetricviewer.Handler{}.ValidatePanelConfig(ctx, map[string]attr.Value{
			"job_ids": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("job-a")}),
		}, testAttrPathPanel)
		require.False(t, diags.HasError())
	})

	t.Run("rejects missing config block on full-panel attrs shape", func(t *testing.T) {
		diags := mlsinglemetricviewer.Handler{}.ValidatePanelConfig(ctx, map[string]attr.Value{}, testAttrPathPanel)
		require.True(t, diags.HasError())
		require.Equal(t, "Missing ML single metric viewer panel configuration", diags[0].Summary())
	})

	t.Run("rejects empty job_ids list", func(t *testing.T) {
		diags := mlsinglemetricviewer.Handler{}.ValidatePanelConfig(ctx, map[string]attr.Value{
			"job_ids": types.ListValueMust(types.StringType, []attr.Value{}),
		}, testAttrPathPanel)
		require.True(t, diags.HasError())
		require.Contains(t, diags[0].Summary(), "Invalid ML single metric viewer configuration")
	})

	t.Run("rejects null job_ids list", func(t *testing.T) {
		diags := mlsinglemetricviewer.Handler{}.ValidatePanelConfig(ctx, map[string]attr.Value{
			"job_ids": types.ListNull(types.StringType),
		}, testAttrPathPanel)
		require.True(t, diags.HasError())
		require.Contains(t, diags[0].Summary(), "Invalid ML single metric viewer configuration")
	})

	t.Run("rejects more than one job_ids entry", func(t *testing.T) {
		diags := mlsinglemetricviewer.Handler{}.ValidatePanelConfig(ctx, map[string]attr.Value{
			"job_ids": types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("job-a"),
				types.StringValue("job-b"),
			}),
		}, testAttrPathPanel)
		require.True(t, diags.HasError())
		require.Contains(t, diags[0].Summary(), "Invalid ML single metric viewer configuration")
	})
}

func Test_sloBurn_rateHandler_ValidatePanelConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("accepts slo_burn_rate_config via flat attrs", func(t *testing.T) {
		diags := sloburnrate.Handler{}.ValidatePanelConfig(ctx, map[string]attr.Value{
			"slo_id":   types.StringValue("foo"),
			"duration": types.StringValue("5m"),
		}, testAttrPathPanel)
		require.False(t, diags.HasError())
	})

	t.Run("rejects missing config block on full-panel attrs shape", func(t *testing.T) {
		diags := sloburnrate.Handler{}.ValidatePanelConfig(ctx, map[string]attr.Value{}, testAttrPathPanel)
		require.True(t, diags.HasError())
		require.Equal(t, "Missing SLO burn rate panel configuration", diags[0].Summary())
	})
}

func Test_rangeSliderControlValueListSizeValidator(t *testing.T) {
	panelSchema := getPanelSchema()
	rsAttr, ok := panelSchema.Attributes["range_slider_control_config"].(schema.SingleNestedAttribute)
	require.True(t, ok)

	byFieldAttr, ok := rsAttr.Attributes["by_field"].(schema.SingleNestedAttribute)
	require.True(t, ok)

	valueAttr, ok := byFieldAttr.Attributes["value"].(schema.ListAttribute)
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
			req := validator.ListRequest{
				Path:           path.Root("value"),
				PathExpression: path.MatchRoot("value"),
				ConfigValue:    tc.value,
			}
			resp := validator.ListResponse{}

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

func Test_slo_error_budget_ValidatePanelConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("accepts slo_error_budget_config via flat slo_id", func(t *testing.T) {
		diags := sloerrorbudget.Handler{}.ValidatePanelConfig(ctx, map[string]attr.Value{
			"slo_id": types.StringValue("sid"),
		}, testAttrPathPanel)
		require.False(t, diags.HasError())
	})

	t.Run("rejects missing slo_error_budget_config", func(t *testing.T) {
		diags := sloerrorbudget.Handler{}.ValidatePanelConfig(ctx, map[string]attr.Value{}, testAttrPathPanel)
		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		require.Equal(t, "Missing slo_error_budget panel configuration", diags[0].Summary())
	})
}

func Test_getSloErrorBudgetSchema_drilldownsHardcodeAPIConstants(t *testing.T) {
	sna, ok := sloerrorbudget.SchemaAttribute().(schema.SingleNestedAttribute)
	require.True(t, ok)
	drilldownsAttr, ok := sna.Attributes["drilldowns"].(schema.ListNestedAttribute)
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
	ctx := context.Background()

	t.Run("accepts no config blocks", func(t *testing.T) {
		diags := appendValidatePanelDiagnostics(ctx, "options_list_control", map[string]attr.Value{})
		require.False(t, diags.HasError())
	})

	t.Run("does not emit diagnostic for practitioner-authored config_json", func(t *testing.T) {
		diags := appendValidatePanelDiagnostics(ctx, "options_list_control", map[string]attr.Value{
			"config_json": types.StringValue(`{}`),
		})
		require.False(t, diags.HasError())
	})

	t.Run("accepts options_list_control when config_json is unknown", func(t *testing.T) {
		diags := appendValidatePanelDiagnostics(ctx, "options_list_control", map[string]attr.Value{
			"config_json": types.StringUnknown(),
		})
		require.False(t, diags.HasError())
	})
}

func Test_optionsListControlSearchTechniqueValidator(t *testing.T) {
	panelSchema := getPanelSchema()
	olAttr, ok := panelSchema.Attributes["options_list_control_config"].(schema.SingleNestedAttribute)
	require.True(t, ok)

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

	// search_technique is duplicated (identically validated) inside both the by_field and
	// by_esql branches of options_list_control_config.
	for _, branch := range []string{"by_field", "by_esql"} {
		branchAttr, ok := olAttr.Attributes[branch].(schema.SingleNestedAttribute)
		require.True(t, ok, "branch %q", branch)

		searchTechAttr, ok := branchAttr.Attributes["search_technique"].(schema.StringAttribute)
		require.True(t, ok, "branch %q", branch)
		require.NotEmpty(t, searchTechAttr.Validators, "branch %q", branch)

		for _, tc := range testCases {
			t.Run(branch+"/"+tc.name, func(t *testing.T) {
				req := validator.StringRequest{
					Path:           path.Root("search_technique"),
					PathExpression: path.MatchRoot("search_technique"),
					ConfigValue:    types.StringValue(tc.value),
				}
				resp := validator.StringResponse{}

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

			req := validator.Float32Request{
				Path:           path.Root(tc.attrName),
				PathExpression: path.MatchRoot(tc.attrName),
				ConfigValue:    types.Float32Value(float32(tc.value)),
			}
			resp := validator.Float32Response{}

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

func Test_imageHandler_ValidatePanelConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("accepts image_config", func(t *testing.T) {
		diags := image.Handler{}.ValidatePanelConfig(ctx, map[string]attr.Value{
			"image_config": types.BoolValue(true),
		}, testAttrPathPanel)
		require.False(t, diags.HasError())
	})

	t.Run("rejects missing image_config", func(t *testing.T) {
		diags := image.Handler{}.ValidatePanelConfig(ctx, map[string]attr.Value{}, testAttrPathPanel)
		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		require.Equal(t, "Missing image panel configuration", diags[0].Summary())
	})
}

func Test_slo_alerts_ValidatePanelConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("accepts slo_alerts_config", func(t *testing.T) {
		diags := sloalerts.Handler{}.ValidatePanelConfig(ctx, map[string]attr.Value{
			"slo_alerts_config": types.BoolValue(true),
		}, testAttrPathPanel)
		require.False(t, diags.HasError())
	})

	t.Run("rejects missing slo_alerts_config", func(t *testing.T) {
		diags := sloalerts.Handler{}.ValidatePanelConfig(ctx, map[string]attr.Value{}, testAttrPathPanel)
		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		require.Equal(t, "Missing SLO alerts panel configuration", diags[0].Summary())
	})
}

func Test_discoverSessionHandler_ValidatePanelConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("accepts discover_session_config", func(t *testing.T) {
		diags := discoversession.Handler{}.ValidatePanelConfig(ctx, map[string]attr.Value{
			"discover_session_config": types.BoolValue(true),
		}, testAttrPathPanel)
		require.False(t, diags.HasError())
	})

	t.Run("rejects missing discover_session_config", func(t *testing.T) {
		diags := discoversession.Handler{}.ValidatePanelConfig(ctx, map[string]attr.Value{}, testAttrPathPanel)
		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		require.Equal(t, "Missing discover_session panel configuration", diags[0].Summary())
	})
}

func Test_slo_overview_ValidatePanelConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("accepts slo_overview_config", func(t *testing.T) {
		diags := slooverview.Handler{}.ValidatePanelConfig(ctx, map[string]attr.Value{
			"slo_overview_config": types.BoolValue(true),
		}, testAttrPathPanel)
		require.False(t, diags.HasError())
	})

	t.Run("missing", func(t *testing.T) {
		diags := slooverview.Handler{}.ValidatePanelConfig(ctx, map[string]attr.Value{}, testAttrPathPanel)
		require.True(t, diags.HasError())
		require.Len(t, diags, 1)
		require.Equal(t, "Missing SLO overview panel configuration", diags[0].Summary())
	})
}

func Test_panelConfigValidator_rejectsRemovedLensDashboardApp(t *testing.T) {
	ctx := context.Background()
	v := panelConfigValidator{}
	req := validator.ObjectRequest{
		ConfigValue: types.ObjectValueMust(map[string]attr.Type{
			"type": types.StringType,
		}, map[string]attr.Value{
			"type": types.StringValue("lens-dashboard-app"),
		}),
		Path: testAttrPathPanel,
	}
	var resp validator.ObjectResponse
	v.ValidateObject(ctx, req, &resp)
	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Removed panel type")
	require.Contains(t, resp.Diagnostics.Errors()[0].Detail(),
		"https://registry.terraform.io/providers/elastic/elasticstack/latest/docs/guides/elasticstack-kibana-dashboard-remove-lens-dashboard-app")
}
