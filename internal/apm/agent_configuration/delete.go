package agent_configuration

import (
	"context"
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *resourceAgentConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

	idParts := strings.Split(state.ID.ValueString(), ":")
	serviceName := idParts[0]
	var serviceEnv *string
	if len(idParts) > 1 {
		serviceEnv = &idParts[1]
	}

	deleteReqBody := kbapi.APMUIDeleteServiceObject{
		Service: &kbapi.APMUIServiceObject{
			Name:        &serviceName,
			Environment: serviceEnv,
		},
	}
	apiResp, err := kibana.API.DeleteAgentConfiguration(
		ctx,
		&kbapi.DeleteAgentConfigurationParams{
			ElasticApiVersion: elasticAPIVersion,
		},
		deleteReqBody,
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete APM agent configuration", err.Error())
		return
	}
	defer apiResp.Body.Close()

	if diags := utils.CheckHttpErrorFromFW(apiResp, "Failed to delete APM agent configuration"); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("Deleted APM agent configuration with ID: %s", state.ID.ValueString()))

	resp.State.RemoveResource(ctx)
}
