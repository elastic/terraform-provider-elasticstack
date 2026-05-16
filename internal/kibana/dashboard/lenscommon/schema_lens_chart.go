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

package lenscommon

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ validator.Object = DrilldownListItemVariantsValidator{}

	exprDrilldownDashboardDrilldown = path.MatchRelative().AtParent().AtName("dashboard_drilldown")
	exprDrilldownDiscoverDrilldown  = path.MatchRelative().AtParent().AtName("discover_drilldown")
	exprDrilldownURLDrilldown       = path.MatchRelative().AtParent().AtName("url_drilldown")
)

// DrilldownListItemVariantsValidator rejects drilldown list items where none of the three variant blocks are set.
// Pairwise mutual exclusion when multiple variants are set is enforced via
// validators.ForbiddenIfDrilldownVariantSiblingNestedPresent on each variant block (REQ-039).
type DrilldownListItemVariantsValidator struct{}

func (DrilldownListItemVariantsValidator) Description(_ context.Context) string {
	return "Requires at least one drilldown variant (`dashboard_drilldown`, `discover_drilldown`, or `url_drilldown`); multiple variants are rejected by sibling object validators."
}

func (v DrilldownListItemVariantsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (DrilldownListItemVariantsValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsUnknown() {
		return
	}
	if req.ConfigValue.IsNull() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid drilldown",
			"Set exactly one of `dashboard_drilldown`, `discover_drilldown`, or `url_drilldown`.",
		)
		return
	}
	attrs := req.ConfigValue.Attributes()
	count := 0
	for _, key := range []string{"dashboard_drilldown", "discover_drilldown", "url_drilldown"} {
		val, okAttr := attrs[key]
		if !okAttr || val.IsNull() || val.IsUnknown() {
			continue
		}
		count++
	}
	if count == 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid drilldown",
			"Set exactly one of `dashboard_drilldown`, `discover_drilldown`, or `url_drilldown`.",
		)
	}
}

// LensChartPresentationTimeRangeAttributes returns nested attributes for chart-root time_range.
func LensChartPresentationTimeRangeAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"from": schema.StringAttribute{
			MarkdownDescription: "Start of the chart time range.",
			Required:            true,
		},
		"to": schema.StringAttribute{
			MarkdownDescription: "End of the chart time range.",
			Required:            true,
		},
		"mode": schema.StringAttribute{
			MarkdownDescription: "Optional time range mode. Valid values are `absolute` or `relative`. " +
				"When the GET API omits `mode`, the provider preserves the prior chart `time_range.mode` from configuration or state " +
				"(same pattern as REQ-009 on the dashboard `time_range`).",
			Optional: true,
			Validators: []validator.String{
				stringvalidator.OneOf("absolute", "relative"),
			},
		},
	}
}

// LensChartDrilldownListItemAttributes returns attributes for one drilldown list entry.
func LensChartDrilldownListItemAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"dashboard_drilldown": schema.SingleNestedAttribute{
			MarkdownDescription: "Navigate to another dashboard using current filters/time range.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"dashboard_id": schema.StringAttribute{
					MarkdownDescription: "Target dashboard id.",
					Required:            true,
				},
				"label": schema.StringAttribute{
					MarkdownDescription: "Human-readable drilldown label.",
					Required:            true,
				},
				"trigger": schema.StringAttribute{
					MarkdownDescription: "**Computed** — Kibana fixes this to `on_apply_filter`; reflected in state after apply. Do not set in configuration.",
					Computed:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"use_filters": schema.BoolAttribute{
					MarkdownDescription: "When true, forwards filter context.",
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(true),
				},
				"use_time_range": schema.BoolAttribute{
					MarkdownDescription: "When true, forwards the time range.",
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(true),
				},
				"open_in_new_tab": schema.BoolAttribute{
					MarkdownDescription: "When true, opens the target dashboard in a new browser tab.",
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(false),
				},
			},
			Validators: []validator.Object{
				validators.ForbiddenIfDrilldownVariantSiblingNestedPresent(exprDrilldownDiscoverDrilldown),
				validators.ForbiddenIfDrilldownVariantSiblingNestedPresent(exprDrilldownURLDrilldown),
			},
		},
		"discover_drilldown": schema.SingleNestedAttribute{
			MarkdownDescription: "Open Discover with contextual filters.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"label": schema.StringAttribute{
					MarkdownDescription: "Human-readable drilldown label.",
					Required:            true,
				},
				"trigger": schema.StringAttribute{
					MarkdownDescription: "**Computed** — Kibana fixes this to `on_apply_filter`; reflected in state after apply. Do not set in configuration.",
					Computed:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"open_in_new_tab": schema.BoolAttribute{
					MarkdownDescription: "When true, opens Discover in a new browser tab.",
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(true),
				},
			},
			Validators: []validator.Object{
				validators.ForbiddenIfDrilldownVariantSiblingNestedPresent(exprDrilldownDashboardDrilldown),
				validators.ForbiddenIfDrilldownVariantSiblingNestedPresent(exprDrilldownURLDrilldown),
			},
		},
		"url_drilldown": schema.SingleNestedAttribute{
			MarkdownDescription: "Open a URL drilldown configured with explicit trigger semantics.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"url": schema.StringAttribute{
					MarkdownDescription: "Destination URL.",
					Required:            true,
				},
				"label": schema.StringAttribute{
					MarkdownDescription: "Human-readable drilldown label.",
					Required:            true,
				},
				"trigger": schema.StringAttribute{
					MarkdownDescription: "Trigger that fires this drilldown.",
					Required:            true,
					Validators: []validator.String{
						stringvalidator.OneOf(
							"on_click_row",
							"on_click_value",
							"on_open_panel_menu",
							"on_select_range",
						),
					},
				},
				"encode_url": schema.BoolAttribute{
					MarkdownDescription: "When true, encodes interpolated URL parameters.",
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(true),
				},
				"open_in_new_tab": schema.BoolAttribute{
					MarkdownDescription: "When true, opens the URL in a new browser tab.",
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(true),
				},
			},
			Validators: []validator.Object{
				validators.ForbiddenIfDrilldownVariantSiblingNestedPresent(exprDrilldownDashboardDrilldown),
				validators.ForbiddenIfDrilldownVariantSiblingNestedPresent(exprDrilldownDiscoverDrilldown),
			},
		},
	}
}

