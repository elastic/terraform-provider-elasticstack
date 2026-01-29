package slo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func (r *Resource) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			StateUpgrader: migrateV0ToV2,
		},
		1: {
			StateUpgrader: migrateV1ToV2,
		},
	}
}

func migrateV0ToV2(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	migrateV0ToV1(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare a new request with the upgraded state from v0 to v1.
	v1ToV2Req := resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{JSON: resp.DynamicValue.JSON},
	}

	migrateV1ToV2(ctx, v1ToV2Req, resp)
}

func migrateV0ToV1(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	if req.RawState == nil || req.RawState.JSON == nil {
		resp.Diagnostics.AddError("Invalid raw state", "Raw state or JSON is nil")
		return
	}

	// Default to returning the original state if no changes are needed.
	resp.DynamicValue = &tfprotov6.DynamicValue{JSON: req.RawState.JSON}

	var stateMap map[string]any
	if err := json.Unmarshal(req.RawState.JSON, &stateMap); err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal raw state", err.Error())
		return
	}

	groupBy, ok := stateMap["group_by"]
	if !ok || groupBy == nil {
		return
	}

	groupByStr, ok := groupBy.(string)
	if !ok {
		return
	}

	if len(groupByStr) == 0 {
		return
	}

	stateMap["group_by"] = []any{groupByStr}

	stateJSON, err := json.Marshal(stateMap)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal raw state", err.Error())
		return
	}

	resp.DynamicValue.JSON = stateJSON
	resp.Diagnostics.AddAttributeWarning(
		path.Root("group_by"),
		"Upgraded legacy group_by state",
		fmt.Sprintf("Upgraded legacy group_by from string to list: %q", groupByStr),
	)
}

func migrateV1ToV2(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	if req.RawState == nil || req.RawState.JSON == nil {
		resp.Diagnostics.AddError("Invalid raw state", "Raw state or JSON is nil")
		return
	}

	// Default to returning the original state if no changes are needed.
	resp.DynamicValue = &tfprotov6.DynamicValue{JSON: req.RawState.JSON}

	var stateMap map[string]any
	if err := json.Unmarshal(req.RawState.JSON, &stateMap); err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal raw state", err.Error())
		return
	}

	settings, ok := stateMap["settings"]
	if !ok || settings == nil {
		return
	}

	// If settings already has the new shape (object), nothing to do.
	if _, ok := settings.(map[string]any); ok {
		return
	}

	settingsList, ok := settings.([]any)
	if !ok {
		return
	}
	if len(settingsList) == 0 {
		stateMap["settings"] = nil
	} else {
		first, ok := settingsList[0].(map[string]any)
		if !ok {
			return
		}
		stateMap["settings"] = first
	}

	stateJSON, err := json.Marshal(stateMap)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal raw state", err.Error())
		return
	}

	resp.DynamicValue.JSON = stateJSON
	resp.Diagnostics.AddAttributeWarning(
		path.Root("settings"),
		"Upgraded legacy settings state",
		"Upgraded legacy settings from a nested block to a single nested attribute.",
	)
}
