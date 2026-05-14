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
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const discoverSessionDrilldownMax = 100
const discoverSessionFilterMax = 100
const discoverSessionColumnOrderMax = 100

// discoverSessionRowHeightStringValidator validates optional Discover row/header height strings:
// either the literal "auto" or a base-10 integer in [min, max], parsed with strconv.Atoi.
// Note: leading-zero strings such as "01" are accepted and normalize to the same integer value
// (Atoi does not enforce a no-leading-zeros rule); only non-numeric input or values outside the
// configured range are rejected.
type discoverSessionRowHeightStringValidator struct {
	min, max int
}

func makeDiscoverSessionRowHeightStringValidator(maxHeight int) validator.String {
	return discoverSessionRowHeightStringValidator{min: 1, max: maxHeight}
}

func (v discoverSessionRowHeightStringValidator) Description(_ context.Context) string {
	return "Must be \"auto\" or a decimal integer in the configured inclusive range."
}

func (v discoverSessionRowHeightStringValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v discoverSessionRowHeightStringValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	s := req.ConfigValue.ValueString()
	if s == dashboardValueAuto {
		return
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < v.min || n > v.max {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid row height",
			`Must be "auto" or a decimal integer between `+strconv.Itoa(v.min)+` and `+strconv.Itoa(v.max)+`, inclusive.`,
		)
	}
}

var _ validator.Object = discoverSessionConfigModeValidator{}

// discoverSessionConfigModeValidator ensures exactly one of by_value or by_reference is set.
type discoverSessionConfigModeValidator struct{}

func (discoverSessionConfigModeValidator) Description(_ context.Context) string {
	return "Ensures exactly one of `by_value` or `by_reference` is set inside `discover_session_config`."
}

func (v discoverSessionConfigModeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (discoverSessionConfigModeValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	attrs := req.ConfigValue.Attributes()
	byValue := attrs["by_value"]
	byRef := attrs["by_reference"]
	valueSet := func(av attr.Value) bool {
		return av != nil && !av.IsNull() && !av.IsUnknown()
	}
	byValueSet := valueSet(byValue)
	byRefSet := valueSet(byRef)
	if byValueSet && byRefSet {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid discover_session_config", "Exactly one of `by_value` or `by_reference` must be set inside `discover_session_config`, not both.")
		return
	}
	if !byValueSet && !byRefSet {
		if byValue != nil && byValue.IsUnknown() || byRef != nil && byRef.IsUnknown() {
			return
		}
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid discover_session_config", "Exactly one of `by_value` or `by_reference` must be set inside `discover_session_config`.")
	}
}

var _ validator.Object = discoverSessionTabModeValidator{}

// discoverSessionTabModeValidator ensures exactly one of dsl or esql is set on `by_value.tab`.
type discoverSessionTabModeValidator struct{}

func (discoverSessionTabModeValidator) Description(_ context.Context) string {
	return "Ensures exactly one of `dsl` or `esql` is set inside `discover_session_config.by_value.tab`."
}

func (v discoverSessionTabModeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (discoverSessionTabModeValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	attrs := req.ConfigValue.Attributes()
	dsl := attrs["dsl"]
	esql := attrs["esql"]
	valueSet := func(av attr.Value) bool {
		return av != nil && !av.IsNull() && !av.IsUnknown()
	}
	dslSet := valueSet(dsl)
	esqlSet := valueSet(esql)
	if dslSet && esqlSet {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid discover_session tab", "Exactly one of `dsl` or `esql` must be set inside `tab`, not both.")
		return
	}
	if !dslSet && !esqlSet {
		if dsl != nil && dsl.IsUnknown() || esql != nil && esql.IsUnknown() {
			return
		}
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid discover_session tab", "Exactly one of `dsl` or `esql` must be set inside `tab`.")
	}
}

func discoverSessionColumnSettingsAttribute() schema.Attribute {
	return schema.MapNestedAttribute{
		MarkdownDescription: "Per-column presentation settings keyed by field name (for example column widths).",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"width": schema.Float64Attribute{
					MarkdownDescription: "Optional column width in pixels.",
					Optional:            true,
				},
			},
		},
	}
}

func discoverSessionSortListAttribute(desc string) schema.Attribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: desc,
		Optional:            true,
		Validators: []validator.List{
			listvalidator.SizeAtMost(discoverSessionColumnOrderMax),
		},
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: "Field name to sort by.",
					Required:            true,
				},
				"direction": schema.StringAttribute{
					MarkdownDescription: "Sort direction.",
					Required:            true,
					Validators: []validator.String{
						stringvalidator.OneOf("asc", "desc"),
					},
				},
			},
		},
	}
}

func discoverSessionDSLTabAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"column_order": schema.ListAttribute{
			MarkdownDescription: "Ordered list of field names shown in the Discover grid.",
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.List{
				listvalidator.SizeAtMost(discoverSessionColumnOrderMax),
			},
		},
		"column_settings": discoverSessionColumnSettingsAttribute(),
		"sort": discoverSessionSortListAttribute(
			"Sort configuration for the Discover grid.",
		),
		"density": schema.StringAttribute{
			MarkdownDescription: "Data grid density.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("compact", "expanded", "normal"),
			},
		},
		"header_row_height": schema.StringAttribute{
			MarkdownDescription: `Header row height: numbers "1"–"5" (as decimal strings) or "auto".`,
			Optional:            true,
			Validators: []validator.String{
				makeDiscoverSessionRowHeightStringValidator(5),
			},
		},
		"row_height": schema.StringAttribute{
			MarkdownDescription: `Data row height: numbers "1"–"20" (as decimal strings) or "auto".`,
			Optional:            true,
			Validators: []validator.String{
				makeDiscoverSessionRowHeightStringValidator(20),
			},
		},
		"rows_per_page": schema.Int64Attribute{
			MarkdownDescription: "Rows per page in the Discover grid.",
			Optional:            true,
			Validators: []validator.Int64{
				int64validator.Between(1, 10000),
			},
		},
		"sample_size": schema.Int64Attribute{
			MarkdownDescription: "Sample size (documents) for the Discover grid.",
			Optional:            true,
			Validators: []validator.Int64{
				int64validator.Between(10, 10000),
			},
		},
		"view_mode": schema.StringAttribute{
			MarkdownDescription: discoverSessionPanelViewModeDescription,
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("documents", "patterns", "aggregated"),
			},
		},
		"query": schema.SingleNestedAttribute{
			MarkdownDescription: "Kibana query used by this DSL Discover tab (KQL or Lucene).",
			Required:            true,
			Attributes:          getFilterSimple(),
		},
		"data_source_json": schema.StringAttribute{
			MarkdownDescription: discoverSessionPanelDataSourceJSONDescription,
			Required:            true,
			CustomType:          jsontypes.NormalizedType{},
		},
		"filters": schema.ListNestedAttribute{
			MarkdownDescription: discoverSessionPanelDSLFiltersDescription,
			Optional:            true,
			Validators: []validator.List{
				listvalidator.SizeAtMost(discoverSessionFilterMax),
			},
			NestedObject: getChartFilter(),
		},
	}
}

func discoverSessionESQLTabAttributes() map[string]schema.Attribute {
	attrs := map[string]schema.Attribute{
		"column_order": schema.ListAttribute{
			MarkdownDescription: "Ordered list of field names shown in the Discover grid.",
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.List{
				listvalidator.SizeAtMost(discoverSessionColumnOrderMax),
			},
		},
		"column_settings": discoverSessionColumnSettingsAttribute(),
		"sort": discoverSessionSortListAttribute(
			"Sort configuration for the Discover grid.",
		),
		"density": schema.StringAttribute{
			MarkdownDescription: "Data grid density.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("compact", "expanded", "normal"),
			},
		},
		"header_row_height": schema.StringAttribute{
			MarkdownDescription: `Header row height: numbers "1"–"5" (as decimal strings) or "auto".`,
			Optional:            true,
			Validators: []validator.String{
				makeDiscoverSessionRowHeightStringValidator(5),
			},
		},
		"row_height": schema.StringAttribute{
			MarkdownDescription: `Data row height: numbers "1"–"20" (as decimal strings) or "auto".`,
			Optional:            true,
			Validators: []validator.String{
				makeDiscoverSessionRowHeightStringValidator(20),
			},
		},
		"data_source_json": schema.StringAttribute{
			MarkdownDescription: discoverSessionPanelDataSourceJSONDescription,
			Required:            true,
			CustomType:          jsontypes.NormalizedType{},
		},
	}
	return attrs
}

func discoverSessionOverridesAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"column_order": schema.ListAttribute{
			MarkdownDescription: "Overrides column order relative to the referenced Discover session.",
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.List{
				listvalidator.SizeAtMost(discoverSessionColumnOrderMax),
			},
		},
		"column_settings": discoverSessionColumnSettingsAttribute(),
		"sort": discoverSessionSortListAttribute(
			"Overrides sort configuration relative to the referenced Discover session.",
		),
		"density": schema.StringAttribute{
			MarkdownDescription: "Overrides data grid density.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("compact", "expanded", "normal"),
			},
		},
		"header_row_height": schema.StringAttribute{
			MarkdownDescription: `Overrides header row height: numbers "1"–"5" or "auto".`,
			Optional:            true,
			Validators: []validator.String{
				makeDiscoverSessionRowHeightStringValidator(5),
			},
		},
		"row_height": schema.StringAttribute{
			MarkdownDescription: `Overrides data row height: numbers "1"–"20" or "auto".`,
			Optional:            true,
			Validators: []validator.String{
				makeDiscoverSessionRowHeightStringValidator(20),
			},
		},
		"rows_per_page": schema.Int64Attribute{
			MarkdownDescription: "Overrides rows per page.",
			Optional:            true,
			Validators: []validator.Int64{
				int64validator.Between(1, 10000),
			},
		},
		"sample_size": schema.Int64Attribute{
			MarkdownDescription: "Overrides sample size.",
			Optional:            true,
			Validators: []validator.Int64{
				int64validator.Between(10, 10000),
			},
		},
	}
}

// getDiscoverSessionPanelConfigAttributes returns attributes for `discover_session_config`.
func getDiscoverSessionPanelConfigAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"title": schema.StringAttribute{
			MarkdownDescription: "Optional panel title.",
			Optional:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "Optional panel description.",
			Optional:            true,
		},
		"hide_title": schema.BoolAttribute{
			MarkdownDescription: "When true, suppresses the panel title.",
			Optional:            true,
		},
		"hide_border": schema.BoolAttribute{
			MarkdownDescription: "When true, suppresses the panel border.",
			Optional:            true,
		},
		"drilldowns": schema.ListNestedAttribute{
			MarkdownDescription: discoverSessionPanelDrilldownsDescription,
			Optional:            true,
			Validators: []validator.List{
				listvalidator.SizeAtMost(discoverSessionDrilldownMax),
			},
			NestedObject: urlDrilldownNestedAttributeObject(URLDrilldownNestedOpts{
				AllowedTriggers:                 []string{"on_open_panel_menu"},
				URLMarkdownDescription:          "The URL template for the drilldown. Variables are documented at https://www.elastic.co/docs/explore-analyze/dashboards/drilldowns#url-template-variable.",
				LabelMarkdownDescription:        "The display label for the drilldown link.",
				EncodeURLMarkdownDescription:    "When true, the URL is percent-encoded.",
				OpenInNewTabMarkdownDescription: "When true, the drilldown URL opens in a new browser tab.",
			}),
		},
		"by_value": schema.SingleNestedAttribute{
			MarkdownDescription: discoverSessionPanelByValueDescription,
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"time_range": timeRangeSingleNestedAttribute(
					"Optional time range for this panel. When omitted, the dashboard root `time_range` is sent to the API at write time while this attribute stays null in state (REQ-009).",
					false,
				),
				"tab": schema.SingleNestedAttribute{
					MarkdownDescription: "Single Discover tab configuration (the API currently allows one tab). Exactly one of `dsl` or `esql` must be set.",
					Required:            true,
					Attributes: map[string]schema.Attribute{
						"dsl": schema.SingleNestedAttribute{
							MarkdownDescription: "DSL / data view Discover tab.",
							Optional:            true,
							Attributes:          discoverSessionDSLTabAttributes(),
							Validators: []validator.Object{
								objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("esql")),
							},
						},
						"esql": schema.SingleNestedAttribute{
							MarkdownDescription: "ES|QL Discover tab.",
							Optional:            true,
							Attributes:          discoverSessionESQLTabAttributes(),
							Validators: []validator.Object{
								objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("dsl")),
							},
						},
					},
					Validators: []validator.Object{
						discoverSessionTabModeValidator{},
					},
				},
			},
			Validators: []validator.Object{
				objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("by_reference")),
			},
		},
		"by_reference": schema.SingleNestedAttribute{
			MarkdownDescription: discoverSessionPanelByReferenceDescription,
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"time_range": timeRangeSingleNestedAttribute(
					"Optional time range for this panel. When omitted, the dashboard root `time_range` is sent to the API at write time while this attribute stays null in state (REQ-009).",
					false,
				),
				"ref_id": schema.StringAttribute{
					MarkdownDescription: "Discover session saved object reference id (`ref_id` in the API).",
					Required:            true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
				"selected_tab_id": schema.StringAttribute{
					MarkdownDescription: "Tab id within the referenced Discover session. Omit to let Kibana choose; after apply the API value is reflected here.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"overrides": schema.SingleNestedAttribute{
					MarkdownDescription: "Optional typed presentation overrides applied on top of the referenced session.",
					Optional:            true,
					Attributes:          discoverSessionOverridesAttributes(),
				},
			},
			Validators: []validator.Object{
				objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("by_value")),
			},
		},
	}
}
