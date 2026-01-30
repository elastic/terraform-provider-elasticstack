package alerting_rule

import (
	"context"
	"regexp"

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

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates a Kibana alerting rule. See https://www.elastic.co/guide/en/kibana/current/alerting-apis.html",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Generated ID for the rule (space_id/rule_id).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"rule_id": schema.StringAttribute{
				Description: "The identifier for the rule. If it is omitted, an ID is randomly generated.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
			"name": schema.StringAttribute{
				Description: "The name of the rule. While this name does not have to be unique, a distinctive name can help you identify a rule.",
				Required:    true,
			},
			"consumer": schema.StringAttribute{
				Description: "The name of the application or feature that owns the rule.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"notify_when": schema.StringAttribute{
				Description: "Defines how often alerts generate actions. Valid values include: onActionGroupChange, onActiveAlert, onThrottleInterval.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("onActionGroupChange", "onActiveAlert", "onThrottleInterval"),
				},
			},
			"params": schema.StringAttribute{
				Description: "The rule parameters as a JSON string, which differ for each rule type.",
				Required:    true,
				Validators: []validator.String{
					ValidJSON(),
				},
			},
			"rule_type_id": schema.StringAttribute{
				Description: "The ID of the rule type that you want to call when the rule is scheduled to run.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"interval": schema.StringAttribute{
				Description: "The check interval, which specifies how frequently the rule conditions are checked. The interval must be specified in seconds (s), minutes (m), hours (h), or days (d).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[1-9][0-9]*(?:d|h|m|s)$`),
						"must be a valid duration (e.g., 1m, 5m, 1h, 1d)",
					),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Indicates if you want to run the rule on an interval basis.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"tags": schema.ListAttribute{
				Description: "A list of tag names that are applied to the rule.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"throttle": schema.StringAttribute{
				Description: "Deprecated in 8.13.0. Defines how often an alert generates repeated actions.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[1-9][0-9]*(?:d|h|m|s)$`),
						"must be a valid duration (e.g., 1m, 5m, 1h, 1d)",
					),
				},
			},
			"scheduled_task_id": schema.StringAttribute{
				Description: "ID of the scheduled task that will execute the alert.",
				Computed:    true,
			},
			"last_execution_status": schema.StringAttribute{
				Description: "Status of the last execution of this rule.",
				Computed:    true,
			},
			"last_execution_date": schema.StringAttribute{
				Description: "Date of the last execution of this rule.",
				Computed:    true,
			},
			"alert_delay": schema.Float64Attribute{
				Description: "A number that indicates how many consecutive runs need to meet the rule conditions for an alert to occur.",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"actions": schema.ListNestedBlock{
				Description: "An action that runs under defined conditions.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"group": schema.StringAttribute{
							Description: "The group name, which affects when the action runs.",
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString("default"),
						},
						"id": schema.StringAttribute{
							Description: "The identifier for the connector saved object.",
							Required:    true,
						},
						"params": schema.StringAttribute{
							Description: "The parameters for the action as a JSON string.",
							Required:    true,
							Validators: []validator.String{
								ValidJSON(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"frequency": schema.SingleNestedBlock{
							Description: "The properties that affect how often actions are generated.",
							Attributes: map[string]schema.Attribute{
								"summary": schema.BoolAttribute{
									Description: "Indicates whether the action is a summary.",
									Required:    true,
								},
								"notify_when": schema.StringAttribute{
									Description: "Defines how often alerts generate actions.",
									Required:    true,
									Validators: []validator.String{
										stringvalidator.OneOf("onActionGroupChange", "onActiveAlert", "onThrottleInterval"),
									},
								},
								"throttle": schema.StringAttribute{
									Description: "Defines how often an alert generates repeated actions.",
									Optional:    true,
									Validators: []validator.String{
										stringvalidator.RegexMatches(
											regexp.MustCompile(`^[1-9][0-9]*(?:d|h|m|s)$`),
											"must be a valid duration",
										),
									},
								},
							},
						},
						"alerts_filter": schema.SingleNestedBlock{
							Description: "Conditions that affect whether the action runs.",
							Attributes: map[string]schema.Attribute{
								"kql": schema.StringAttribute{
									Description: "Defines a query filter that determines whether the action runs.",
									Optional:    true,
								},
							},
							Blocks: map[string]schema.Block{
								"timeframe": schema.SingleNestedBlock{
									Description: "Defines a period that limits whether the action runs.",
									Attributes: map[string]schema.Attribute{
										"days": schema.ListAttribute{
											Description: "Defines the days of the week that the action can run (1=Monday, 7=Sunday).",
											Optional:    true,
											ElementType: types.Int64Type,
										},
										"timezone": schema.StringAttribute{
											Description: "The ISO time zone for the hours values.",
											Optional:    true,
										},
										"hours_start": schema.StringAttribute{
											Description: "The start of the time frame in 24-hour notation (hh:mm).",
											Optional:    true,
										},
										"hours_end": schema.StringAttribute{
											Description: "The end of the time frame in 24-hour notation (hh:mm).",
											Optional:    true,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
