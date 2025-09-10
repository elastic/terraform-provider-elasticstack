package security_detection_rule

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *securityDetectionRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SecurityDetectionRuleData

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the rule using kbapi client
	kbClient, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting Kibana client",
			"Could not get Kibana OAPI client: "+err.Error(),
		)
		return
	}

	// Build the create request
	createProps, diags := data.toCreateProps(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the rule
	response, err := kbClient.API.CreateRuleWithResponse(ctx, createProps)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating security detection rule",
			"Could not create security detection rule: "+err.Error(),
		)
		return
	}

	if response.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"Error creating security detection rule",
			fmt.Sprintf("API returned status %d: %s", response.StatusCode(), string(response.Body)),
		)
		return
	}

	// Parse the response to get the ID, then use Read logic for consistency
	ruleResponse, diags := r.parseRuleResponse(ctx, response.JSON200)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the ID based on the created rule
	ruleId, err := extractRuleId(ruleResponse)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error extracting rule ID",
			"Could not extract ID from created rule: "+err.Error(),
		)
		return
	}

	compId := clients.CompositeId{
		ClusterId:  data.SpaceId.ValueString(),
		ResourceId: ruleId,
	}
	data.Id = types.StringValue(compId.String())

	// Use Read logic to populate the state with fresh data from the API
	readReq := resource.ReadRequest{
		State: resp.State,
	}
	var readResp resource.ReadResponse
	readReq.State.Set(ctx, &data)
	r.Read(ctx, readReq, &readResp)

	resp.Diagnostics.Append(readResp.Diagnostics...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State = readResp.State
}
