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
	"github.com/elastic/terraform-provider-elasticstack/internal/asyncutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func readEntityLink(ctx context.Context, client *clients.KibanaScopedClient, resourceID string, spaceID string, prior entityLinkModel) (entityLinkModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	targetID := resourceID

	expectedEntityIDs := typeutils.SetTypeAs[string](ctx, prior.EntityIDs, path.Root("entity_ids"), &diags)
	if diags.HasError() {
		return prior, false, diags
	}

	// Only verify consistency (retry) during read-after-write when
	// ResolutionGroupJSON is still unknown. During normal refresh the state
	// is already populated; retrying would delay unnecessarily when IDs
	// were removed out-of-band.
	verifyConsistency := prior.ResolutionGroupJSON.IsUnknown() || prior.ResolutionGroupJSON.IsNull()

	body, payload, statusCode, d := readResolutionGroupWithRetry(ctx, client, targetID, spaceID, expectedEntityIDs, verifyConsistency)
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
	if typeutils.IsKnown(prior.TargetID) {
		result.TargetID = prior.TargetID
	} else {
		result.TargetID = types.StringValue(targetID)
	}
	diags.Append(result.populateFromAPI(ctx, spaceID, payload)...)
	return result, true, diags
}

// resolutionGroupPoll is the retry-loop unit of work: one call to the
// resolution-group endpoint plus its parsed payload.
type resolutionGroupPoll struct {
	body       []byte
	payload    map[string]any
	statusCode int
}

func readResolutionGroupWithRetry(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	targetID, spaceID string,
	expectedEntityIDs []string,
	verifyConsistency bool,
) ([]byte, map[string]any, int, diag.Diagnostics) {
	var diags diag.Diagnostics

	cfg := asyncutils.BackoffConfig{
		Initial:    100 * time.Millisecond,
		Max:        500 * time.Millisecond,
		MaxElapsed: 2 * time.Second,
	}

	result, err := asyncutils.PollWithBackoff(ctx, cfg, func(ctx context.Context, _ int) (resolutionGroupPoll, bool, error) {
		resp, err := client.GetKibanaOapiClient().API.GetSecurityEntityStoreResolutionGroupWithResponse(
			ctx,
			&kbapi.GetSecurityEntityStoreResolutionGroupParams{EntityId: targetID},
			kibanautil.SpaceAwarePathRequestEditor(spaceID),
		)
		if err != nil {
			diags.AddError("Failed to read resolution group", err.Error())
			return resolutionGroupPoll{}, true, err
		}

		statusCode := resp.StatusCode()
		body := resp.Body

		if statusCode != http.StatusOK {
			return resolutionGroupPoll{body: body, statusCode: statusCode}, true, nil
		}

		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			diags.AddError("Failed to parse resolution group response", err.Error())
			return resolutionGroupPoll{}, true, err
		}

		poll := resolutionGroupPoll{body: body, payload: payload, statusCode: statusCode}

		// No expected IDs to validate against, or consistency verification
		// is disabled for this read – accept the response immediately.
		if len(expectedEntityIDs) == 0 || !verifyConsistency {
			return poll, true, nil
		}

		apiEntityIDs := extractEntityIDsFromPayload(payload, targetID)
		return poll, containsAll(apiEntityIDs, expectedEntityIDs), nil
	})
	if err != nil {
		// The closure already reports its own errors via diags before
		// returning one; a bare error here means ctx was cancelled while
		// waiting between attempts.
		if !diags.HasError() {
			diags.AddError("Context cancelled during read-with-retry", err.Error())
		}
		return nil, nil, 0, diags
	}

	return result.body, result.payload, result.statusCode, diags
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
