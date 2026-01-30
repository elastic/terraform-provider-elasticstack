package alerting_rule

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func (r *Resource) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			StateUpgrader: migrateV0ToV1,
		},
	}
}

// migrateV0ToV1 converts SDKv2 state format to Plugin Framework format.
// In SDKv2, nested blocks (frequency, alerts_filter, timeframe) were stored as JSON arrays.
// In Plugin Framework with SingleNestedBlock, they must be JSON objects.
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

	actions, ok := stateMap["actions"]
	if !ok || actions == nil {
		return
	}

	actionsList, ok := actions.([]any)
	if !ok {
		return
	}

	upgraded := false

	for i, action := range actionsList {
		actionMap, ok := action.(map[string]any)
		if !ok {
			continue
		}

		// Convert frequency from list to object
		if frequency, ok := actionMap["frequency"]; ok && frequency != nil {
			if frequencyList, ok := frequency.([]any); ok {
				upgraded = true
				if len(frequencyList) == 0 {
					actionMap["frequency"] = nil
				} else if first, ok := frequencyList[0].(map[string]any); ok {
					actionMap["frequency"] = first
				}
			}
		}

		// Convert alerts_filter from list to object
		if alertsFilter, ok := actionMap["alerts_filter"]; ok && alertsFilter != nil {
			if alertsFilterList, ok := alertsFilter.([]any); ok {
				upgraded = true
				if len(alertsFilterList) == 0 {
					actionMap["alerts_filter"] = nil
				} else if first, ok := alertsFilterList[0].(map[string]any); ok {
					// Also convert nested timeframe from list to object
					if timeframe, ok := first["timeframe"]; ok && timeframe != nil {
						if timeframeList, ok := timeframe.([]any); ok {
							if len(timeframeList) == 0 {
								first["timeframe"] = nil
							} else if timeframeFirst, ok := timeframeList[0].(map[string]any); ok {
								first["timeframe"] = timeframeFirst
							}
						}
					}
					actionMap["alerts_filter"] = first
				}
			}
		}

		actionsList[i] = actionMap
	}

	if !upgraded {
		return
	}

	stateMap["actions"] = actionsList

	stateJSON, err := json.Marshal(stateMap)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal raw state", err.Error())
		return
	}

	resp.DynamicValue.JSON = stateJSON
	resp.Diagnostics.AddAttributeWarning(
		path.Root("actions"),
		"Upgraded legacy actions state",
		"Upgraded legacy actions state format from SDKv2 list-based nested blocks to Plugin Framework object-based nested blocks.",
	)
}
