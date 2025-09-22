package agent_configuration

import (
	"context"
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *resourceAgentConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AgentConfiguration
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updatedState, diags := r.read(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if updatedState == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, updatedState)...)
}

func (r *resourceAgentConfiguration) read(ctx context.Context, state *AgentConfiguration) (*AgentConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics

	kibana, err := r.client.GetKibanaOapiClient()
	if err != nil {
		diags.AddError("Unable to get Kibana client", err.Error())
		return nil, diags
	}

	apiResp, err := kibana.API.GetAgentConfigurationsWithResponse(
		ctx,
		&kbapi.GetAgentConfigurationsParams{
			ElasticApiVersion: elasticAPIVersion,
		},
	)
	if err != nil {
		diags.AddError("Failed to get APM agent configurations", err.Error())
		return nil, diags
	}

	if httpDiags := diagutil.CheckHttpErrorFromFW(apiResp.HTTPResponse, "Failed to get APM agent configurations"); httpDiags.HasError() {
		diags.Append(httpDiags...)
		return nil, diags
	}

	if apiResp.JSON200 == nil {
		diags.AddError("Failed to get APM agent configurations from body", "Expected 200 response body to not be nil")
		return nil, diags
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
		return nil, diags
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

	settings, mapDiags := types.MapValueFrom(ctx, types.StringType, stringSettings)
	diags.Append(mapDiags...)
	if diags.HasError() {
		return nil, diags
	}
	state.Settings = settings

	return state, diags
}

func createAgentConfigIDfromAPI(config kbapi.APMUIAgentConfigurationObject) string {
	parts := []string{*config.Service.Name}
	if config.Service.Environment != nil && *config.Service.Environment != "" {
		parts = append(parts, *config.Service.Environment)
	}
	return strings.Join(parts, ":")
}
