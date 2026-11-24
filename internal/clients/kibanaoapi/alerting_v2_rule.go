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
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetAlertingV2Rule reads an Alerting v2 rule. A nil result without error
// diagnostics means the rule does not exist.
func GetAlertingV2Rule(ctx context.Context, client *Client, spaceID, ruleID string) (*kbapi.KibanaHTTPAPIsAlertingRuleResponse, diag.Diagnostics) {
	resp, err := client.API.GetAlertingV2RulesIdWithResponse(ctx, ruleID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.ErrDiag("Unable to get alerting v2 rule", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		rule, diags := diagutil.UnwrapJSON200(resp.JSON200, "alerting v2 rule")
		if diags.HasError() {
			return nil, diags
		}
		restoreArtifactDataPrecision(rule, resp.Body)
		return rule, diags
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, alertingV2ErrorDiagnostics(resp.StatusCode(), resp.Body)
	}
}

// restoreArtifactDataPrecision re-reads artifacts[].data from the raw response
// body as json.RawMessage and writes it back onto the parsed rule. The
// generated Data type is map[string]any, so the client has already
// decoded numbers as float64 and rounded any integer above 2^53; re-reading
// the raw bytes restores arbitrary JSON before it is marshalled back to state.
func restoreArtifactDataPrecision(rule *kbapi.KibanaHTTPAPIsAlertingRuleResponse, body []byte) {
	if rule == nil || rule.Artifacts == nil || len(body) == 0 {
		return
	}
	var raw struct {
		Artifacts []struct {
			Data map[string]json.RawMessage `json:"data"`
		} `json:"artifacts"`
	}
	if err := json.Unmarshal(body, &raw); err != nil || len(raw.Artifacts) != len(*rule.Artifacts) {
		return
	}
	for i := range *rule.Artifacts {
		precise := make(map[string]any, len(raw.Artifacts[i].Data))
		for key, value := range raw.Artifacts[i].Data {
			precise[key] = value
		}
		(*rule.Artifacts)[i].Data = precise
	}
}

// PutAlertingV2Rule creates or replaces an Alerting v2 rule and reports
// whether the server created it (HTTP 201) rather than replaced it (HTTP
// 200). The response body is never used as state, so only the outcome is
// returned; callers read the rule back afterwards.
func PutAlertingV2Rule(ctx context.Context, client *Client, spaceID, ruleID string, body kbapi.KibanaHTTPAPIsAlertingPutRule) (created bool, diags diag.Diagnostics) {
	resp, err := client.API.PutAlertingV2RulesIdWithResponse(ctx, ruleID, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return false, diagutil.ErrDiag("Unable to put alerting v2 rule", err)
	}

	switch resp.StatusCode() {
	case http.StatusCreated:
		_, diags = diagutil.UnwrapJSONStatus(resp.JSON201, "alerting v2 rule", http.StatusCreated)
		return true, diags
	case http.StatusOK:
		_, diags = diagutil.UnwrapJSONStatus(resp.JSON200, "alerting v2 rule", http.StatusOK)
		return false, diags
	default:
		return false, alertingV2ErrorDiagnostics(resp.StatusCode(), resp.Body)
	}
}

// DeleteAlertingV2Rule deletes an Alerting v2 rule. An already-absent rule
// (HTTP 404) is treated as success.
func DeleteAlertingV2Rule(ctx context.Context, client *Client, spaceID, ruleID string) diag.Diagnostics {
	resp, err := client.API.DeleteAlertingV2RulesIdWithResponse(ctx, ruleID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.ErrDiag("Unable to delete alerting v2 rule", err)
	}

	switch resp.StatusCode() {
	case http.StatusNoContent, http.StatusNotFound:
		return nil
	default:
		return alertingV2ErrorDiagnostics(resp.StatusCode(), resp.Body)
	}
}
