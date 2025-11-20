package integration_policy

import (
	"context"
	_ "embed"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type integrationPolicyModelV1 struct {
	ID                 types.String         `tfsdk:"id"`
	PolicyID           types.String         `tfsdk:"policy_id"`
	Name               types.String         `tfsdk:"name"`
	Namespace          types.String         `tfsdk:"namespace"`
	AgentPolicyID      types.String         `tfsdk:"agent_policy_id"`
	AgentPolicyIDs     types.List           `tfsdk:"agent_policy_ids"`
	Description        types.String         `tfsdk:"description"`
	Enabled            types.Bool           `tfsdk:"enabled"`
	Force              types.Bool           `tfsdk:"force"`
	IntegrationName    types.String         `tfsdk:"integration_name"`
	IntegrationVersion types.String         `tfsdk:"integration_version"`
	OutputID           types.String         `tfsdk:"output_id"`
	Input              types.List           `tfsdk:"input"` //> integrationPolicyInputModel
	VarsJson           jsontypes.Normalized `tfsdk:"vars_json"`
	SpaceIds           types.Set            `tfsdk:"space_ids"`
}

type integrationPolicyInputModelV1 struct {
	InputID     types.String         `tfsdk:"input_id"`
	Enabled     types.Bool           `tfsdk:"enabled"`
	StreamsJson jsontypes.Normalized `tfsdk:"streams_json"`
	VarsJson    jsontypes.Normalized `tfsdk:"vars_json"`
}

func (m integrationPolicyModelV1) toV2(ctx context.Context) (integrationPolicyModel, diag.Diagnostics) {
	// Convert V1 model to V2 model
	stateModelV2 := integrationPolicyModel{
		ID:                 m.ID,
		PolicyID:           m.PolicyID,
		Name:               m.Name,
		Namespace:          m.Namespace,
		AgentPolicyID:      m.AgentPolicyID,
		AgentPolicyIDs:     m.AgentPolicyIDs,
		Description:        m.Description,
		Enabled:            m.Enabled,
		Force:              m.Force,
		IntegrationName:    m.IntegrationName,
		IntegrationVersion: m.IntegrationVersion,
		OutputID:           m.OutputID,
		SpaceIds:           m.SpaceIds,
		VarsJson:           m.VarsJson,
	}

	// Convert inputs from V1 to V2
	var diags diag.Diagnostics
	inputsV1 := utils.ListTypeAs[integrationPolicyInputModelV1](ctx, m.Input, path.Root("input"), &diags)
	inputsV2 := make(map[string]integrationPolicyInputsModel, len(inputsV1))

	for _, inputV1 := range inputsV1 {
		id := inputV1.InputID.ValueString()
		streams, d := updateStreamsV1ToV2(ctx, inputV1.StreamsJson, id)
		diags.Append(d...)
		if diags.HasError() {
			return stateModelV2, diags
		}

		inputsV2[id] = integrationPolicyInputsModel{
			Enabled:  inputV1.Enabled,
			Vars:     inputV1.VarsJson,
			Streams:  streams,
			Defaults: types.ObjectNull(getInputDefaultsAttrTypes()),
		}
	}

	inputsValue, d := NewInputsValueFrom(ctx, getInputsElementType(), inputsV2)
	diags.Append(d...)

	stateModelV2.Inputs = inputsValue
	return stateModelV2, diags
}

// The schema between V1 and V2 is mostly the same. Except for:
// * The input block was moved to an map attribute.
// * The streams attribute inside the input block was also moved to a map attribute.
// This upgrader translates the old list structures into the new map structures.
func upgradeV1ToV2(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var stateModelV1 integrationPolicyModelV1

	diags := req.State.Get(ctx, &stateModelV1)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateModelV2, diags := stateModelV1.toV2(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, stateModelV2)
	resp.Diagnostics.Append(diags...)
}

func updateStreamsV1ToV2(ctx context.Context, v1 jsontypes.Normalized, inputID string) (types.Map, diag.Diagnostics) {
	if !utils.IsKnown(v1) {
		return types.MapNull(getInputStreamType()), nil
	}

	var apiStreams map[string]kbapi.PackagePolicyInputStream
	diags := v1.Unmarshal(&apiStreams)
	if diags.HasError() {
		return types.MapNull(getInputStreamType()), diags
	}

	if len(apiStreams) == 0 {
		return types.MapNull(getInputStreamType()), nil
	}

	streams := make(map[string]integrationPolicyInputStreamModel)
	for streamID, streamData := range apiStreams {
		streamModel := integrationPolicyInputStreamModel{
			Enabled: types.BoolPointerValue(streamData.Enabled),
			Vars:    utils.MapToNormalizedType(utils.Deref(streamData.Vars), path.Root("inputs").AtMapKey(inputID).AtName("streams").AtMapKey(streamID).AtName("vars"), &diags),
		}

		streams[streamID] = streamModel
	}

	return types.MapValueFrom(ctx, getInputStreamType(), streams)
}

func getSchemaV1() *schema.Schema {
	return &schema.Schema{
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
				Sensitive:   true,
			},
			"space_ids": schema.SetAttribute{
				Description: "The Kibana space IDs where this integration policy is available. When set, must match the space_ids of the referenced agent policy. If not set, will be inherited from the agent policy. Note: The order of space IDs does not matter as this is a set.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
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
