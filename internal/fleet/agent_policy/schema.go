package agent_policy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/float32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *agentPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchema()
}

func getSchema() schema.Schema {
	return schema.Schema{
		Description: "Creates a new Fleet Agent Policy. See https://www.elastic.co/guide/en/fleet/current/agent-policy.html",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_id": schema.StringAttribute{
				Description: "Unique identifier of the agent policy.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the agent policy.",
				Required:    true,
			},
			"namespace": schema.StringAttribute{
				Description: "The namespace of the agent policy.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the agent policy.",
				Optional:    true,
			},
			"data_output_id": schema.StringAttribute{
				Description: "The identifier for the data output.",
				Optional:    true,
			},
			"monitoring_output_id": schema.StringAttribute{
				Description: "The identifier for monitoring output.",
				Optional:    true,
			},
			"fleet_server_host_id": schema.StringAttribute{
				Description: "The identifier for the Fleet server host.",
				Optional:    true,
			},
			"download_source_id": schema.StringAttribute{
				Description: "The identifier for the Elastic Agent binary download server.",
				Optional:    true,
			},
			"monitor_logs": schema.BoolAttribute{
				Description: "Enable collection of agent logs.",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(false),
			},
			"monitor_metrics": schema.BoolAttribute{
				Description: "Enable collection of agent metrics.",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(false),
			},
			"skip_destroy": schema.BoolAttribute{
				Description: "Set to true if you do not wish the agent policy to be deleted at destroy time, and instead just remove the agent policy from the Terraform state.",
				Optional:    true,
			},
			"supports_agentless": schema.BoolAttribute{
				Description: "Set to true to enable agentless data collection.",
				Optional:    true,
			},
			"sys_monitoring": schema.BoolAttribute{
				Description: "Enable collection of system logs and metrics.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"inactivity_timeout": schema.StringAttribute{
				Description: "The inactivity timeout for the agent policy. If an agent does not report within this time period, it will be considered inactive. Supports duration strings (e.g., '30s', '2m', '1h').",
				Computed:    true,
				Optional:    true,
				CustomType:  customtypes.DurationType{},
			},
			"unenrollment_timeout": schema.StringAttribute{
				Description: "The unenrollment timeout for the agent policy. If an agent is inactive for this period, it will be automatically unenrolled. Supports duration strings (e.g., '30s', '2m', '1h').",
				Computed:    true,
				Optional:    true,
				CustomType:  customtypes.DurationType{},
			},
			"global_data_tags": schema.MapNestedAttribute{
				Description: "User-defined data tags to apply to all inputs. Values can be strings (string_value) or numbers (number_value) but not both. Example -- key1 = {string_value = value1}, key2 = {number_value = 42}",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_value": schema.StringAttribute{
							Description: "String value for the field. If this is set, number_value must not be defined.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("number_value")),
							},
						},
						"number_value": schema.Float32Attribute{
							Description: "Number value for the field. If this is set, string_value must not be defined.",
							Optional:    true,
							Validators: []validator.Float32{
								float32validator.ConflictsWith(path.MatchRelative().AtParent().AtName("string_value")),
							},
						},
					},
				},
				Computed: true,
				Optional: true,
				Default: mapdefault.StaticValue(types.MapValueMust(types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"string_value": types.StringType,
						"number_value": types.Float32Type,
					},
				}, map[string]attr.Value{})),
			},
			"space_ids": schema.SetAttribute{
				Description: "The Kibana space IDs that this agent policy should be available in. When not specified, defaults to [\"default\"]. Note: The order of space IDs does not matter as this is a set.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"required_versions": schema.MapAttribute{
				Description: "Map of agent versions to target percentages for automatic upgrade. The key is the target version and the value is the percentage of agents to upgrade to that version.",
				ElementType: types.Int32Type,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Map{
					mapvalidator.ValueInt32sAre(
						int32validator.Between(0, 100),
					),
				},
			},
		}}
}
func getGlobalDataTagsAttrTypes() attr.Type {
	return getSchema().Attributes["global_data_tags"].GetType()
}
