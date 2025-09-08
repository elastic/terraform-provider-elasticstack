package detection_rule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// update handles updating an existing detection rule.
func (r *Resource) update(ctx context.Context, plan *tfsdk.Plan, state *tfsdk.State, newState *tfsdk.State, diags *diag.Diagnostics) {
	var planModel DetectionRuleModel
	diags.Append(plan.Get(ctx, &planModel)...)
	if diags.HasError() {
		return
	}

	var stateModel DetectionRuleModel
	diags.Append(state.Get(ctx, &stateModel)...)
	if diags.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating detection rule", map[string]interface{}{
		"id":       stateModel.ID.ValueString(),
		"rule_id":  stateModel.RuleID.ValueString(),
		"space_id": stateModel.SpaceID.ValueString(),
	})

	// TODO: Implement proper update when PUT /api/detection_engine/rules/{ruleId} is available in the generated client
	// For now, we'll treat this as unsupported and return an error
	diags.AddError(
		"Update not yet supported",
		"Detection rule updates are not yet implemented. The generated API client needs to be extended to support the PUT endpoint.",
	)
}
