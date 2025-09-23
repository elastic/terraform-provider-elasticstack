package integration_policy

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

//go:embed resource-description.md
var integrationPolicyDescription string

func (r *integrationPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchemaV1()
}

func getSchemaV1() schema.Schema {
	return schema.Schema{
		Version:     1,
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
					listvalidator.ConflictsWith(path.Root("agent_policy_ids").Expression()),
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
			"vars_json": schema.StringAttribute{
				Description: "Integration-level variables as JSON.",
				CustomType:  jsontypes.NormalizedType{},
				Computed:    true,
				Optional:    true,
				Sensitive:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"input": schema.ListNestedBlock{
				Description: "Integration inputs.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"input_id": schema.StringAttribute{
							Description: "The identifier of the input.",
							Required:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Enable the input.",
							Computed:    true,
							Optional:    true,
							Default:     booldefault.StaticBool(true),
						},
						"streams_json": schema.StringAttribute{
							Description: "Input streams as JSON.",
							CustomType:  jsontypes.NormalizedType{},
							Computed:    true,
							Optional:    true,
							Sensitive:   true,
						},
						"vars_json": schema.StringAttribute{
							Description: "Input variables as JSON.",
							CustomType:  jsontypes.NormalizedType{},
							Computed:    true,
							Optional:    true,
							Sensitive:   true,
						},
					},
				},
			},
		},
	}
}

func getInputTypeV1() attr.Type {
	return getSchemaV1().Blocks["input"].Type().(attr.TypeWithElementType).ElementType()
}
