package alerting_rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan alertingRuleModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Kibana client", err.Error())
		return
	}

	// Generate rule ID if not provided
	ruleID := plan.RuleID.ValueString()
	if ruleID == "" {
		ruleID = uuid.New().String()
	}

	// Convert to API request
	apiReq, diags := plan.toAPICreateRequest(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the rule
	ruleResp, diags := kibana_oapi.CreateAlertingRule(ctx, client, ruleID, apiReq)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate the model from response
	spaceID := plan.SpaceID.ValueString()
	diags = plan.populateFromAPI(ctx, ruleResp, spaceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the composite ID
	plan.ID = types.StringValue((&clients.CompositeId{ClusterId: spaceID, ResourceId: ruleResp.Id}).String())

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}
