package output

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *outputResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var planModel outputModel

	diags := req.Plan.Get(ctx, &planModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	body, diags := planModel.toAPICreateModel(ctx, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If space_ids is set, use space-aware CREATE request
	var spaceID string
	if !planModel.SpaceIds.IsNull() && !planModel.SpaceIds.IsUnknown() {
		var tempDiags diag.Diagnostics
		spaceIDs := utils.SetTypeAs[types.String](ctx, planModel.SpaceIds, path.Root("space_ids"), &tempDiags)
		if !tempDiags.HasError() && len(spaceIDs) > 0 {
			spaceID = spaceIDs[0].ValueString()
		}
	}

	output, diags := fleet.CreateOutput(ctx, client, spaceID, body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve sensitive field values from plan before populating from API
	// The Fleet API does not return sensitive field values for security reasons
	originalConfigYaml := planModel.ConfigYaml
	originalSslKey := extractSslKeyFromObject(ctx, planModel.Ssl)
	originalKafkaPassword := extractKafkaPasswordFromObject(ctx, planModel.Kafka)

	diags = planModel.populateFromAPI(ctx, output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Restore sensitive fields so they are not lost in state
	// config_yaml is sensitive and not returned by the API
	planModel.ConfigYaml = originalConfigYaml

	// ssl.key is sensitive and not returned by the API
	if originalSslKey != nil {
		planModel.Ssl = restoreSslKeyToObject(ctx, planModel.Ssl, *originalSslKey, &diags)
	}

	// kafka.password is sensitive and not returned by the API
	if originalKafkaPassword != nil {
		planModel.Kafka = restoreKafkaPasswordToObject(ctx, planModel.Kafka, *originalKafkaPassword, &diags)
	}

	diags = resp.State.Set(ctx, planModel)
	resp.Diagnostics.Append(diags...)
}
