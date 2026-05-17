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

package lensdashboardapp

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func byValueAttributes() map[string]schema.Attribute {
	out := map[string]schema.Attribute{
		"config_json": schema.StringAttribute{
			MarkdownDescription: "Optional raw normalized JSON for the by-value Lens chart `config` (full API shape, including chart `type` and `time_range` where the API requires them). " +
				"Use as the single `by_value` source, or use one supported typed chart block instead (not both). " +
				"Distinct from panel-level `config_json` on the panel.",
			Optional:   true,
			CustomType: jsontypes.NormalizedType{},
		},
	}
	for _, c := range lenscommon.All() {
		out[chartBlockKeyOrPanic(c.VizType())] = c.SchemaAttribute()
	}
	return out
}

// ByValueSourceAttrNames lists mutually exclusive `by_value` sources (`config_json` plus typed chart blocks).
func ByValueSourceAttrNames() []string {
	return lensByValueSourceAttrNames()
}

func lensByValueSourceAttrNames() []string {
	out := []string{"config_json"}
	for _, c := range lenscommon.All() {
		out = append(out, chartBlockKeyOrPanic(c.VizType()))
	}
	return out
}

func chartBlockKeyOrPanic(vizType string) string {
	key := lenscommon.TerraformChartBlockKey(vizType)
	if key == "" {
		panic("lensdashboardapp: missing terraform chart block key for VizType " + vizType)
	}
	return key
}

func innerSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"by_value": schema.SingleNestedAttribute{
			MarkdownDescription: "Inline by-value `lens-dashboard-app` configuration. " +
				"Set exactly one of `config_json` (raw JSON) or one supported typed Lens chart block, not both.",
			Optional:   true,
			Attributes: byValueAttributes(),
			Validators: []validator.Object{
				panelkit.ExactlyOneOfNestedAttrsValidator(panelkit.ExactlyOneOfNestedAttrsOpts{
					AttrNames:     lensByValueSourceAttrNames(),
					Summary:       "Invalid lens_dashboard_app_config.by_value",
					MissingDetail: "Set exactly one of `config_json` or one supported typed Lens chart block inside `by_value`.",
					TooManyDetail: "Set exactly one of `config_json` or one supported typed Lens chart block inside `by_value` (more than one by-value source is set).",
					Description:   "Ensures exactly one of `config_json` or one supported typed Lens chart block is set inside `by_value`.",
				}),
			},
		},
		"by_reference": schema.SingleNestedAttribute{
			MarkdownDescription: "By-reference `lens-dashboard-app` configuration: structured `drilldowns`, `ref_id`, optional `references_json`, and required `time_range`.",
			Optional:            true,
			Attributes:          panelkit.ByReferenceAttributes(),
		},
	}
}

// SchemaAttribute returns the Terraform schema for `lens_dashboard_app_config`.
func SchemaAttribute() schema.Attribute {
	return panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{
		Description: "Configuration for a `lens-dashboard-app` panel (the Kibana Dashboard API `lens-dashboard-app` panel type). " +
			"Set exactly one of `by_value` or `by_reference`. " +
			"With `by_value`, set exactly one of `config_json` or one supported typed Lens chart block. " +
			"With `by_reference`, use `ref_id` and `references_json` to map the API `references` list. " +
			"Supported typed by-value blocks are sent as the `lens-dashboard-app` API `config` and do not create `type = \"vis\"` panels.",
		BlockName:  "lens_dashboard_app_config",
		PanelType:  "lens-dashboard-app",
		Required:   true,
		Attributes: innerSchemaAttributes(),
		ExtraValidators: []validator.Object{
			panelkit.ExactlyOneOfNestedAttrsValidator(panelkit.ExactlyOneOfNestedAttrsOpts{
				AttrNames:     []string{"by_value", "by_reference"},
				Summary:       "Invalid lens_dashboard_app_config",
				MissingDetail: "Exactly one of `by_value` or `by_reference` must be set inside `lens_dashboard_app_config`.",
				TooManyDetail: "Exactly one of `by_value` or `by_reference` must be set inside `lens_dashboard_app_config`, not both.",
				Description:   "Ensures exactly one of `by_value` or `by_reference` is set inside `lens_dashboard_app_config`.",
			}),
		},
	})
}
