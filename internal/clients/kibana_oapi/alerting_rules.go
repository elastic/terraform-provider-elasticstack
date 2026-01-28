package kibana_oapi

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// CreateAlertingRule creates a new alerting rule with a specific rule ID.
func CreateAlertingRule(ctx context.Context, client *Client, spaceID string, ruleID string, req kbapi.PostAlertingRuleIdJSONRequestBody) (*kbapi.PostAlertingRuleIdResponse, diag.Diagnostics) {
	resp, err := client.API.PostAlertingRuleIdWithResponse(ctx, ruleID, req, SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp, nil
	case http.StatusConflict:
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Rule ID Already Exists",
				fmt.Sprintf("Rule ID %s already exists in space %s", ruleID, spaceID),
			),
		}
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// GetAlertingRule retrieves a specific alerting rule by ID.
func GetAlertingRule(ctx context.Context, client *Client, spaceID string, ruleID string) (*kbapi.GetAlertingRuleIdResponse, diag.Diagnostics) {
	resp, err := client.API.GetAlertingRuleIdWithResponse(ctx, ruleID, SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateAlertingRule updates an existing alerting rule.
func UpdateAlertingRule(ctx context.Context, client *Client, spaceID string, ruleID string, req kbapi.PutAlertingRuleIdJSONRequestBody) (*kbapi.PutAlertingRuleIdResponse, diag.Diagnostics) {
	resp, err := client.API.PutAlertingRuleIdWithResponse(ctx, ruleID, req, SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// DeleteAlertingRule deletes an alerting rule.
func DeleteAlertingRule(ctx context.Context, client *Client, spaceID string, ruleID string) diag.Diagnostics {
	resp, err := client.API.DeleteAlertingRuleId(ctx, ruleID, SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNoContent, http.StatusNotFound:
		return nil
	default:
		body, _ := io.ReadAll(resp.Body)
		return reportUnknownError(resp.StatusCode, body)
	}
}

// EnableAlertingRule enables an alerting rule.
func EnableAlertingRule(ctx context.Context, client *Client, spaceID string, ruleID string) diag.Diagnostics {
	resp, err := client.API.PostAlertingRuleIdEnable(ctx, ruleID, SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNoContent:
		return nil
	default:
		body, _ := io.ReadAll(resp.Body)
		return reportUnknownError(resp.StatusCode, body)
	}
}

// DisableAlertingRule disables an alerting rule.
func DisableAlertingRule(ctx context.Context, client *Client, spaceID string, ruleID string) diag.Diagnostics {
	resp, err := client.API.PostAlertingRuleIdDisable(ctx, ruleID, kbapi.PostAlertingRuleIdDisableJSONRequestBody{}, SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	if resp.StatusCode() != http.StatusOK {
		return reportUnknownError(resp.StatusCode(), resp.Body)
	}

	return nil
}
