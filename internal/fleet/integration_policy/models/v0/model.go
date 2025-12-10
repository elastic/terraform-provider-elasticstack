package v0

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// V0 model structures - used regular string types for JSON fields
type IntegrationPolicyModel struct {
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

type IntegrationPolicyInputModel struct {
	InputID     types.String `tfsdk:"input_id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	StreamsJson types.String `tfsdk:"streams_json"`
	VarsJson    types.String `tfsdk:"vars_json"`
}
