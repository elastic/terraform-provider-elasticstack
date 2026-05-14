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
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const panelConfigBlock = "markdown_config"

var byValueReferenceAttrNames = []string{"by_value", "by_reference"}

// SchemaAttribute returns the markdown_config SingleNestedAttribute (lifted from dashboard/schema.go).
func SchemaAttribute() schema.Attribute {
	const panelMarkdown = "markdown"
	return schema.SingleNestedAttribute{
		MarkdownDescription: panelkit.PanelConfigDescription(
			"Configuration for a `markdown` panel (the Kibana Dashboard API `kbn-dashboard-panel-type-markdown` shape). "+
				"Set exactly one of `by_value` (inline `content` with required nested `settings`) or `by_reference` (existing library item via `ref_id`). "+
				"Presentation fields (`description`, `hide_title`, `title`, `hide_border`) are supported in both branches.",
			panelConfigBlock,
			panelkit.TypedSiblingPanelConfigBlockNames(),
		),
		Optional:   true,
		Attributes: nestedAttributes(),
		Validators: []validator.Object{
			objectvalidator.ConflictsWith(
				panelkit.SiblingTypedPanelConfigConflictPathsExcept(panelConfigBlock, panelkit.TypedSiblingPanelConfigBlockNames())...,
			),
			validators.AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("type"), []string{panelMarkdown}),
			configModeValidator{},
		},
	}
}

func nestedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"by_value": schema.SingleNestedAttribute{
			MarkdownDescription: "Inline markdown: required `content` and nested `settings` (API `settings` object). " +
				"Optional `description`, `hide_title`, `title`, and `hide_border`.",
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"content": schema.StringAttribute{
					MarkdownDescription: "Markdown source for the panel body (API `content`).",
					Required:            true,
				},
				"settings": schema.SingleNestedAttribute{
					MarkdownDescription: "Required settings object for by-value markdown. " +
						"`open_links_in_new_tab` is optional; when unset, Kibana applies its default (`true`).",
					Required: true,
					Attributes: map[string]schema.Attribute{
						"open_links_in_new_tab": schema.BoolAttribute{
							MarkdownDescription: "When true, links in the markdown open in a new tab. When omitted, Kibana defaults to true.",
							Optional:            true,
						},
					},
				},
				"description": schema.StringAttribute{
					MarkdownDescription: "Optional panel description.",
					Optional:            true,
				},
				"hide_title": schema.BoolAttribute{
					MarkdownDescription: "When true, suppresses the panel title.",
					Optional:            true,
				},
				"title": schema.StringAttribute{
					MarkdownDescription: "Optional panel title.",
					Optional:            true,
				},
				"hide_border": schema.BoolAttribute{
					MarkdownDescription: "When true, suppresses the panel border.",
					Optional:            true,
				},
			},
		},
		"by_reference": schema.SingleNestedAttribute{
			MarkdownDescription: "Reference an existing markdown library item via `ref_id`. " +
				"Optional `description`, `hide_title`, `title`, and `hide_border`.",
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"ref_id": schema.StringAttribute{
					MarkdownDescription: "Unique identifier of the markdown library item (API `ref_id`). The provider does not verify the item exists at plan time.",
					Required:            true,
				},
				"description": schema.StringAttribute{
					MarkdownDescription: "Optional panel description.",
					Optional:            true,
				},
				"hide_title": schema.BoolAttribute{
					MarkdownDescription: "When true, suppresses the panel title.",
					Optional:            true,
				},
				"title": schema.StringAttribute{
					MarkdownDescription: "Optional panel title.",
					Optional:            true,
				},
				"hide_border": schema.BoolAttribute{
					MarkdownDescription: "When true, suppresses the panel border.",
					Optional:            true,
				},
			},
		},
	}
}

var _ validator.Object = configModeValidator{}

type configModeValidator struct{}

func (v configModeValidator) Description(_ context.Context) string {
	return "Ensures exactly one of `by_value` or `by_reference` is set inside `markdown_config`."
}

func (v configModeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v configModeValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	validateExactlyOneNestedAttr(
		req, resp,
		panelConfigBlock,
		byValueReferenceAttrNames,
		"Exactly one of `by_value` or `by_reference` must be set inside `markdown_config`.",
		"Exactly one of `by_value` or `by_reference` must be set inside `markdown_config`, not both.",
	)
}

func validateExactlyOneNestedAttr(
	req validator.ObjectRequest,
	resp *validator.ObjectResponse,
	blockLabel string,
	attrNames []string,
	missingDetail string,
	tooManyDetail string,
) {
	attrs := req.ConfigValue.Attributes()
	count := 0
	hasUnknown := false
	for _, name := range attrNames {
		av, ok := attrs[name]
		if !ok || av == nil {
			continue
		}
		switch {
		case av.IsUnknown():
			hasUnknown = true
		case av.IsNull():
			// not set
		default:
			count++
		}
	}
	if count > 1 {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid "+blockLabel, tooManyDetail)
		return
	}
	if hasUnknown {
		return
	}
	if count == 0 {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid "+blockLabel, missingDetail)
	}
}
