package agent_configuration

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AgentConfiguration holds the agent configuration.
type AgentConfiguration struct {
	ID                 types.String `tfsdk:"id"`
	ServiceName        types.String `tfsdk:"service_name"`
	ServiceEnvironment types.String `tfsdk:"service_environment"`
	AgentName          types.String `tfsdk:"agent_name"`
	Settings           types.Map    `tfsdk:"settings"`
}

func (ac *AgentConfiguration) SetIDFromService() string {
	parts := []string{ac.ServiceName.ValueString()}
	if !ac.ServiceEnvironment.IsNull() && !ac.ServiceEnvironment.IsUnknown() && ac.ServiceEnvironment.ValueString() != "" {
		parts = append(parts, ac.ServiceEnvironment.ValueString())
	}

	id := strings.Join(parts, ":")
	ac.ID = types.StringValue(id)
	return id
}
