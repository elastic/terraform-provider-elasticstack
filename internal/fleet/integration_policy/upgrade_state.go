package integration_policy

import (
	"context"

	v0 "github.com/elastic/terraform-provider-elasticstack/internal/fleet/integration_policy/models/v0"
	v1 "github.com/elastic/terraform-provider-elasticstack/internal/fleet/integration_policy/models/v1"
	v2 "github.com/elastic/terraform-provider-elasticstack/internal/fleet/integration_policy/models/v2"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *integrationPolicyResource) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {PriorSchema: v0.GetSchema(), StateUpgrader: UpgradeV0ToV2},
		1: {PriorSchema: v1.GetSchema(), StateUpgrader: UpgradeV1ToV2},
	}
}

// This function first upgrades v0 to v1, and then re-uses the v1 to v2 upgrader.
func UpgradeV0ToV2(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var stateModelV0 v0.IntegrationPolicyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &stateModelV0)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateModelV1, diags := v1.NewFromV0(ctx, stateModelV0)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateModelV2, diags := v2.NewFromV1(ctx, stateModelV1)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, stateModelV2)
	resp.Diagnostics.Append(diags...)
}

// The schema between V1 and V2 is mostly the same. Except for:
// * The input block was moved to an map attribute.
// * The streams attribute inside the input block was also moved to a map attribute.
// This upgrader translates the old list structures into the new map structures.
func UpgradeV1ToV2(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var stateModelV1 v1.IntegrationPolicyModel

	diags := req.State.Get(ctx, &stateModelV1)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateModelV2, diags := v2.NewFromV1(ctx, stateModelV1)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, stateModelV2)
	resp.Diagnostics.Append(diags...)
}
