package integration_policy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *integrationPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var planModel integrationPolicyModel

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

	feat, diags := r.buildFeatures(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := planModel.toAPIModel(ctx, false, feat)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine space context for creating the package policy
	// The package policy must be created in the same space as the agent policy it references
	var spaceID string
	if !planModel.SpaceIds.IsNull() && !planModel.SpaceIds.IsUnknown() {
		// Explicit space_ids provided - use the first one
		var tempDiags diag.Diagnostics
		spaceIDs := utils.SetTypeAs[types.String](ctx, planModel.SpaceIds, path.Root("space_ids"), &tempDiags)
		if !tempDiags.HasError() && len(spaceIDs) > 0 {
			spaceID = spaceIDs[0].ValueString()
		}
	}

	// Create package policy with appropriate space context
	var policy *kbapi.PackagePolicy
	if spaceID != "" && spaceID != "default" {
		policy, diags = fleet.CreatePackagePolicyInSpace(ctx, client, spaceID, body)
	} else {
		policy, diags = fleet.CreatePackagePolicy(ctx, client, body)
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = HandleReqRespSecrets(ctx, body, policy, resp.Private)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Remember if the user configured input in the plan
	planHadInput := utils.IsKnown(planModel.Input) && !planModel.Input.IsNull() && len(planModel.Input.Elements()) > 0

	diags = planModel.populateFromAPI(ctx, policy)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If plan didn't have input configured, ensure we don't add it now
	// This prevents "Provider produced inconsistent result" errors
	if !planHadInput {
		planModel.Input = types.ListNull(getInputTypeV1())
	}

	diags = resp.State.Set(ctx, planModel)
	resp.Diagnostics.Append(diags...)
}
