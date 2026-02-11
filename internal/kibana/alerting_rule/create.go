package alerting_rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan alertingRuleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Expected configured API client")
		return
	}

	// Get server version to validate version-specific features
	serverVersion, versionDiags := r.client.ServerVersion(ctx)
	if versionDiags.HasError() {
		resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(versionDiags)...)
		return
	}

	// Convert to API model (includes version-specific validation)
	rule, diags := plan.toAPIModel(ctx, serverVersion)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	oapiClient, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Kibana client", err.Error())
		return
	}

	createdRule, createDiags := kibana_oapi.CreateAlertingRule(ctx, oapiClient, rule.SpaceID, rule)
	resp.Diagnostics.Append(createDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Initialize plan with rule ID and space ID from created rule for re-reading
	resp.Diagnostics.Append(plan.populateFromAPI(ctx, createdRule)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Re-read rule from API to get the authoritative state
	// (sometimes create response differs from what's actually stored)
	exists, readDiags := r.readRuleFromAPI(ctx, &plan)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !exists {
		resp.Diagnostics.AddError("Rule not found after creation", "The alerting rule was created but could not be read back from the API")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
