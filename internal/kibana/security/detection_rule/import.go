package detection_rule

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// importState handles importing an existing detection rule.
func (r *Resource) importState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Expected import format: "space_id/rule_id" or just "rule_id" for default space
	parts := strings.Split(req.ID, "/")

	var spaceID, ruleID string

	switch len(parts) {
	case 1:
		// Just rule_id provided, use default space
		spaceID = "default"
		ruleID = parts[0]
	case 2:
		// space_id/rule_id format
		spaceID = parts[0]
		ruleID = parts[1]
	default:
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			"Expected format: 'rule_id' or 'space_id/rule_id'. For example: 'default/my-rule-id' or just 'my-rule-id' for default space.",
		)
		return
	}

	if ruleID == "" {
		resp.Diagnostics.AddError(
			"Invalid rule ID",
			"Rule ID cannot be empty.",
		)
		return
	}

	// Set the initial state with the import information
	// The read operation will populate the rest of the fields
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("rule_id"), types.StringValue(ruleID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("space_id"), types.StringValue(spaceID))...)

	// Set a placeholder ID - this will be updated by the read operation
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(ruleID))...)
}
