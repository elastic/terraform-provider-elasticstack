package integration_policy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getSchemaV0() *schema.Schema {
	return &schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id":                  schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"policy_id":           schema.StringAttribute{Computed: true, Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()}},
			"name":                schema.StringAttribute{Required: true},
			"namespace":           schema.StringAttribute{Required: true},
			"agent_policy_id":     schema.StringAttribute{Required: true},
			"description":         schema.StringAttribute{Optional: true},
			"enabled":             schema.BoolAttribute{Computed: true, Optional: true, Default: booldefault.StaticBool(true)},
			"force":               schema.BoolAttribute{Optional: true},
			"integration_name":    schema.StringAttribute{Required: true},
			"integration_version": schema.StringAttribute{Required: true},
			"vars_json":           schema.StringAttribute{Computed: true, Optional: true, Sensitive: true},
		},
		Blocks: map[string]schema.Block{
			"input": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"input_id":     schema.StringAttribute{Required: true},
						"enabled":      schema.BoolAttribute{Computed: true, Optional: true, Default: booldefault.StaticBool(true)},
						"streams_json": schema.StringAttribute{Computed: true, Optional: true, Sensitive: true},
						"vars_json":    schema.StringAttribute{Computed: true, Optional: true, Sensitive: true},
					},
				},
			},
		},
	}
}

func getInputTypeV0() attr.Type {
	return getSchemaV0().Blocks["input"].Type().(attr.TypeWithElementType).ElementType()
}

type integrationPolicyModelV0 struct {
	ID                 types.String `tfsdk:"id"`
	PolicyID           types.String `tfsdk:"policy_id"`
	Name               types.String `tfsdk:"name"`
	Namespace          types.String `tfsdk:"namespace"`
	AgentPolicyID      types.String `tfsdk:"agent_policy_id"`
	Description        types.String `tfsdk:"description"`
	Enabled            types.Bool   `tfsdk:"enabled"`
	Force              types.Bool   `tfsdk:"force"`
	IntegrationName    types.String `tfsdk:"integration_name"`
	IntegrationVersion types.String `tfsdk:"integration_version"`
	Input              types.List   `tfsdk:"input"` //> integrationPolicyInputModelV0
	VarsJson           types.String `tfsdk:"vars_json"`
}

type integrationPolicyInputModelV0 struct {
	InputID     types.String `tfsdk:"input_id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	StreamsJson types.String `tfsdk:"streams_json"`
	VarsJson    types.String `tfsdk:"vars_json"`
}

// The schema between V0 and V1 is mostly the same, however vars_json and
// streams_json saved "" values to the state when null values were in the
// config. jsontypes.Normalized correctly states this is invalid JSON.
func upgradeV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var stateModel integrationPolicyModelV0

	diags := req.State.Get(ctx, &stateModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if varsJSON := stateModel.VarsJson.ValueStringPointer(); varsJSON != nil {
		if *varsJSON == "" {
			stateModel.VarsJson = types.StringNull()
		}
	}

	inputs := utils.ListTypeAs[integrationPolicyInputModelV0](ctx, stateModel.Input, path.Root("input"), &resp.Diagnostics)
	for index, input := range inputs {
		if varsJSON := input.VarsJson.ValueStringPointer(); varsJSON != nil {
			if *varsJSON == "" {
				input.VarsJson = types.StringNull()
			}
		}
		if streamsJSON := input.StreamsJson.ValueStringPointer(); streamsJSON != nil {
			if *streamsJSON == "" {
				input.StreamsJson = types.StringNull()
			}
		}
		inputs[index] = input
	}

	stateModel.Input = utils.ListValueFrom(ctx, inputs, getInputTypeV0(), path.Root("input"), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, stateModel)
	resp.Diagnostics.Append(diags...)
}
