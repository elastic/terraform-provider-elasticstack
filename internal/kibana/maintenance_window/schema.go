package maintenance_window

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *MaintenanceWindowResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Kibana maintenance windows",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Generated ID for the maintenance window.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "An identifier for the space. If space_id is not provided, the default space is used.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("default"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"title": schema.StringAttribute{
				Description: "The name of the maintenance window.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the current maintenance window is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"custom_schedule": schema.SingleNestedAttribute{
				Description: "A set schedule over which the maintenance window applies.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"start": schema.StringAttribute{
						Description: "The start date and time of the schedule, provided in ISO 8601 format and set to the UTC timezone. For example: `2025-03-12T12:00:00.000Z`.",
						Required:    true,
						Validators: []validator.String{
							validators.StringIsISO8601{},
						},
					},
					"duration": schema.StringAttribute{
						Description: "The duration of the schedule. It allows values in `<integer><unit>` format. `<unit>` is one of `d`, `h`, `m`, or `s` for hours, minutes, seconds. For example: `1d`, `5h`, `30m`, `5000s`.",
						Required:    true,
						Validators: []validator.String{
							validators.StringIsAlertingDuration{},
						},
					},
					"timezone": schema.StringAttribute{
						Description: "The timezone of the schedule. The default timezone is UTC.",
						Optional:    true,
						Computed:    true,
					},
					"recurring": schema.SingleNestedAttribute{
						Description: "A set schedule over which the maintenance window applies.",
						Required:    true,
						Attributes: map[string]schema.Attribute{
							"end": schema.StringAttribute{
								Description: "The end date and time of the schedule, provided in ISO 8601 format and set to the UTC timezone. For example: `2025-03-12T12:00:00.000Z`.",
								Optional:    true,
								Validators: []validator.String{
									validators.StringIsISO8601{},
								},
							},
							"every": schema.StringAttribute{
								Description: "The duration of the schedule. It allows values in `<integer><unit>` format. `<unit>` is one of `d`, `h`, `m`, or `s` for hours, minutes, seconds. For example: `1d`, `5h`, `30m`, `5000s`.",
								Optional:    true,
								Validators: []validator.String{
									validators.StringIsMaintenanceWindowIntervalFrequency{},
								},
							},
							"occurrences": schema.Int32Attribute{
								Description: "The total number of recurrences of the schedule.",
								Optional:    true,
								Validators: []validator.Int32{
									int32validator.AtLeast(1),
								},
							},
							"on_week_day": schema.ListAttribute{
								Description: "The specific days of the week (`[MO,TU,WE,TH,FR,SA,SU]`) or nth day of month (`[+1MO, -3FR, +2WE, -4SA, -5SU]`) for a recurring schedule.",
								ElementType: types.StringType,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.ValueStringsAre(
										validators.StringIsMaintenanceWindowOnWeekDay{},
									),
								},
							},
							"on_month_day": schema.ListAttribute{
								Description: "The specific days of the month for a recurring schedule. Valid values are 1-31.",
								ElementType: types.Int32Type,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.ValueInt32sAre(
										int32validator.Between(1, 31),
									),
								},
							},
							"on_month": schema.ListAttribute{
								Description: "The specific months for a recurring schedule. Valid values are 1-12.",
								ElementType: types.Int32Type,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.ValueInt32sAre(
										int32validator.Between(1, 12),
									),
								},
							},
						},
					},
				},
			},
			"scope": schema.SingleNestedAttribute{
				Description: "An object that narrows the scope of what is affected by this maintenance window.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"alerting": schema.SingleNestedAttribute{
						Description: "A set schedule over which the maintenance window applies.",
						Required:    true,
						Attributes: map[string]schema.Attribute{
							"kql": schema.StringAttribute{
								Description: "A filter written in Kibana Query Language (KQL).",
								Required:    true,
							},
						},
					},
				},
			},
		},
	}
}
