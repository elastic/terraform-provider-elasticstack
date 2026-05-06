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
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// The Dashboard API currently requires allowUnmappedKeys for these requests.
func addDashboardRequestShapeEditor() func(ctx context.Context, req *http.Request) error {
	return func(_ context.Context, req *http.Request) error {
		query := req.URL.Query()
		query.Add("allowUnmappedKeys", "true")
		req.URL.RawQuery = query.Encode()
		return nil
	}
}

// GetDashboard reads a specific dashboard from the API.
func GetDashboard(ctx context.Context, client *Client, spaceID string, dashboardID string) (*kbapi.GetDashboardsIdResponse, diag.Diagnostics) {
	resp, err := client.API.GetDashboardsIdWithResponse(
		ctx, dashboardID,
		kibanautil.SpaceAwarePathRequestEditor(spaceID),
		addDashboardRequestShapeEditor(),
	)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// CreateDashboard creates a new dashboard.
func CreateDashboard(ctx context.Context, client *Client, spaceID string, req kbapi.PostDashboardsJSONRequestBody) (*kbapi.PostDashboardsResponse, diag.Diagnostics) {
	resp, err := client.API.PostDashboardsWithResponse(
		ctx, req,
		kibanautil.SpaceAwarePathRequestEditor(spaceID),
		addDashboardRequestShapeEditor(),
	)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusCreated:
		return resp, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// UpdateDashboard updates an existing dashboard.
func UpdateDashboard(ctx context.Context, client *Client, spaceID string, dashboardID string, req kbapi.PutDashboardsIdJSONRequestBody) (*kbapi.PutDashboardsIdResponse, diag.Diagnostics) {
	resp, err := client.API.PutDashboardsIdWithResponse(
		ctx, dashboardID, req,
		kibanautil.SpaceAwarePathRequestEditor(spaceID),
		addDashboardRequestShapeEditor(),
	)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// DeleteDashboard deletes an existing dashboard.
func DeleteDashboard(ctx context.Context, client *Client, spaceID string, dashboardID string) diag.Diagnostics {
	resp, err := client.API.DeleteDashboardsIdWithResponse(
		ctx, dashboardID,
		kibanautil.SpaceAwarePathRequestEditor(spaceID),
		addDashboardRequestShapeEditor(),
	)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return handleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK, http.StatusNoContent, http.StatusNotFound)
}
