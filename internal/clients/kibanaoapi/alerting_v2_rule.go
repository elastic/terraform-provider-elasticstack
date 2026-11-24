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
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetAlertingV2Rule reads an Alerting v2 rule. A nil result without error
// diagnostics means the rule does not exist.
//
// It uses the raw request and decodes the body itself with json.Number instead
// of the generated typed helper. The generated Data field is
// map[string]interface{}, so the typed decoder turns every artifacts[].data
// number into a float64: it rounds integers above 2^53 and fails outright on a
// value outside float64 range (for example 1e400). Decoding with UseNumber
// keeps artifacts[].data as json.Number, so arbitrary JSON round-trips exactly.
func GetAlertingV2Rule(ctx context.Context, client *Client, spaceID, ruleID string) (*kbapi.KibanaHTTPAPIsAlertingRuleResponse, diag.Diagnostics) {
	resp, err := client.API.GetAlertingV2RulesId(ctx, ruleID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.ErrDiag("Unable to get alerting v2 rule", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, diagutil.ErrDiag("Unable to read alerting v2 rule response", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		rule, decodeDiags := decodeRulePreservingNumbers(body)
		if decodeDiags.HasError() {
			return nil, decodeDiags
		}
		return rule, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, alertingV2ErrorDiagnostics(resp.StatusCode, body)
	}
}

// decodeRulePreservingNumbers decodes a rule response with json.Number so that
// artifacts[].data keeps arbitrary-precision numbers instead of lossy float64.
func decodeRulePreservingNumbers(body []byte) (*kbapi.KibanaHTTPAPIsAlertingRuleResponse, diag.Diagnostics) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var rule kbapi.KibanaHTTPAPIsAlertingRuleResponse
	if err := decoder.Decode(&rule); err != nil {
		return nil, diagutil.ErrDiag("Failed to parse alerting v2 rule response", err)
	}
	return &rule, nil
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
