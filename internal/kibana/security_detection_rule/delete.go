package securitydetectionrule

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *securityDetectionRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data Data

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse ID to get space_id and rule_id
	compID, diags := clients.CompositeIDFromStrFw(data.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	spaceID := compID.ClusterID

	// Get the rule using kbapi client
	kbClient, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting Kibana client",
			"Could not get Kibana OAPI client: "+err.Error(),
		)
		return
	}

	// Delete the rule
	uid, err := uuid.Parse(compID.ResourceID)
	if err != nil {
		resp.Diagnostics.AddError("ID was not a valid UUID", err.Error())
		return
	}
	params := &kbapi.DeleteRuleParams{
		Id: &uid,
	}

	response, err := kbClient.API.DeleteRuleWithResponse(ctx, spaceID, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting security detection rule",
			"Could not delete security detection rule: "+err.Error(),
		)
		return
	}

	if response.StatusCode() == 404 {
		// Rule was already deleted, which is fine
		return
	}

	if response.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"Error deleting security detection rule",
			fmt.Sprintf("API returned status %d: %s", response.StatusCode(), string(response.Body)),
		)
		return
	}
}
