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

// This file contains Fleet package policy helpers for the Elastic Defend
// typed-input encoding. The generic GetPackagePolicy, CreatePackagePolicy, and
// UpdatePackagePolicy helpers in fleet.go use format=simplified which converts
// inputs to a map-keyed structure. Defend requires the raw (non-simplified)
// format so that typed inputs, their "type" discriminator, "config" payloads,
// and the top-level "version" token are preserved.

package fleet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetDefendPackagePolicy reads a specific Elastic Defend package policy from
// the Fleet API without requesting the simplified format. This preserves the
// typed input shape, input config payloads, and the top-level version token
// required for subsequent update operations.
func GetDefendPackagePolicy(ctx context.Context, client *Client, id string, spaceID string) (*kbapi.DefendPackagePolicy, diag.Diagnostics) {
	path := buildSpaceAwarePath(spaceID, fmt.Sprintf("/api/fleet/package_policies/%s", id))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, client.URL+path, nil)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var result struct {
			Item kbapi.DefendPackagePolicy `json:"item"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, diagutil.FrameworkDiagFromError(fmt.Errorf("failed to decode Defend package policy response: %w", err))
		}
		return &result.Item, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode, body)
	}
}

// CreateDefendPackagePolicy creates a new Elastic Defend package policy using
// the typed-input request body without requesting the simplified format. This
// is used for the Defend bootstrap create step.
func CreateDefendPackagePolicy(ctx context.Context, client *Client, spaceID string, req kbapi.DefendPackagePolicyRequest) (*kbapi.DefendPackagePolicy, diag.Diagnostics) {
	return sendDefendPackagePolicyRequest(ctx, client, http.MethodPost, spaceID, "", req)
}

// UpdateDefendPackagePolicy updates an existing Elastic Defend package policy
// using the typed-input request body without requesting the simplified format.
// The request body must include the top-level "version" token from the last
// successful read so Kibana can perform optimistic concurrency control.
func UpdateDefendPackagePolicy(ctx context.Context, client *Client, id string, spaceID string, req kbapi.DefendPackagePolicyRequest) (*kbapi.DefendPackagePolicy, diag.Diagnostics) {
	return sendDefendPackagePolicyRequest(ctx, client, http.MethodPut, spaceID, id, req)
}

// sendDefendPackagePolicyRequest is the shared transport for Defend create and
// update. It serialises the request, sends it with the appropriate method and
// path, and deserialises the typed response.
func sendDefendPackagePolicyRequest(ctx context.Context, client *Client, method, spaceID, id string, reqBody kbapi.DefendPackagePolicyRequest) (*kbapi.DefendPackagePolicy, diag.Diagnostics) {
	basePath := "/api/fleet/package_policies"
	if id != "" {
		basePath = fmt.Sprintf("%s/%s", basePath, id)
	}
	path := buildSpaceAwarePath(spaceID, basePath)

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(fmt.Errorf("failed to marshal Defend package policy request: %w", err))
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, client.URL+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.HTTP.Do(httpReq)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var result struct {
			Item kbapi.DefendPackagePolicy `json:"item"`
		}
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, diagutil.FrameworkDiagFromError(fmt.Errorf("failed to decode Defend package policy response: %w", err))
		}
		return &result.Item, nil
	default:
		return nil, reportUnknownError(resp.StatusCode, respBody)
	}
}
