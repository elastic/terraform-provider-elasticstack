package alerting_rule

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func (r *Resource) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		// Version 0 is the SDKv2 state format - migrate to PFW format
		0: {
			StateUpgrader: migrateV0ToV1,
		},
	}
}

// migrateV0ToV1 handles migration from SDKv2 state format to Plugin Framework state format.
// The main differences are:
// - JSON fields may need normalization for jsontypes.Normalized
// - notify_when may be empty string instead of null
// - throttle may be empty string instead of null
// - frequency, alerts_filter, and timeframe change from lists to single objects
func migrateV0ToV1(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	if req.RawState == nil || req.RawState.JSON == nil {
		resp.Diagnostics.AddError("Invalid raw state", "Raw state or JSON is nil")
		return
	}

	// Default to returning the original state if no changes are needed
	resp.DynamicValue = &tfprotov6.DynamicValue{JSON: req.RawState.JSON}

	var stateMap map[string]any
	if err := json.Unmarshal(req.RawState.JSON, &stateMap); err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal raw state", err.Error())
		return
	}

	modified := false

	// Handle notify_when: convert empty string to null for proper handling
	if notifyWhen, ok := stateMap["notify_when"]; ok {
		if notifyWhenStr, ok := notifyWhen.(string); ok && notifyWhenStr == "" {
			stateMap["notify_when"] = nil
			modified = true
		}
	}

	// Handle throttle: convert empty string to null
	if throttle, ok := stateMap["throttle"]; ok {
		if throttleStr, ok := throttle.(string); ok && throttleStr == "" {
			stateMap["throttle"] = nil
			modified = true
		}
	}

	// Handle actions: convert frequency, alerts_filter, and timeframe from lists to objects
	if actions, ok := stateMap["actions"].([]any); ok {
		for _, actionAny := range actions {
			if action, ok := actionAny.(map[string]any); ok {
				// Convert frequency from list to object
				if freq, ok := action["frequency"].([]any); ok {
					if len(freq) > 0 {
						action["frequency"] = freq[0]
					} else {
						action["frequency"] = nil
					}
					modified = true
				}

				// Convert alerts_filter from list to object
				if filter, ok := action["alerts_filter"].([]any); ok {
					if len(filter) > 0 {
						filterObj := filter[0]
						action["alerts_filter"] = filterObj

						// Also convert timeframe within alerts_filter
						if filterMap, ok := filterObj.(map[string]any); ok {
							if tf, ok := filterMap["timeframe"].([]any); ok {
								if len(tf) > 0 {
									filterMap["timeframe"] = tf[0]
								} else {
									filterMap["timeframe"] = nil
								}
							}
						}
					} else {
						action["alerts_filter"] = nil
					}
					modified = true
				}
			}
		}
	}

	// Only re-marshal if we made changes
	if !modified {
		return
	}

	stateJSON, err := json.Marshal(stateMap)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal upgraded state", err.Error())
		return
	}

	resp.DynamicValue.JSON = stateJSON
}
