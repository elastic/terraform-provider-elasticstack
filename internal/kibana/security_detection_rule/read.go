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

	// Use the extracted read method
	readData, diags := r.read(ctx, compId.ResourceId, compId.ClusterId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if the rule was found (nil data indicates 404)
	if readData == nil {
		// Rule was deleted outside of Terraform
		resp.State.RemoveResource(ctx)
		return
	}

	// Set the composite ID and state
	readData.Id = data.Id
	resp.Diagnostics.Append(resp.State.Set(ctx, readData)...)
}

// read extracts the core functionality of reading a security detection rule
func (r *securityDetectionRuleResource) read(ctx context.Context, resourceId, spaceId string) (*SecurityDetectionRuleData, diag.Diagnostics) {
	var diags diag.Diagnostics

	data := &SecurityDetectionRuleData{}
	data.initializeAllFieldsToDefaults(ctx, &diags)

	// Get the rule using kbapi client
	kbClient, err := r.client.GetKibanaOapiClient()
	if err != nil {
		diags.AddError(
			"Error getting Kibana client",
			"Could not get Kibana OAPI client: "+err.Error(),
		)
		return nil, diags
	}

	// Read the rule
	uid, err := uuid.Parse(resourceId)
	if err != nil {
		diags.AddError("ID was not a valid UUID", err.Error())
		return nil, diags
	}
	ruleObjectId := kbapi.SecurityDetectionsAPIUUID(uid)
	params := &kbapi.ReadRuleParams{
		Id: &ruleObjectId,
	}

	response, err := kbClient.API.ReadRuleWithResponse(ctx, spaceId, params)
	if err != nil {
		diags.AddError(
			"Error reading security detection rule",
			"Could not read security detection rule: "+err.Error(),
		)
		return nil, diags
	}

	if response.StatusCode() == 404 {
		// Rule was deleted - return nil to indicate this
		return nil, diags
	}

	if response.StatusCode() != 200 {
		diags.AddError(
			"Error reading security detection rule",
			fmt.Sprintf("API returned status %d: %s", response.StatusCode(), string(response.Body)),
		)
		return nil, diags
	}

	// Parse the response
	updateDiags := data.updateFromRule(ctx, response.JSON200)
	diags.Append(updateDiags...)
	if diags.HasError() {
		return nil, diags
	}

	// Ensure space_id is set correctly
	data.SpaceId = types.StringValue(spaceId)

	compId := clients.CompositeId{
		ResourceId: resourceId,
		ClusterId:  spaceId,
	}

	data.Id = types.StringValue(compId.String())

	return data, diags
}
