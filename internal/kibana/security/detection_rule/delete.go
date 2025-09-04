package detection_rule

import (
	"context"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/google/uuid"
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

	kibanaClient, err := r.client.GetKibanaOapiClient()
	if err != nil {
		diags.AddError("Unable to get Kibana client", err.Error())
		return
	}

	ruleID := stateModel.RuleID.ValueString()
	spaceID := stateModel.SpaceID.ValueString()
	if spaceID == "" {
		spaceID = "default"
	}

	tflog.Debug(ctx, "Deleting detection rule", map[string]interface{}{
		"id":       stateModel.ID.ValueString(),
		"rule_id":  ruleID,
		"space_id": spaceID,
	})

	// Prepare delete parameters - use rule_id if available, otherwise fall back to id
	params := &kbapi.DeleteRuleParams{}
	if ruleID != "" {
		ruleIDParam := kbapi.SecurityDetectionsAPIRuleSignatureId(ruleID)
		params.RuleId = &ruleIDParam
	} else {
		// Use the UUID id as fallback
		id := stateModel.ID.ValueString()
		if id != "" {
			// Parse the UUID from the string
			parsedUUID, err := uuid.Parse(id)
			if err != nil {
				diags.AddError("Invalid rule ID format", fmt.Sprintf("Failed to parse rule ID as UUID: %s", err))
				return
			}
			idParam := kbapi.SecurityDetectionsAPIRuleObjectId(parsedUUID)
			params.Id = &idParam
		} else {
			diags.AddError("Missing rule identifier", "Both rule_id and id are empty, cannot delete rule")
			return
		}
	}

	// Delete the rule
	resp, err := kibanaClient.API.DeleteRuleWithResponse(ctx, params,
		func(ctx context.Context, req *http.Request) error {
			// Add space ID header if not default space
			if spaceID != "default" {
				req.Header.Set("kbn-space-id", spaceID)
			}
			return nil
		},
	)
	if err != nil {
		diags.AddError("Error deleting detection rule", fmt.Sprintf("Request failed: %s", err))
		return
	}

	if resp.StatusCode() != http.StatusOK {
		body := "unknown error"
		if resp.Body != nil {
			body = string(resp.Body)
		}
		diags.AddError("Error deleting detection rule", fmt.Sprintf("API returned status %d: %s", resp.StatusCode(), body))
		return
	}

	tflog.Debug(ctx, "Successfully deleted detection rule", map[string]interface{}{
		"rule_id":  ruleID,
		"space_id": spaceID,
	})

	// State is automatically removed by the framework when delete succeeds
}