// LensChartPresentationAttributes returns optional chart-root presentation fields shared by typed Lens chart blocks.
func LensChartPresentationAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"time_range": schema.SingleNestedAttribute{
			MarkdownDescription: "Chart-level time selection (`from`, `to`, optional `mode`), same shape as the dashboard root `time_range`. " +
				"When omitted (null), the provider inherits the dashboard-level `time_range` on write and preserves null in state when the API echoes the inherited value on read.",
			Optional:   true,
			Attributes: LensChartPresentationTimeRangeAttributes(),
		},
		"hide_title": schema.BoolAttribute{
			MarkdownDescription: "When true, suppresses the chart title.",
			Optional:            true,
		},
		"hide_border": schema.BoolAttribute{
			MarkdownDescription: "When true, suppresses the chart panel border.",
			Optional:            true,
		},
		"references_json": schema.StringAttribute{
			MarkdownDescription: "Optional normalized JSON array of `{ id, name, type }` saved-object references, matching the chart root API `references` list.",
			Optional:            true,
			CustomType:          jsontypes.NormalizedType{},
		},
		"drilldowns": schema.ListNestedAttribute{
			MarkdownDescription: "Optional drilldowns for this chart (max 100 per Kibana API). Each entry sets exactly one of `dashboard_drilldown`, `discover_drilldown`, or `url_drilldown`.",
			Optional:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: LensChartDrilldownListItemAttributes(),
				Validators: []validator.Object{
					DrilldownListItemVariantsValidator{},
				},
			},
			Validators: []validator.List{
				listvalidator.SizeAtMost(100),
			},
		},
	}
}

// LensChartBaseAttributes returns title, description, sampling, ignore_global_filters, and filters.
func LensChartBaseAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"title": schema.StringAttribute{
			MarkdownDescription: "The title of the chart displayed in the panel.",
			Optional:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "The description of the chart.",
			Optional:            true,
		},
		"sampling": schema.Float64Attribute{
			MarkdownDescription: "Sampling factor between 0 (no sampling) and 1 (full sampling). Default is 1.",
			Optional:            true,
			Computed:            true,
		},
		"ignore_global_filters": schema.BoolAttribute{
			MarkdownDescription: "If true, ignore global filters when fetching data for this chart. Default is false.",
			Optional:            true,
			Computed:            true,
		},
		"filters": schema.ListNestedAttribute{
			MarkdownDescription: "Additional filters to apply to the chart data (maximum 100).",
			Optional:            true,
			NestedObject:        LensChartFilterNestedObject(),
			Validators: []validator.List{
				listvalidator.SizeAtMost(100),
			},
		},
	}
}

// LensChartFilterSimpleAttributes returns attributes for chart query.language / query.expression.
func LensChartFilterSimpleAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"language": schema.StringAttribute{
			MarkdownDescription: "Query language (default: 'kql').",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("kql", "lucene"),
			},
		},
		"expression": schema.StringAttribute{
			MarkdownDescription: "Filter expression string.",
			Required:            true,
		},
	}
}

// LensChartFilterNestedObject describes one chart-level filter_json entry.
func LensChartFilterNestedObject() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"filter_json": schema.StringAttribute{
				MarkdownDescription: "Chart filter as normalized JSON. Must match the Kibana dashboard API for this chart: " +
					"one of the filter union members (condition, group, DSL, or spatial) described in the dashboards OpenAPI specification.",
				CustomType: jsontypes.NormalizedType{},
				Required:   true,
			},
		},
	}
}
