package kibana_oapi

import (
	"context"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// CreateAlertingRule creates a new alerting rule.
func CreateAlertingRule(ctx context.Context, client *Client, ruleID string, req kbapi.AlertingRuleCreateRequest) (*kbapi.AlertingRuleResponse, diag.Diagnostics) {
	resp, err := client.API.PostAlertingRuleIdWithResponse(ctx, ruleID, req)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200, nil
	case http.StatusConflict:
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
			"Rule ID conflict",
			"A rule with the specified ID already exists",
		)}
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// GetAlertingRule retrieves an alerting rule by ID.
func GetAlertingRule(ctx context.Context, client *Client, ruleID string) (*kbapi.AlertingRuleResponse, diag.Diagnostics) {
	resp, err := client.API.GetAlertingRuleIdWithResponse(ctx, ruleID)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateAlertingRule updates an existing alerting rule.
func UpdateAlertingRule(ctx context.Context, client *Client, ruleID string, req kbapi.AlertingRuleUpdateRequest) (*kbapi.AlertingRuleResponse, diag.Diagnostics) {
	resp, err := client.API.PutAlertingRuleIdWithResponse(ctx, ruleID, req)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200, nil
	case http.StatusNotFound:
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
			"Rule not found",
			"The alerting rule with the specified ID does not exist",
		)}
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// DeleteAlertingRule deletes an alerting rule by ID.
func DeleteAlertingRule(ctx context.Context, client *Client, ruleID string) diag.Diagnostics {
	resp, err := client.API.DeleteAlertingRuleIdWithResponse(ctx, ruleID)
	if err != nil {
		return utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusNoContent, http.StatusOK:
		return nil
	case http.StatusNotFound:
		return nil // Already deleted
	default:
		return reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// EnableAlertingRule enables a disabled alerting rule.
func EnableAlertingRule(ctx context.Context, client *Client, ruleID string) diag.Diagnostics {
	resp, err := client.API.PostAlertingRuleIdEnableWithResponse(ctx, ruleID)
	if err != nil {
		return utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusNoContent, http.StatusOK:
		return nil
	case http.StatusNotFound:
		return diag.Diagnostics{diag.NewErrorDiagnostic(
			"Rule not found",
			"The alerting rule with the specified ID does not exist",
		)}
	default:
		return reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// DisableAlertingRule disables an alerting rule.
func DisableAlertingRule(ctx context.Context, client *Client, ruleID string) diag.Diagnostics {
	resp, err := client.API.PostAlertingRuleIdDisableWithResponse(ctx, ruleID, kbapi.PostAlertingRuleIdDisableJSONRequestBody{})
	if err != nil {
		return utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusNoContent, http.StatusOK:
		return nil
	case http.StatusNotFound:
		return diag.Diagnostics{diag.NewErrorDiagnostic(
			"Rule not found",
			"The alerting rule with the specified ID does not exist",
		)}
	default:
		return reportUnknownError(resp.StatusCode(), resp.Body)
	}
}
