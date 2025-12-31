package dashboard

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchema()
}

func getSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages Kibana [dashboards](https://www.elastic.co/docs/api/doc/kibana). This functionality is in technical preview and may be changed or removed in a future release.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Generated composite identifier for the dashboard.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: "An identifier for the space. If space_id is not provided, the default space is used.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("default"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"dashboard_id": schema.StringAttribute{
				MarkdownDescription: "A unique identifier for the dashboard. If not provided, one will be generated.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "A human-readable title for the dashboard.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A short description of the dashboard.",
				Optional:            true,
			},
			"time_from": schema.StringAttribute{
				MarkdownDescription: "The start time for the dashboard's time range (e.g., 'now-15m', '2023-01-01T00:00:00Z').",
				Required:            true,
			},
			"time_to": schema.StringAttribute{
				MarkdownDescription: "The end time for the dashboard's time range (e.g., 'now', '2023-12-31T23:59:59Z').",
				Required:            true,
			},
			"time_range_mode": schema.StringAttribute{
				MarkdownDescription: "The time range mode. Valid values are 'absolute' or 'relative'.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("absolute", "relative"),
				},
			},
			"refresh_interval_pause": schema.BoolAttribute{
				MarkdownDescription: "Set to false to auto-refresh data on an interval.",
				Required:            true,
			},
			"refresh_interval_value": schema.Int64Attribute{
				MarkdownDescription: "A numeric value indicating refresh frequency in milliseconds.",
				Required:            true,
			},
			"query_language": schema.StringAttribute{
				MarkdownDescription: "The query language (e.g., 'kuery', 'lucene').",
				Required:            true,
			},
			"query_text": schema.StringAttribute{
				MarkdownDescription: "The query text for text-based queries such as Kibana Query Language (KQL) or Lucene query language. Mutually exclusive with `query_json`.",
				Optional:            true,
			},
			"query_json": schema.StringAttribute{
				MarkdownDescription: "The query as a JSON object for structured queries. Mutually exclusive with `query_text`.",
				CustomType:          jsontypes.NormalizedType{},
				Optional:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "An array of tag IDs applied to this dashboard.",
				ElementType:         types.StringType,
				Optional:            true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"control_group_input": schema.SingleNestedAttribute{
				MarkdownDescription: "Configuration for dashboard controls (filters, time range selectors, etc.).",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"auto_apply_selections": schema.BoolAttribute{
						MarkdownDescription: "Show apply selections button in controls.",
						Optional:            true,
					},
					"chaining_system": schema.StringAttribute{
						MarkdownDescription: "The chaining strategy for multiple controls. Valid values are 'HIERARCHICAL' or 'NONE'.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("HIERARCHICAL", "NONE"),
						},
					},
					"label_position": schema.StringAttribute{
						MarkdownDescription: "Position of the labels for controls. Valid values are 'oneLine' or 'twoLine'.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("oneLine", "twoLine"),
						},
					},
					"ignore_parent_settings": schema.SingleNestedAttribute{
						MarkdownDescription: "Settings to ignore global dashboard settings in controls.",
						Optional:            true,
						CustomType:          NewIgnoreParentSettingsType(),
						Attributes: map[string]schema.Attribute{
							"ignore_filters": schema.BoolAttribute{
								MarkdownDescription: "Ignore global filters in controls.",
								Optional:            true,
							},
							"ignore_query": schema.BoolAttribute{
								MarkdownDescription: "Ignore the global query bar in controls.",
								Optional:            true,
							},
							"ignore_timerange": schema.BoolAttribute{
								MarkdownDescription: "Ignore the global time range in controls.",
								Optional:            true,
							},
							"ignore_validations": schema.BoolAttribute{
								MarkdownDescription: "Ignore validations in controls.",
								Optional:            true,
							},
						},
					},
					"controls": schema.ListNestedAttribute{
						MarkdownDescription: "An array of control panels and their state in the control group.",
						Optional:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									MarkdownDescription: "The unique ID of the control.",
									Optional:            true,
									Computed:            true,
								},
								"type": schema.StringAttribute{
									MarkdownDescription: "The type of the control panel.",
									Required:            true,
								},
								"order": schema.Float64Attribute{
									MarkdownDescription: "The order of the control panel in the control group.",
									Required:            true,
								},
								"width": schema.StringAttribute{
									MarkdownDescription: "Minimum width of the control panel in the control group. Valid values are 'small', 'medium', or 'large'.",
									Optional:            true,
									Validators: []validator.String{
										stringvalidator.OneOf("small", "medium", "large"),
									},
								},
								"grow": schema.BoolAttribute{
									MarkdownDescription: "Expand width of the control panel to fit available space.",
									Optional:            true,
								},
								"control_config": schema.StringAttribute{
									MarkdownDescription: "The control configuration as a JSON object.",
									CustomType:          jsontypes.NormalizedType{},
									Optional:            true,
								},
							},
						},
					},
					"enhancements": schema.StringAttribute{
						MarkdownDescription: "Enhancements configuration as a JSON object.",
						CustomType:          jsontypes.NormalizedType{},
						Optional:            true,
					},
				},
			},
			"options": schema.SingleNestedAttribute{
				MarkdownDescription: "Display options for the dashboard.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"hide_panel_titles": schema.BoolAttribute{
						MarkdownDescription: "Hide the panel titles in the dashboard.",
						Optional:            true,
					},
					"use_margins": schema.BoolAttribute{
						MarkdownDescription: "Show margins between panels in the dashboard layout.",
						Optional:            true,
					},
					"sync_colors": schema.BoolAttribute{
						MarkdownDescription: "Synchronize colors between related panels in the dashboard.",
						Optional:            true,
					},
					"sync_tooltips": schema.BoolAttribute{
						MarkdownDescription: "Synchronize tooltips between related panels in the dashboard.",
						Optional:            true,
					},
					"sync_cursor": schema.BoolAttribute{
						MarkdownDescription: "Synchronize cursor position between related panels in the dashboard.",
						Optional:            true,
					},
				},
			},
		},
	}
}

func controlGroupInputControlsType() basetypes.ObjectTypable {
	return getSchema().Attributes["control_group_input"].(schema.SingleNestedAttribute).Attributes["controls"].(schema.ListNestedAttribute).NestedObject.Type()
}
