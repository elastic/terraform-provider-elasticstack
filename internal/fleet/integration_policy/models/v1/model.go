package v1

import (
	"context"

	v0 "github.com/elastic/terraform-provider-elasticstack/internal/fleet/integration_policy/models/v0"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type IntegrationPolicyModel struct {
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

type IntegrationPolicyInputModel struct {
	InputID     types.String         `tfsdk:"input_id"`
	Enabled     types.Bool           `tfsdk:"enabled"`
	StreamsJson jsontypes.Normalized `tfsdk:"streams_json"`
	VarsJson    jsontypes.Normalized `tfsdk:"vars_json"`
}

// The schema between V0 and V1 is mostly the same, however vars_json and
// streams_json saved "" values to the state when null values were in the
// config. jsontypes.Normalized correctly states this is invalid JSON.
func NewFromV0(ctx context.Context, m v0.IntegrationPolicyModel) (IntegrationPolicyModel, diag.Diagnostics) {
	// Convert V0 model to V1 model
	stateModelV1 := IntegrationPolicyModel{
		ID:                 m.ID,
		PolicyID:           m.PolicyID,
		Name:               m.Name,
		Namespace:          m.Namespace,
		AgentPolicyID:      m.AgentPolicyID,
		AgentPolicyIDs:     types.ListNull(types.StringType), // V0 didn't have agent_policy_ids
		Description:        m.Description,
		Enabled:            m.Enabled,
		Force:              m.Force,
		IntegrationName:    m.IntegrationName,
		IntegrationVersion: m.IntegrationVersion,
		SpaceIds:           types.SetNull(types.StringType), // V0 didn't have space_ids
	}

	// Convert vars_json from string to normalized JSON type
	if varsJSON := m.VarsJson.ValueStringPointer(); varsJSON != nil {
		if *varsJSON == "" {
			stateModelV1.VarsJson = jsontypes.NewNormalizedNull()
		} else {
			stateModelV1.VarsJson = jsontypes.NewNormalizedValue(*varsJSON)
		}
	} else {
		stateModelV1.VarsJson = jsontypes.NewNormalizedNull()
	}

	// Convert inputs from V0 to V1
	var diags diag.Diagnostics
	inputsV0 := utils.ListTypeAs[v0.IntegrationPolicyInputModel](ctx, m.Input, path.Root("input"), &diags)
	var inputsV1 []IntegrationPolicyInputModel

	for _, inputV0 := range inputsV0 {
		inputV1 := IntegrationPolicyInputModel{
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

	stateModelV1.Input = utils.ListValueFrom(ctx, inputsV1, GetInputType(), path.Root("input"), &diags)
	return stateModelV1, diags
}
