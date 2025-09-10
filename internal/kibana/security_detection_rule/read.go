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

	// For now, only query rules are fully supported in read operations
	// TODO: Add support for other rule types
	queryRule, err := response.AsSecurityDetectionsAPIQueryRule()
	if err != nil {
		// Try to determine the rule type for a better error message
		ruleTypeErr := "Could not parse rule as query rule: " + err.Error()

		// Try to parse as other rule types to determine the actual type
		if _, eqlErr := response.AsSecurityDetectionsAPIEqlRule(); eqlErr == nil {
			ruleTypeErr = "This appears to be an EQL rule, which is not yet fully supported for read operations."
		} else if _, esqlErr := response.AsSecurityDetectionsAPIEsqlRule(); esqlErr == nil {
			ruleTypeErr = "This appears to be an ESQL rule, which is not yet fully supported for read operations."
		} else if _, mlErr := response.AsSecurityDetectionsAPIMachineLearningRule(); mlErr == nil {
			ruleTypeErr = "This appears to be a Machine Learning rule, which is not yet fully supported for read operations."
		} else if _, newTermsErr := response.AsSecurityDetectionsAPINewTermsRule(); newTermsErr == nil {
			ruleTypeErr = "This appears to be a New Terms rule, which is not yet fully supported for read operations."
		} else if _, savedQueryErr := response.AsSecurityDetectionsAPIQueryRule(); savedQueryErr == nil {
			ruleTypeErr = "This appears to be a Saved Query rule, which is not yet fully supported for read operations."
		} else if _, threatMatchErr := response.AsSecurityDetectionsAPIThreatMatchRule(); threatMatchErr == nil {
			ruleTypeErr = "This appears to be a Threat Match rule, which is not yet fully supported for read operations."
		} else if _, thresholdErr := response.AsSecurityDetectionsAPIThresholdRule(); thresholdErr == nil {
			ruleTypeErr = "This appears to be a Threshold rule, which is not yet fully supported for read operations."
		}

		diags.AddError(
			"Error parsing rule response",
			ruleTypeErr+" Currently only query rules are fully supported.",
		)
		return nil, diags
	}

	return &queryRule, diags
}
