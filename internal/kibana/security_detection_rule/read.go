package security_detection_rule

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *securityDetectionRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SecurityDetectionRuleData

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse ID to get space_id and rule_id
	compId, diags := clients.CompositeIdFromStrFw(data.Id.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the rule using kbapi client
	kbClient, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting Kibana client",
			"Could not get Kibana OAPI client: "+err.Error(),
		)
		return
	}

	// Read the rule
	uid, err := uuid.Parse(compId.ResourceId)
	if err != nil {
		resp.Diagnostics.AddError("ID was not a valid UUID", err.Error())
		return
	}
	ruleObjectId := kbapi.SecurityDetectionsAPIRuleObjectId(uid)
	params := &kbapi.ReadRuleParams{
		Id: &ruleObjectId,
	}

	response, err := kbClient.API.ReadRuleWithResponse(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading security detection rule",
			"Could not read security detection rule: "+err.Error(),
		)
		return
	}

	if response.StatusCode() == 404 {
		// Rule was deleted outside of Terraform
		resp.State.RemoveResource(ctx)
		return
	}

	if response.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"Error reading security detection rule",
			fmt.Sprintf("API returned status %d: %s", response.StatusCode(), string(response.Body)),
		)
		return
	}

	// Parse the response
	ruleResponse, diags := r.parseRuleResponse(ctx, response.JSON200)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the data with response values
	diags = data.updateFromRule(ctx, ruleResponse)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Ensure space_id is set correctly
	data.SpaceId = types.StringValue(compId.ClusterId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *securityDetectionRuleResource) parseRuleResponse(ctx context.Context, response *kbapi.SecurityDetectionsAPIRuleResponse) (*kbapi.SecurityDetectionsAPIQueryRule, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Since we only support query rules for now, try to parse as query rule
	queryRule, err := response.AsSecurityDetectionsAPIQueryRule()
	if err != nil {
		diags.AddError(
			"Error parsing rule response",
			"Could not parse rule as query rule: "+err.Error(),
		)
		return nil, diags
	}

	return &queryRule, diags
}
