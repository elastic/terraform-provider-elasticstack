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
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// CreatePrivateLocation creates a new Synthetics private location via the OpenAPI client.
// On success it returns the SyntheticsGetPrivateLocation from the POST response body.
func CreatePrivateLocation(ctx context.Context, client *Client, spaceID string, body kbapi.PostPrivateLocationJSONRequestBody) (*kbapi.SyntheticsGetPrivateLocation, diag.Diagnostics) {
	resp, err := client.API.PostPrivateLocationWithResponse(ctx, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("HTTP request failed creating private location", err.Error())}
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
				"Create private location returned an empty response",
				fmt.Sprintf("Create private location returned an empty response with HTTP status code [%d].", resp.StatusCode()),
			)}
		}
		return resp.JSON200, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// GetPrivateLocation reads a Synthetics private location by id.
// Returns (nil, nil) when the location is not found (HTTP 404) so the caller
// can remove the resource from state without treating absence as an error.
func GetPrivateLocation(ctx context.Context, client *Client, spaceID string, locationID string) (*kbapi.SyntheticsGetPrivateLocation, diag.Diagnostics) {
	resp, err := client.API.GetPrivateLocationWithResponse(ctx, locationID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
			fmt.Sprintf("HTTP request failed reading private location %q", locationID),
			err.Error(),
		)}
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
				"Get private location returned an empty response",
				fmt.Sprintf("Get private location returned an empty response with HTTP status code [%d].", resp.StatusCode()),
			)}
		}
		return resp.JSON200, nil
	case http.StatusNotFound:
		// Sentinel: caller should remove from state.
		return nil, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// DeletePrivateLocation deletes a Synthetics private location by id.
func DeletePrivateLocation(ctx context.Context, client *Client, spaceID string, locationID string) diag.Diagnostics {
	resp, err := client.API.DeletePrivateLocationWithResponse(ctx, locationID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic(
			fmt.Sprintf("HTTP request failed deleting private location %q", locationID),
			err.Error(),
		)}
	}

	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK, http.StatusNotFound)
}
