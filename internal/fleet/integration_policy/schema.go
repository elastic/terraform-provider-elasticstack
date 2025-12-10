package integration_policy

import (
	"context"
	_ "embed"
	"os"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
)

//go:embed resource-description.md
var integrationPolicyDescription string

func getInputsElementType() attr.Type {
	// Define the element type for the inputs map
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"enabled": types.BoolType,
			"vars":    jsontypes.NormalizedType{},
			"streams": types.MapType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"enabled": types.BoolType,
						"vars":    jsontypes.NormalizedType{},
					},
				},
			},
		},
	}
}

func (r *integrationPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchemaV2()
}

func getSchemaV2() schema.Schema {
	varsAreSensitive := !logging.IsDebugOrHigher() && os.Getenv("TF_ACC") != "1"
	return schema.Schema{
		Version:     2,
		Description: integrationPolicyDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_id": schema.StringAttribute{
				Description: "Unique identifier of the integration policy.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the integration policy.",
				Required:    true,
			},
			"namespace": schema.StringAttribute{
				Description: "The namespace of the integration policy.",
				Required:    true,
			},
			"agent_policy_id": schema.StringAttribute{
				Description: "ID of the agent policy.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Root("agent_policy_ids").Expression()),
				},
			},
			"agent_policy_ids": schema.ListAttribute{
				Description: "List of agent policy IDs.",
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.ConflictsWith(path.Root("agent_policy_id").Expression()),
					listvalidator.SizeAtLeast(1),
				},
			},
			"description": schema.StringAttribute{
				Description: "The description of the integration policy.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Enable the integration policy.",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(true),
			},
			"force": schema.BoolAttribute{
				Description: "Force operations, such as creation and deletion, to occur.",
				Optional:    true,
			},
			"integration_name": schema.StringAttribute{
				Description: "The name of the integration package.",
				Required:    true,
			},
			"integration_version": schema.StringAttribute{
				Description: "The version of the integration package.",
				Required:    true,
			},
			"output_id": schema.StringAttribute{
				Description: "The ID of the output to send data to. When not specified, the default output of the agent policy will be used.",
				Optional:    true,
			},
			"vars_json": schema.StringAttribute{
				Description: "Integration-level variables as JSON.",
				CustomType:  jsontypes.NormalizedType{},
				Computed:    true,
				Optional:    true,
				Sensitive:   varsAreSensitive,
			},
			"space_ids": schema.SetAttribute{
				Description: "The Kibana space IDs where this integration policy is available. When set, must match the space_ids of the referenced agent policy. If not set, will be inherited from the agent policy. Note: The order of space IDs does not matter as this is a set.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"inputs": schema.MapNestedAttribute{
				Description: "Integration inputs mapped by input ID.",
				CustomType:  NewInputsType(getInputsElementType()),
				Computed:    true,
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"enabled": schema.BoolAttribute{
							Description: "Enable the input.",
							Computed:    true,
							Optional:    true,
							Default:     booldefault.StaticBool(true),
						},
						"vars": schema.StringAttribute{
							Description: "Input-level variables as JSON.",
							CustomType:  jsontypes.NormalizedType{},
							Optional:    true,
							Sensitive:   varsAreSensitive,
						},
						"streams": schema.MapNestedAttribute{
							Description: "Input streams mapped by stream ID.",
							Optional:    true,
							Computed:    true,
							Default:     mapdefault.StaticValue(types.MapNull(getInputStreamType())),
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"enabled": schema.BoolAttribute{
										Description: "Enable the stream.",
										Computed:    true,
										Optional:    true,
										Default:     booldefault.StaticBool(true),
									},
									"vars": schema.StringAttribute{
										Description: "Stream-level variables as JSON.",
										CustomType:  jsontypes.NormalizedType{},
										Optional:    true,
										Sensitive:   varsAreSensitive,
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

func getInputsAttributeTypes() map[string]attr.Type {
	return getInputsElementType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

func getInputStreamType() attr.Type {
	return getInputsAttributeTypes()["streams"].(attr.TypeWithElementType).ElementType()
}
