package security_detection_rule

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *securityDetectionRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SecurityDetectionRuleData

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
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

	// Build the update request
	updateProps, diags := data.toUpdateProps(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the rule
	response, err := kbClient.API.UpdateRuleWithResponse(ctx, updateProps)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating security detection rule",
			"Could not update security detection rule: "+err.Error(),
		)
		return
	}

	if response.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"Error updating security detection rule",
			fmt.Sprintf("API returned status %d: %s", response.StatusCode(), string(response.Body)),
		)
		return
	}

	// Parse ID to get space_id and rule_id
	compId, resourceIdDiags := clients.CompositeIdFromStrFw(data.Id.ValueString())
	diags.Append(resourceIdDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	uid, err := uuid.Parse(compId.ResourceId)

	readData, diags := r.read(ctx, uid.String(), data.SpaceId.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &readData)...)
}
