// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package kibanaoapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func CreateAlertingRule(ctx context.Context, client *Client, spaceID string, rule models.AlertingRule) (*models.AlertingRule, diag.Diagnostics) {
	body, err := buildCreateRequestBody(rule)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Unable to build alerting rule create request", err.Error())}
	}

	var req kbapi.PostAlertingRuleIdJSONRequestBody
	err = req.FromAlertingRuleAPIBodyGeneric(body)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Unable to build alerting rule create request", err.Error())}
	}

	resp, err := client.API.PostAlertingRuleIdWithResponse(
		ctx,
		rule.RuleID,
		req,
		kibanautil.SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("HTTP request failed", err.Error())}
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		unwrapped, diags := diagutil.UnwrapJSON200(resp.JSON200, "alerting rule")
		if diags.HasError() {
			return nil, diags
		}
		return ConvertResponseToModel(spaceID, unwrapped)
	case http.StatusConflict:
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
			"Rule ID conflict",
			fmt.Sprintf("Status code [%d], Saved object [%s/%s] conflict (Rule ID already exists in this Space)", resp.StatusCode(), spaceID, rule.RuleID),
		)}
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

func GetAlertingRule(ctx context.Context, client *Client, spaceID string, ruleID string) (*models.AlertingRule, diag.Diagnostics) {
	resp, err := client.API.GetAlertingRuleIdWithResponse(
		ctx,
		ruleID,
		kibanautil.SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Unable to get alerting rule", err.Error())}
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		unwrapped, diags := diagutil.UnwrapJSON200(resp.JSON200, "alerting rule")
		if diags.HasError() {
			return nil, diags
		}
		return ConvertResponseToModel(spaceID, unwrapped)
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

func UpdateAlertingRule(ctx context.Context, client *Client, spaceID string, rule models.AlertingRule) (*models.AlertingRule, diag.Diagnostics) {
	body, err := buildUpdateRequestBody(rule)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Unable to build alerting rule update request", err.Error())}
	}

	resp, err := client.API.PutAlertingRuleIdWithResponse(
		ctx,
		rule.RuleID,
		body,
		kibanautil.SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Unable to update alerting rule", err.Error())}
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		unwrapped, diags := diagutil.UnwrapJSON200(resp.JSON200, "alerting rule")
		if diags.HasError() {
			return nil, diags
		}

		if diags := reconcileRuleEnabled(ctx, client, spaceID, rule, unwrapped); diags.HasError() {
			return nil, diags
		}

		returnedRule, convDiags := ConvertResponseToModel(spaceID, unwrapped)
		if convDiags.HasError() {
			return nil, convDiags
		}

		if rule.Enabled != nil {
			returnedRule.Enabled = rule.Enabled
		}

		return returnedRule, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

func DeleteAlertingRule(ctx context.Context, client *Client, spaceID string, ruleID string) diag.Diagnostics {
	resp, err := client.API.DeleteAlertingRuleIdWithResponse(
		ctx,
		ruleID,
		kibanautil.SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Unable to delete alerting rule", err.Error())}
	}

	switch resp.StatusCode() {
	case http.StatusNoContent, http.StatusOK:
		return nil
	case http.StatusNotFound:
		return nil
	default:
		return diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

func EnableAlertingRule(ctx context.Context, client *Client, spaceID string, ruleID string) diag.Diagnostics {
	resp, err := client.API.PostAlertingRuleIdEnableWithResponse(
		ctx,
		ruleID,
		kibanautil.SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Unable to enable alerting rule", err.Error())}
	}

	switch resp.StatusCode() {
	case http.StatusNoContent, http.StatusOK:
		return nil
	default:
		return diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

func DisableAlertingRule(ctx context.Context, client *Client, spaceID string, ruleID string) diag.Diagnostics {
	body := kbapi.PostAlertingRuleIdDisableJSONRequestBody{}
	resp, err := client.API.PostAlertingRuleIdDisableWithResponse(
		ctx,
		ruleID,
		body,
		kibanautil.SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Unable to disable alerting rule", err.Error())}
	}

	switch resp.StatusCode() {
	case http.StatusNoContent, http.StatusOK:
		return nil
	default:
		return diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

func reconcileRuleEnabled(ctx context.Context, client *Client, spaceID string, rule models.AlertingRule, unwrapped any) diag.Diagnostics {
	var wasEnabled bool
	if data, err := json.Marshal(unwrapped); err == nil {
		var temp struct {
			Enabled bool `json:"enabled"`
		}
		if err := json.Unmarshal(data, &temp); err == nil {
			wasEnabled = temp.Enabled
		}
	}

	shouldBeEnabled := rule.Enabled != nil && *rule.Enabled

	if shouldBeEnabled && !wasEnabled {
		return EnableAlertingRule(ctx, client, spaceID, rule.RuleID)
	}

	if !shouldBeEnabled && wasEnabled {
		return DisableAlertingRule(ctx, client, spaceID, rule.RuleID)
	}

	return nil
}
