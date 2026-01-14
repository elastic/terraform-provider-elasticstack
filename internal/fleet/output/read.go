package output

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	fleetutils "github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *outputResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateModel outputModel

	diags := req.State.Get(ctx, &stateModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	outputID := stateModel.OutputID.ValueString()

	// Read the existing spaces from state to determine where to query
	spaceID, diags := fleetutils.GetOperationalSpaceFromState(ctx, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Query using the operational space from STATE
	output, diags := fleet.GetOutput(ctx, client, outputID, spaceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		resp.State.RemoveResource(ctx)
		return
	}

	if output == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Preserve sensitive field values from state before populating from API
	// The Fleet API does not return sensitive field values for security reasons
	originalConfigYaml := stateModel.ConfigYaml
	originalSslKey := extractSslKeyFromObject(ctx, stateModel.Ssl)
	originalKafkaPassword := extractKafkaPasswordFromObject(ctx, stateModel.Kafka)

	diags = stateModel.populateFromAPI(ctx, output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Restore sensitive fields so they are not lost on refresh
	// Only restore if they were known in the previous state (not during import)
	// config_yaml is sensitive and not returned by the API
	if utils.IsKnown(originalConfigYaml) {
		stateModel.ConfigYaml = originalConfigYaml
	}

	// ssl.key is sensitive and not returned by the API
	if originalSslKey != nil && utils.IsKnown(*originalSslKey) {
		stateModel.Ssl = restoreSslKeyToObject(ctx, stateModel.Ssl, *originalSslKey, &diags)
	}

	// kafka.password is sensitive and not returned by the API
	if originalKafkaPassword != nil && utils.IsKnown(*originalKafkaPassword) {
		stateModel.Kafka = restoreKafkaPasswordToObject(ctx, stateModel.Kafka, *originalKafkaPassword, &diags)
	}

	diags = resp.State.Set(ctx, stateModel)
	resp.Diagnostics.Append(diags...)
}
