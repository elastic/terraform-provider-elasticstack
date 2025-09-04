package detection_rule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// read handles reading an existing detection rule.
func (r *Resource) read(ctx context.Context, state *tfsdk.State, newState *tfsdk.State, diags *diag.Diagnostics) {
	var stateModel DetectionRuleModel
	diags.Append(state.Get(ctx, &stateModel)...)
	if diags.HasError() {
		return
	}

	// For now, we'll implement a basic read that preserves the current state
	// since the generated API client doesn't include individual rule GET endpoints yet.
	// In a complete implementation, we would fetch the current rule state from the API
	// and update any computed fields.

	tflog.Debug(ctx, "Reading detection rule (basic implementation)", map[string]interface{}{
		"id":       stateModel.ID.ValueString(),
		"rule_id":  stateModel.RuleID.ValueString(),
		"space_id": stateModel.SpaceID.ValueString(),
	})

	// For now, just preserve the existing state
	// TODO: Implement proper API call when GET /api/detection_engine/rules/{ruleId} is available
	diags.Append(newState.Set(ctx, stateModel)...)
}
