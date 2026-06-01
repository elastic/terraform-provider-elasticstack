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

package security_entity_store_entity_link

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/agentbuilder"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func readEntityLink(ctx context.Context, client *clients.KibanaScopedClient, resourceID string, spaceID string, prior entityLinkModel) (entityLinkModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	targetID := resourceID

	expectedEntityIDs, d := agentbuilder.SetToStrings(ctx, prior.EntityIDs)
	diags.Append(d...)
	if diags.HasError() {
		return prior, false, diags
	}

	body, statusCode, d := readResolutionGroupWithRetry(ctx, client, targetID, spaceID, expectedEntityIDs)
	diags.Append(d...)
	if diags.HasError() {
		return prior, false, diags
	}

	if statusCode == http.StatusNotFound {
		return prior, false, diags
	}
	if statusCode != http.StatusOK {
		diags.Append(diagutil.ReportUnknownHTTPError(statusCode, body)...)
		return prior, false, diags
	}

	result := prior
	result.TargetID = prior.TargetID
	if result.TargetID.IsNull() || result.TargetID.IsUnknown() {
		result.TargetID = types.StringValue(targetID)
	}
	diags.Append(result.populateFromAPI(ctx, spaceID, body, expectedEntityIDs)...)
	return result, true, diags
}

func readResolutionGroupWithRetry(ctx context.Context, client *clients.KibanaScopedClient, targetID, spaceID string, expectedEntityIDs []string) ([]byte, int, diag.Diagnostics) {
	var diags diag.Diagnostics
	backoff := 100 * time.Millisecond
	const maxDuration = 2 * time.Second
	start := time.Now()

	for {
		resp, err := client.GetKibanaOapiClient().API.GetSecurityEntityStoreResolutionGroupWithResponse(
			ctx,
			&kbapi.GetSecurityEntityStoreResolutionGroupParams{EntityId: targetID},
			kibanautil.SpaceAwarePathRequestEditor(spaceID),
		)
		if err != nil {
			diags.AddError("Failed to read resolution group", err.Error())
			return nil, 0, diags
		}

		statusCode := resp.StatusCode()
		body := resp.Body

		if statusCode == http.StatusNotFound {
			return body, statusCode, diags
		}
		if statusCode != http.StatusOK {
			return body, statusCode, diags
		}

		// No expected IDs to validate against – accept the response immediately.
		if len(expectedEntityIDs) == 0 {
			return body, statusCode, diags
		}

		apiEntityIDs := extractEntityIDsFromBody(body, targetID)
		if containsAll(apiEntityIDs, expectedEntityIDs) {
			return body, statusCode, diags
		}

		if time.Since(start) >= maxDuration {
			return body, statusCode, diags
		}

		time.Sleep(backoff)
		backoff *= 2
		if backoff > 500*time.Millisecond {
			backoff = 500 * time.Millisecond
		}
	}
}

func extractEntityIDsFromBody(body []byte, targetID string) []string {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}
	return extractEntityIDsFromPayload(payload, targetID)
}

func containsAll(haystack, needles []string) bool {
	set := make(map[string]struct{}, len(haystack))
	for _, h := range haystack {
		set[h] = struct{}{}
	}
	for _, n := range needles {
		if _, ok := set[n]; !ok {
			return false
		}
	}
	return true
}
