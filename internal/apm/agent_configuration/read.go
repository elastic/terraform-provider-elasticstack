package agent_configuration

import (
	"context"
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *resourceAgentConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AgentConfiguration
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	kibana, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Unable to get Kibana client", err.Error())
		return
	}

	apiResp, err := kibana.API.GetAgentConfigurationsWithResponse(ctx, &kbapi.GetAgentConfigurationsParams{})
	if err != nil {
		resp.Diagnostics.AddError("Failed to get APM agent configurations", err.Error())
		return
	}

	if diags := utils.CheckHttpErrorFromFW(apiResp.HTTPResponse, "Failed to get APM agent configurations"); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError("Failed to get APM agent configurations from body", "Expected 200 response body to not be nil")
		return
	}

	idFromState := state.ID.ValueString()
	var foundConfig *kbapi.APMUIAgentConfigurationObject
	for _, config := range *apiResp.JSON200.Configurations {
		if config.Service.Name == nil {
			continue
		}
		idFromAPI := createAgentConfigIDfromAPI(config)
		if idFromAPI == idFromState {
			foundConfig = &config
			break
		}
	}

	if foundConfig == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(idFromState)
	state.ServiceName = types.StringPointerValue(foundConfig.Service.Name)
	state.ServiceEnvironment = types.StringPointerValue(foundConfig.Service.Environment)
	state.AgentName = types.StringPointerValue(foundConfig.AgentName)

	stringSettings := make(map[string]interface{})
	if foundConfig.Settings != nil {
		for k, v := range foundConfig.Settings {
			stringSettings[k] = fmt.Sprintf("%v", v)
		}
	}

	settings, diags := types.MapValueFrom(ctx, types.StringType, stringSettings)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Settings = settings

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func createAgentConfigIDfromAPI(config kbapi.APMUIAgentConfigurationObject) string {
	parts := []string{*config.Service.Name}
	if config.Service.Environment != nil && *config.Service.Environment != "" {
		parts = append(parts, *config.Service.Environment)
	}
	return strings.Join(parts, ":")
}
