package alerting_rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var planModel alertingRuleModel
	diags := req.Plan.Get(ctx, &planModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	kibanaClient, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	// Generate rule ID if not provided
	ruleID := planModel.RuleID.ValueString()
	if ruleID == "" {
		// Kibana will generate a UUID
		ruleID = "temp-id-" + planModel.Name.ValueString()
	}

	spaceID := planModel.SpaceID.ValueString()

	// Convert model to API request
	createReq, diags := planModel.toAPICreateModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the rule
	createResp, diags := kibana_oapi.CreateAlertingRule(ctx, kibanaClient, spaceID, ruleID, createReq)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the model with the created rule ID from response
	if createResp.JSON200 != nil {
		planModel.RuleID = types.StringValue(createResp.JSON200.Id)
	}

	// Read back to get complete state
	diags = read(ctx, kibanaClient, &planModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, planModel)
	resp.Diagnostics.Append(diags...)
}
