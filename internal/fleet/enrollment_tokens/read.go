package enrollment_tokens

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d *enrollmentTokensDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model enrollmentTokensModel

	diags := req.Config.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.client.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	var tokens []kbapi.EnrollmentApiKey
	policyID := model.PolicyID.ValueString()

	// Determine space context for querying enrollment tokens
	var spaceID string
	if !model.SpaceIds.IsNull() && !model.SpaceIds.IsUnknown() {
		var spaceIDs []types.String
		diags = model.SpaceIds.ElementsAs(ctx, &spaceIDs, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(spaceIDs) > 0 {
			spaceID = spaceIDs[0].ValueString()
		}
	}

	// Query enrollment tokens with space context if needed
	if policyID == "" {
		// Get all tokens, with space awareness if specified
		if spaceID != "" && spaceID != "default" {
			tokens, diags = fleet.GetEnrollmentTokensInSpace(ctx, client, spaceID)
		} else {
			tokens, diags = fleet.GetEnrollmentTokens(ctx, client)
		}
	} else {
		// Get tokens by policy, with space awareness if specified
		if spaceID != "" && spaceID != "default" {
			tokens, diags = fleet.GetEnrollmentTokensByPolicyInSpace(ctx, client, policyID, spaceID)
		} else {
			tokens, diags = fleet.GetEnrollmentTokensByPolicy(ctx, client, policyID)
		}
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if policyID != "" {
		model.ID = types.StringValue(policyID)
	} else {
		hash, err := utils.StringToHash(client.URL)
		if err != nil {
			resp.Diagnostics.AddError(err.Error(), "")
			return
		}
		model.ID = types.StringPointerValue(hash)
	}

	diags = model.populateFromAPI(ctx, tokens)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}
