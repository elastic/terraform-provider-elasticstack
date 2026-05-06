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

package fleet

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetEnrollmentTokens reads all enrollment tokens from the API.
func GetEnrollmentTokens(ctx context.Context, client *Client, spaceID string) ([]kbapi.EnrollmentApiKey, diag.Diagnostics) {
	resp, err := client.API.GetFleetEnrollmentApiKeysWithResponse(ctx, nil, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, clientError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200.Items, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// GetEnrollmentTokensByPolicy Get enrollment tokens by given policy ID.
func GetEnrollmentTokensByPolicy(ctx context.Context, client *Client, policyID string) ([]kbapi.EnrollmentApiKey, diag.Diagnostics) {
	params := kbapi.GetFleetEnrollmentApiKeysParams{
		Kuery: new("policy_id:" + policyID),
	}

	resp, err := client.API.GetFleetEnrollmentApiKeysWithResponse(ctx, &params)
	if err != nil {
		return nil, clientError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200.Items, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// GetEnrollmentTokensByPolicyInSpace Get enrollment tokens by policy ID within a specific Kibana space.
func GetEnrollmentTokensByPolicyInSpace(ctx context.Context, client *Client, policyID string, spaceID string) ([]kbapi.EnrollmentApiKey, diag.Diagnostics) {
	path := kibanautil.BuildSpaceAwarePath(spaceID, "/api/fleet/enrollment_api_keys?kuery=policy_id:"+policyID)

	req, err := http.NewRequestWithContext(ctx, "GET", client.URL+path, nil)
	if err != nil {
		return nil, clientError(err)
	}

	httpResp, err := client.HTTP.Do(req)
	if err != nil {
		return nil, clientError(err)
	}
	defer httpResp.Body.Close()

	switch httpResp.StatusCode {
	case http.StatusOK:
		var result struct {
			Items []kbapi.EnrollmentApiKey `json:"items"`
		}
		if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
			return nil, clientError(err)
		}
		return result.Items, nil
	default:
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return nil, diagutil.ReportUnknownHTTPError(httpResp.StatusCode, bodyBytes)
	}
}
