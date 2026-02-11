package kibana_oapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// EnableRulesByTag enables security detection rules that match a specific tag key-value pair.
func EnableRulesByTag(ctx context.Context, client *Client, spaceID, key, value string) diag.Diagnostics {
	query := fmt.Sprintf("alert.attributes.tags:(\"%s: %s\")", key, value)

	bulkAction := kbapi.SecurityDetectionsAPIBulkEnableRules{
		Action: kbapi.Enable,
		Query:  &query,
	}

	bodyBytes, err := json.Marshal(bulkAction)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Failed to marshal bulk action request", err.Error())}
	}

	tflog.Debug(ctx, "Enabling rules by tag", map[string]interface{}{
		"space_id":     spaceID,
		"key":          key,
		"value":        value,
		"query":        query,
		"request_body": string(bodyBytes),
	})

	resp, err := client.API.PerformRulesBulkActionWithBodyWithResponse(ctx, &kbapi.PerformRulesBulkActionParams{}, "application/json", bytes.NewReader(bodyBytes), SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Failed to enable rules by tag", err.Error())}
	}

	tflog.Debug(ctx, "Bulk action response", map[string]interface{}{
		"status_code":   resp.StatusCode(),
		"response_body": string(resp.Body),
	})

	if resp.StatusCode() != 200 {
		return diagutil.CheckHttpErrorFromFW(resp.HTTPResponse, "failed to enable rules by tag")
	}

	return nil
}

// DisableRulesByTag disables security detection rules that match a specific tag key-value pair.
func DisableRulesByTag(ctx context.Context, client *Client, spaceID, key, value string) diag.Diagnostics {
	query := fmt.Sprintf("alert.attributes.tags:(\"%s: %s\")", key, value)

	bulkAction := kbapi.SecurityDetectionsAPIBulkDisableRules{
		Action: kbapi.Disable,
		Query:  &query,
	}

	bodyBytes, err := json.Marshal(bulkAction)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Failed to marshal bulk action request", err.Error())}
	}

	resp, err := client.API.PerformRulesBulkActionWithBodyWithResponse(ctx, &kbapi.PerformRulesBulkActionParams{}, "application/json", bytes.NewReader(bodyBytes), SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Failed to disable rules by tag", err.Error())}
	}

	if resp.StatusCode() != 200 {
		return diagutil.CheckHttpErrorFromFW(resp.HTTPResponse, "failed to disable rules by tag")
	}

	return nil
}

// CheckRulesEnabledByTag checks if all rules matching a tag are enabled.
// Returns true if all matching rules are enabled, false if any are disabled.
func CheckRulesEnabledByTag(ctx context.Context, client *Client, spaceID, key, value string) (bool, diag.Diagnostics) {
	filter := fmt.Sprintf("alert.attributes.enabled: false AND alert.attributes.tags:(\"%s: %s\")", key, value)

	perPage := 1
	page := 1
	params := &kbapi.FindRulesParams{
		Filter:  &filter,
		Page:    &page,
		PerPage: &perPage,
	}

	resp, err := client.API.FindRulesWithResponse(ctx, params, SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return false, diag.Diagnostics{diag.NewErrorDiagnostic("Failed to query rules by tag", err.Error())}
	}

	if resp.StatusCode() != 200 {
		return false, diagutil.CheckHttpErrorFromFW(resp.HTTPResponse, "failed to query rules by tag")
	}

	if resp.JSON200 == nil {
		return false, diag.Diagnostics{diag.NewErrorDiagnostic("Empty response", "FindRules returned empty response")}
	}

	allEnabled := resp.JSON200.Total == 0

	return allEnabled, nil
}
