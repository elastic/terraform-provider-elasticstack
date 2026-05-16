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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// performBulkRulesActionByTag marshals actionBody, calls the bulk-action API, and checks the response.
// verb is used in error messages (e.g. "enable", "disable").
func performBulkRulesActionByTag(ctx context.Context, client *Client, spaceID, key, value string, actionBody any, verb string) diag.Diagnostics {
	bodyBytes, err := json.Marshal(actionBody)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Failed to marshal bulk action request", err.Error())}
	}

	tflog.Debug(ctx, fmt.Sprintf("%sing rules by tag", verb), map[string]any{
		"space_id":     spaceID,
		"key":          key,
		"value":        value,
		"request_body": string(bodyBytes),
	})

	resp, err := client.API.PerformRulesBulkActionWithBodyWithResponse(
		ctx, &kbapi.PerformRulesBulkActionParams{}, "application/json",
		bytes.NewReader(bodyBytes), kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic(fmt.Sprintf("Failed to %s rules by tag", verb), err.Error())}
	}

	tflog.Debug(ctx, "Bulk action response", map[string]any{
		"status_code":   resp.StatusCode(),
		"response_body": string(resp.Body),
	})

	if resp.StatusCode() != 200 {
		return diagutil.CheckHTTPErrorFromFW(resp.HTTPResponse, fmt.Sprintf("failed to %s rules by tag", verb))
	}

	return nil
}

// EnableRulesByTag enables security detection rules that match a specific tag key-value pair.
func EnableRulesByTag(ctx context.Context, client *Client, spaceID, key, value string) diag.Diagnostics {
	query := fmt.Sprintf("alert.attributes.tags:(\"%s: %s\")", key, value)
	return performBulkRulesActionByTag(ctx, client, spaceID, key, value,
		kbapi.SecurityDetectionsAPIBulkEnableRules{Action: kbapi.Enable, Query: &query},
		"enable")
}

// DisableRulesByTag disables security detection rules that match a specific tag key-value pair.
func DisableRulesByTag(ctx context.Context, client *Client, spaceID, key, value string) diag.Diagnostics {
	query := fmt.Sprintf("alert.attributes.tags:(\"%s: %s\")", key, value)
	return performBulkRulesActionByTag(ctx, client, spaceID, key, value,
		kbapi.SecurityDetectionsAPIBulkDisableRules{Action: kbapi.Disable, Query: &query},
		"disable")
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

	resp, err := client.API.FindRulesWithResponse(ctx, params, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return false, diag.Diagnostics{diag.NewErrorDiagnostic("Failed to query rules by tag", err.Error())}
	}

	if resp.StatusCode() != 200 {
		return false, diagutil.CheckHTTPErrorFromFW(resp.HTTPResponse, "failed to query rules by tag")
	}

	unwrapped, unwrapDiags := diagutil.UnwrapJSON200(resp.JSON200, "find rules")
	if unwrapDiags.HasError() {
		return false, unwrapDiags
	}

	allEnabled := unwrapped.Total == 0

	return allEnabled, nil
}
