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

package markdown

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const panelConfigBlock = "markdown_config"

// SchemaAttribute returns the markdown_config SingleNestedAttribute (lifted from dashboard/schema.go).
func SchemaAttribute() schema.Attribute {
	return panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{
		Description: "Configuration for a `markdown` panel (the Kibana Dashboard API `kbn-dashboard-panel-type-markdown` shape). " +
			"Set exactly one of `by_value` (inline `content` with required nested `settings`) or `by_reference` (existing library item via `ref_id`). " +
			"Presentation fields (`description`, `hide_title`, `title`, `hide_border`) are supported in both branches.",
		BlockName:  panelConfigBlock,
		PanelType:  panelType,
		Attributes: nestedAttributes(),
		ExtraValidators: []validator.Object{
			panelkit.ExactlyOneOfNestedAttrsValidator(panelkit.ExactlyOneOfNestedAttrsOpts{
				AttrNames:     []string{"by_value", "by_reference"},
				Summary:       "Invalid " + panelConfigBlock,
				MissingDetail: "Exactly one of `by_value` or `by_reference` must be set inside `markdown_config`.",
				TooManyDetail: "Exactly one of `by_value` or `by_reference` must be set inside `markdown_config`, not both.",
				Description:   "Ensures exactly one of `by_value` or `by_reference` is set inside `markdown_config`.",
			}),
		},
	})
}

func nestedAttributes() map[string]schema.Attribute {
	byValueAttrs := panelkit.PanelPresentationAttributes()
	byValueAttrs["content"] = schema.StringAttribute{
		MarkdownDescription: "Markdown source for the panel body (API `content`).",
		Required:            true,
	}
	byValueAttrs["settings"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Required settings object for by-value markdown. " +
			"`open_links_in_new_tab` is optional; when unset, Kibana applies its default (`true`).",
		Required: true,
		Attributes: map[string]schema.Attribute{
			"open_links_in_new_tab": schema.BoolAttribute{
				MarkdownDescription: "When true, links in the markdown open in a new tab. When omitted, Kibana defaults to true.",
				Optional:            true,
			},
		},
	}

	byReferenceAttrs := panelkit.PanelPresentationAttributes()
	byReferenceAttrs["ref_id"] = schema.StringAttribute{
		MarkdownDescription: "Unique identifier of the markdown library item (API `ref_id`). The provider does not verify the item exists at plan time.",
		Required:            true,
	}

	return map[string]schema.Attribute{
		"by_value": schema.SingleNestedAttribute{
			MarkdownDescription: "Inline markdown: required `content` and nested `settings` (API `settings` object). " +
				"Optional `description`, `hide_title`, `title`, and `hide_border`.",
			Optional:   true,
			Attributes: byValueAttrs,
		},
		"by_reference": schema.SingleNestedAttribute{
			MarkdownDescription: "Reference an existing markdown library item via `ref_id`. " +
				"Optional `description`, `hide_title`, `title`, and `hide_border`.",
			Optional:   true,
			Attributes: byReferenceAttrs,
		},
	}
}
