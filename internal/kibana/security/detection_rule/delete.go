package detection_rule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// delete handles deleting an existing detection rule.
func (r *Resource) delete(ctx context.Context, state *tfsdk.State, diags *diag.Diagnostics) {
	var stateModel DetectionRuleModel
	diags.Append(state.Get(ctx, &stateModel)...)
	if diags.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting detection rule", map[string]interface{}{
		"id":       stateModel.ID.ValueString(),
		"rule_id":  stateModel.RuleID.ValueString(),
		"space_id": stateModel.SpaceID.ValueString(),
	})

	// TODO: Implement proper delete when DELETE /api/detection_engine/rules/{ruleId} is available in the generated client
	// For now, we'll treat this as unsupported and return an error
	diags.AddError(
		"Delete not yet supported",
		"Detection rule deletion is not yet implemented. The generated API client needs to be extended to support the DELETE endpoint.",
	)
}
