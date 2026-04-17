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
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetSlo retrieves a single SLO by space and ID. Returns (nil, nil) when
// the SLO is not found (HTTP 404), consistent with the resource layer's
// "not found" contract.
func GetSlo(ctx context.Context, client *Client, spaceID string, sloID string) (*kbapi.SLOsSloWithSummaryResponse, diag.Diagnostics) {
	resp, err := client.API.GetSloOpWithResponse(
		ctx,
		spaceID,
		sloID,
		&kbapi.GetSloOpParams{},
	)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Unable to get SLO", err.Error())}
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
				"Get SLO returned an empty response",
				"Get SLO returned an empty response body with HTTP status 200.",
			)}
		}
		return resp.JSON200, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// CreateSlo creates a new SLO in the given space and returns the created SLO's ID.
func CreateSlo(ctx context.Context, client *Client, spaceID string, req kbapi.SLOsCreateSloRequest) (*kbapi.SLOsCreateSloResponse, diag.Diagnostics) {
	resp, err := client.API.CreateSloOpWithResponse(
		ctx,
		spaceID,
		req,
	)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Unable to create SLO", err.Error())}
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
				"Create SLO returned an empty response",
				"Create SLO returned an empty response body with HTTP status 200.",
			)}
		}
		return resp.JSON200, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateSlo updates an existing SLO by space and ID.
func UpdateSlo(ctx context.Context, client *Client, spaceID string, sloID string, req kbapi.SLOsUpdateSloRequest) diag.Diagnostics {
	resp, err := client.API.UpdateSloOpWithResponse(
		ctx,
		spaceID,
		sloID,
		req,
	)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Unable to update SLO", err.Error())}
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return diag.Diagnostics{diag.NewErrorDiagnostic(
			"SLO not found during update",
			"The SLO with ID "+sloID+" was not found in space "+spaceID+".",
		)}
	default:
		return reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// DeleteSlo deletes an SLO by space and ID. A 404 response is treated as
// success (idempotent delete).
func DeleteSlo(ctx context.Context, client *Client, spaceID string, sloID string) diag.Diagnostics {
	resp, err := client.API.DeleteSloOpWithResponse(
		ctx,
		spaceID,
		sloID,
	)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Unable to delete SLO", err.Error())}
	}

	switch resp.StatusCode() {
	case http.StatusNoContent, http.StatusOK:
		return nil
	case http.StatusNotFound:
		return nil
	default:
		return reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// FindSlos performs a paginated search for SLOs in the given space. The
// optional params allow filtering by KQL query, pagination, and sorting.
func FindSlos(ctx context.Context, client *Client, spaceID string, params *kbapi.FindSlosOpParams) (*kbapi.SLOsFindSloResponse, diag.Diagnostics) {
	resp, err := client.API.FindSlosOpWithResponse(
		ctx,
		spaceID,
		params,
	)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Unable to find SLOs", err.Error())}
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
				"Find SLOs returned an empty response",
				"Find SLOs returned an empty response body with HTTP status 200.",
			)}
		}
		return resp.JSON200, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}
