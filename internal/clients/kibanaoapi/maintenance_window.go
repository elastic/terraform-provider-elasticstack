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
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetMaintenanceWindow reads a maintenance window from the API by ID
func GetMaintenanceWindow(ctx context.Context, client *Client, spaceID string, maintenanceWindowID string) (*kbapi.GetMaintenanceWindowIdResponse, diag.Diagnostics) {
	resp, err := client.API.GetMaintenanceWindowIdWithResponse(ctx, spaceID, maintenanceWindowID)

	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// CreateMaintenanceWindow creates a new maintenance window.
func CreateMaintenanceWindow(ctx context.Context, client *Client, spaceID string, body kbapi.PostMaintenanceWindowJSONRequestBody) (*kbapi.PostMaintenanceWindowResponse, diag.Diagnostics) {
	resp, err := client.API.PostMaintenanceWindowWithResponse(ctx, spaceID, body)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateMaintenanceWindow updates an existing maintenance window.
func UpdateMaintenanceWindow(ctx context.Context, client *Client, spaceID string, maintenanceWindowID string, req kbapi.PatchMaintenanceWindowIdJSONRequestBody) diag.Diagnostics {
	resp, err := client.API.PatchMaintenanceWindowIdWithResponse(ctx, spaceID, maintenanceWindowID, req)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return nil
	default:
		return reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// DeleteMaintenanceWindow deletes an existing maintenance window.
func DeleteMaintenanceWindow(ctx context.Context, client *Client, spaceID string, maintenanceWindowID string) diag.Diagnostics {
	resp, err := client.API.DeleteMaintenanceWindowIdWithResponse(ctx, spaceID, maintenanceWindowID)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusNoContent:
		return nil
	case http.StatusNotFound:
		return nil
	default:
		return reportUnknownError(resp.StatusCode(), resp.Body)
	}
}
