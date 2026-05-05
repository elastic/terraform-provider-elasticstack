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
	"io"
	"net/http"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// CreateMonitor creates a new synthetics monitor via POST /api/synthetics/monitors.
func CreateMonitor(ctx context.Context, client *Client, spaceID string, req kbapi.SyntheticsMonitorRequest) (*kbapi.SyntheticsMonitor, diag.Diagnostics) {
	resp, err := client.API.PostSyntheticMonitorsWithResponse(
		ctx, req,
		SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, diagutil.FrameworkDiagFromError(fmt.Errorf("empty monitor response body"))
		}
		return resp.JSON200, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// GetMonitor reads a synthetics monitor by ID via GET /api/synthetics/monitors/{id}.
// Returns nil, nil when the monitor is not found (404).
func GetMonitor(ctx context.Context, client *Client, spaceID string, monitorID string) (*kbapi.SyntheticsMonitor, diag.Diagnostics) {
	resp, err := client.API.GetSyntheticMonitorWithResponse(
		ctx, monitorID,
		SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, diagutil.FrameworkDiagFromError(fmt.Errorf("empty monitor response body"))
		}
		return resp.JSON200, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// UpdateMonitor updates a synthetics monitor via PUT /api/synthetics/monitors/{id}.
func UpdateMonitor(ctx context.Context, client *Client, spaceID string, monitorID string, req kbapi.SyntheticsMonitorRequest) (*kbapi.SyntheticsMonitor, diag.Diagnostics) {
	resp, err := client.API.PutSyntheticMonitorWithResponse(
		ctx, monitorID, req,
		SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return GetMonitor(ctx, client, spaceID, monitorID)
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// monitorDeleteStatus represents a single item in the Kibana bulk-delete response body.
type monitorDeleteStatus struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

// DeleteMonitor deletes a synthetics monitor via DELETE /api/synthetics/monitors.
// It uses the bulk-delete endpoint (body: {"ids": [monitorID]}) which is supported
// across all Kibana versions that have the synthetics monitor resource (8.14+).
// The per-id endpoint (DELETE /api/synthetics/monitors/{id}) was not available in
// older Kibana versions and would return 404, causing silent deletion failures.
// The response body is parsed to detect per-item failures reported as HTTP 200 with
// deleted=false (e.g., when the monitor is in use or an internal error occurs).
func DeleteMonitor(ctx context.Context, client *Client, spaceID string, monitorID string) diag.Diagnostics {
	body, err := json.Marshal(map[string]any{"ids": []string{monitorID}})
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	path := BuildSpaceAwarePath(spaceID, "/api/synthetics/monitors")
	url := strings.TrimRight(client.URL, "/") + path

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, bytes.NewReader(body))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusNoContent, http.StatusNotFound:
		return nil
	case http.StatusOK:
		// Parse the per-item result array to detect partial failures.
		var results []monitorDeleteStatus
		if jsonErr := json.Unmarshal(respBody, &results); jsonErr != nil {
			// If the response body cannot be parsed, assume success — the HTTP 200
			// indicates the request was accepted by the server.
			return nil
		}
		for _, r := range results {
			if r.ID == monitorID && !r.Deleted {
				return diagutil.FrameworkDiagFromError(fmt.Errorf("monitor %s was not deleted by Kibana", monitorID))
			}
		}
		return nil
	default:
		return diagutil.FrameworkDiagFromError(fmt.Errorf("unexpected status %d deleting monitor %s: %s", resp.StatusCode, monitorID, string(respBody)))
	}
}
