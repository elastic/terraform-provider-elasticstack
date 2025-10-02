package integration_policy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// V0 model structures - used regular string types for JSON fields
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

// The schema between V0 and V1 is mostly the same, however vars_json and
// streams_json saved "" values to the state when null values were in the
// config. jsontypes.Normalized correctly states this is invalid JSON.
func upgradeV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var stateModelV0 integrationPolicyModelV0

	diags := req.State.Get(ctx, &stateModelV0)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert V0 model to V1 model
	stateModelV1 := integrationPolicyModel{
		ID:                 stateModelV0.ID,
		PolicyID:           stateModelV0.PolicyID,
		Name:               stateModelV0.Name,
		Namespace:          stateModelV0.Namespace,
		AgentPolicyID:      stateModelV0.AgentPolicyID,
		AgentPolicyIDs:     types.ListNull(types.StringType), // V0 didn't have agent_policy_ids
		Description:        stateModelV0.Description,
		Enabled:            stateModelV0.Enabled,
		Force:              stateModelV0.Force,
		IntegrationName:    stateModelV0.IntegrationName,
		IntegrationVersion: stateModelV0.IntegrationVersion,
	}

	// Convert vars_json from string to normalized JSON type
	if varsJSON := stateModelV0.VarsJson.ValueStringPointer(); varsJSON != nil {
		if *varsJSON == "" {
			stateModelV1.VarsJson = jsontypes.NewNormalizedNull()
		} else {
			stateModelV1.VarsJson = jsontypes.NewNormalizedValue(*varsJSON)
		}
	} else {
		stateModelV1.VarsJson = jsontypes.NewNormalizedNull()
	}

	// Convert inputs from V0 to V1
	inputsV0 := utils.ListTypeAs[integrationPolicyInputModelV0](ctx, stateModelV0.Input, path.Root("input"), &resp.Diagnostics)
	var inputsV1 []integrationPolicyInputModel

	for _, inputV0 := range inputsV0 {
		inputV1 := integrationPolicyInputModel{
			InputID: inputV0.InputID,
			Enabled: inputV0.Enabled,
		}

		// Convert vars_json
		if varsJSON := inputV0.VarsJson.ValueStringPointer(); varsJSON != nil {
			if *varsJSON == "" {
				inputV1.VarsJson = jsontypes.NewNormalizedNull()
			} else {
				inputV1.VarsJson = jsontypes.NewNormalizedValue(*varsJSON)
			}
		} else {
			inputV1.VarsJson = jsontypes.NewNormalizedNull()
		}

		// Convert streams_json
		if streamsJSON := inputV0.StreamsJson.ValueStringPointer(); streamsJSON != nil {
			if *streamsJSON == "" {
				inputV1.StreamsJson = jsontypes.NewNormalizedNull()
			} else {
				inputV1.StreamsJson = jsontypes.NewNormalizedValue(*streamsJSON)
			}
		} else {
			inputV1.StreamsJson = jsontypes.NewNormalizedNull()
		}

		inputsV1 = append(inputsV1, inputV1)
	}

	stateModelV1.Input = utils.ListValueFrom(ctx, inputsV1, getInputTypeV1(), path.Root("input"), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, stateModelV1)
	resp.Diagnostics.Append(diags...)
}
