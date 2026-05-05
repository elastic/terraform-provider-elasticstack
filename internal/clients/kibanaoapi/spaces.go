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
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	sdkdiag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// ListSpaces returns all Kibana spaces.
func ListSpaces(ctx context.Context, client *Client) ([]kbapi.SpaceResponse, fwdiag.Diagnostics) {
	resp, err := client.API.GetSpacesSpaceWithResponse(ctx, &kbapi.GetSpacesSpaceParams{})
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, fwdiag.Diagnostics{
				fwdiag.NewErrorDiagnostic(
					"Unexpected empty response from Kibana Spaces API",
					"Got HTTP 200 but response body was empty or not JSON. This is likely a bug.",
				),
			}
		}
		return *resp.JSON200, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// GetSpace returns a single Kibana space by ID.
// Returns (nil, nil) when the space is not found (HTTP 404).
func GetSpace(ctx context.Context, client *Client, id string) (*kbapi.SpaceResponse, fwdiag.Diagnostics) {
	resp, err := client.API.GetSpacesSpaceIdWithResponse(ctx, id)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, fwdiag.Diagnostics{
				fwdiag.NewErrorDiagnostic(
					"Unexpected empty response from Kibana Spaces API",
					"Got HTTP 200 but response body was empty or not JSON. This is likely a bug.",
				),
			}
		}
		return resp.JSON200, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// GetSpaceSDK returns a single Kibana space by ID using SDK diagnostics.
// Returns (nil, nil) when the space is not found (HTTP 404).
func GetSpaceSDK(ctx context.Context, client *Client, id string) (*kbapi.SpaceResponse, sdkdiag.Diagnostics) {
	resp, err := client.API.GetSpacesSpaceIdWithResponse(ctx, id)
	if err != nil {
		return nil, sdkdiag.Diagnostics{
			{
				Severity: sdkdiag.Error,
				Summary:  "Error calling Kibana Spaces API",
				Detail:   err.Error(),
			},
		}
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, sdkdiag.Diagnostics{{
				Severity: sdkdiag.Error,
				Summary:  "Unexpected empty response from Kibana Spaces API",
				Detail:   "Got HTTP 200 but response body was empty or not JSON. This is likely a bug.",
			}}
		}
		return resp.JSON200, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, diagutil.ReportUnknownHTTPErrorSDK(resp.StatusCode(), resp.Body)
	}
}

// CreateSpace creates a new Kibana space.
func CreateSpace(ctx context.Context, client *Client, body kbapi.PostSpacesSpaceJSONRequestBody) (*kbapi.SpaceResponse, sdkdiag.Diagnostics) {
	resp, err := client.API.PostSpacesSpaceWithResponse(ctx, body)
	if err != nil {
		return nil, sdkdiag.Diagnostics{
			{
				Severity: sdkdiag.Error,
				Summary:  "Error calling Kibana Spaces API",
				Detail:   err.Error(),
			},
		}
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, sdkdiag.Diagnostics{{
				Severity: sdkdiag.Error,
				Summary:  "Unexpected empty response from Kibana Spaces API",
				Detail:   "Got HTTP 200 but response body was empty or not JSON. This is likely a bug.",
			}}
		}
		return resp.JSON200, nil
	default:
		return nil, diagutil.ReportUnknownHTTPErrorSDK(resp.StatusCode(), resp.Body)
	}
}

// UpdateSpace updates an existing Kibana space.
func UpdateSpace(ctx context.Context, client *Client, id string, body kbapi.PutSpacesSpaceIdJSONRequestBody) (*kbapi.SpaceResponse, sdkdiag.Diagnostics) {
	resp, err := client.API.PutSpacesSpaceIdWithResponse(ctx, id, body)
	if err != nil {
		return nil, sdkdiag.Diagnostics{
			{
				Severity: sdkdiag.Error,
				Summary:  "Error calling Kibana Spaces API",
				Detail:   err.Error(),
			},
		}
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, sdkdiag.Diagnostics{{
				Severity: sdkdiag.Error,
				Summary:  "Unexpected empty response from Kibana Spaces API",
				Detail:   "Got HTTP 200 but response body was empty or not JSON. This is likely a bug.",
			}}
		}
		return resp.JSON200, nil
	default:
		return nil, diagutil.ReportUnknownHTTPErrorSDK(resp.StatusCode(), resp.Body)
	}
}

// DeleteSpace deletes a Kibana space by ID.
func DeleteSpace(ctx context.Context, client *Client, id string) sdkdiag.Diagnostics {
	resp, err := client.API.DeleteSpacesSpaceIdWithResponse(ctx, id)
	if err != nil {
		return sdkdiag.Diagnostics{
			{
				Severity: sdkdiag.Error,
				Summary:  "Error calling Kibana Spaces API",
				Detail:   err.Error(),
			},
		}
	}

	switch resp.StatusCode() {
	case http.StatusNoContent:
		return nil
	case http.StatusNotFound:
		return nil
	default:
		return diagutil.ReportUnknownHTTPErrorSDK(resp.StatusCode(), resp.Body)
	}
}
