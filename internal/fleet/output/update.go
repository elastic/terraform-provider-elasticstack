package output

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	fleetutils "github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *outputResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planModel outputModel
	var stateModel outputModel

	diags := req.Plan.Get(ctx, &planModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &stateModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	body, diags := planModel.toAPIUpdateModel(ctx, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	outputID := planModel.OutputID.ValueString()

	// Read the existing spaces from state to avoid updating in a space where it's not yet visible
	spaceID, diags := fleetutils.GetOperationalSpaceFromState(ctx, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update using the operational space from STATE
	// The API will handle adding/removing output from spaces based on space_ids in body
	output, diags := fleet.UpdateOutput(ctx, client, outputID, spaceID, body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve sensitive fields from plan before populating from API
	// The Fleet API does not return sensitive field values for security reasons
	originalConfigYaml := planModel.ConfigYaml
	originalSsl := planModel.Ssl
	originalKafka := planModel.Kafka

	// Populate from API response
	// With Sets, we don't need order preservation - Terraform handles set comparison automatically
	diags = planModel.populateFromAPI(ctx, output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Restore sensitive fields so they are not lost in state
	// config_yaml is sensitive and not returned by the API
	planModel.ConfigYaml = originalConfigYaml

	// ssl.key is sensitive and not returned by the API
	planModel.Ssl = originalSsl

	// kafka.password is sensitive and not returned by the API
	planModel.Kafka = originalKafka

	diags = resp.State.Set(ctx, planModel)
	resp.Diagnostics.Append(diags...)
}
