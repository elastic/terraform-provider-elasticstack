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

package visconfig

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func byValueAttributes() map[string]schema.Attribute {
	out := make(map[string]schema.Attribute)
	for _, c := range lenscommon.All() {
		out[chartBlockKeyOrPanic(c.VizType())] = c.SchemaAttribute()
	}
	return out
}

func visByValueChartAttrNames() []string {
	converters := lenscommon.All()
	out := make([]string, 0, len(converters))
	for _, c := range converters {
		out = append(out, chartBlockKeyOrPanic(c.VizType()))
	}
	return out
}

func chartBlockKeyOrPanic(vizType string) string {
	key := lenscommon.TerraformChartBlockKey(vizType)
	if key == "" {
		panic("visconfig: missing terraform chart block key for VizType " + vizType)
	}
	return key
}

func innerSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"by_value": schema.SingleNestedAttribute{
			MarkdownDescription: "Inline by-value Lens visualization configuration for `type = \"vis\"` panels (`vis_config`). " +
				"Exactly one typed chart kind must be set (no raw JSON here — use panel-level `config_json` for that).",
			Optional:   true,
			Attributes: byValueAttributes(),
			Validators: []validator.Object{
				panelkit.ExactlyOneOfNestedAttrsValidator(panelkit.ExactlyOneOfNestedAttrsOpts{
					AttrNames:     visByValueChartAttrNames(),
					Summary:       "Invalid vis_config.by_value",
					MissingDetail: "Set exactly one supported typed Lens chart block inside `vis_config.by_value`.",
					TooManyDetail: "Set exactly one typed chart block inside `vis_config.by_value` (more than one by-value chart is set).",
					Description:   "Ensures exactly one supported typed Lens chart block is set inside `vis_config.by_value`.",
				}),
			},
		},
		"by_reference": schema.SingleNestedAttribute{
			MarkdownDescription: "By-reference `vis` configuration: structured `drilldowns`, `ref_id`, optional `references_json`, and required `time_range`.",
			Optional:            true,
			Attributes:          lenscommon.LensByReferenceAttributes(),
		},
	}
}

// SchemaAttribute returns the Terraform schema for the vis panel typed configuration block (`vis_config`).
func SchemaAttribute() schema.Attribute {
	return panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{
		Description: "Configuration for a `vis` panel (`type = \"vis\"`). " +
			"Typed alternative to panel-level `config_json`: set exactly one of `by_value` (exactly one of 12 Lens chart kinds) or `by_reference`. " +
			"With `by_reference`, use structured `drilldowns` and required `time_range` like `lens_dashboard_app_config.by_reference`.",
		BlockName:  "vis_config",
		PanelType:  panelType,
		Required:   false,
		Attributes: innerSchemaAttributes(),
		ExtraValidators: []validator.Object{
			panelkit.ExactlyOneOfNestedAttrsValidator(panelkit.ExactlyOneOfNestedAttrsOpts{
				AttrNames:     []string{"by_value", "by_reference"},
				Summary:       "Invalid vis_config",
				MissingDetail: "Exactly one of `by_value` or `by_reference` must be set inside `vis_config`.",
				TooManyDetail: "Exactly one of `by_value` or `by_reference` must be set inside `vis_config`, not both.",
				Description:   "Ensures exactly one of `by_value` or `by_reference` is set inside `vis_config`.",
			}),
		},
	})
}
