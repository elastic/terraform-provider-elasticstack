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
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

// ListSpaces returns all Kibana spaces.
func ListSpaces(ctx context.Context, client *Client) ([]kbapi.SpaceResponse, fwdiag.Diagnostics) {
	resp, err := client.API.GetSpacesSpaceWithResponse(ctx, &kbapi.GetSpacesSpaceParams{})
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		spaces, diags := diagutil.UnwrapJSON200(resp.JSON200, "spaces")
		if diags.HasError() {
			return nil, diags
		}
		return *spaces, nil
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

	return HandleGetTypedResponse(resp.StatusCode(), resp.Body,
		func() *kbapi.SpaceResponse { return resp.JSON200 })
}

// CreateSpace creates a new Kibana space.
func CreateSpace(ctx context.Context, client *Client, body kbapi.PostSpacesSpaceJSONRequestBody) (*kbapi.SpaceResponse, fwdiag.Diagnostics) {
	resp, err := client.API.PostSpacesSpaceWithResponse(ctx, body)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return HandleMutateTypedResponse(resp.StatusCode(), resp.Body,
			func() *kbapi.SpaceResponse { return resp.JSON200 })
	case http.StatusConflict:
		return nil, fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic(
			"Kibana space already exists",
			fmt.Sprintf("Space %q already exists. To manage an existing Kibana space with Terraform, import it first:\n\n    terraform import elasticstack_kibana_space.<NAME> %s", body.Id, body.Id),
		)}
	default:
		return HandleMutateTypedResponse(resp.StatusCode(), resp.Body,
			func() *kbapi.SpaceResponse { return resp.JSON200 })
	}
}

// UpdateSpace updates an existing Kibana space.
func UpdateSpace(ctx context.Context, client *Client, id string, body kbapi.PutSpacesSpaceIdJSONRequestBody) (*kbapi.SpaceResponse, fwdiag.Diagnostics) {
	resp, err := client.API.PutSpacesSpaceIdWithResponse(ctx, id, body)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	return HandleMutateTypedResponse(resp.StatusCode(), resp.Body,
		func() *kbapi.SpaceResponse { return resp.JSON200 })
}

// DeleteSpace deletes a Kibana space by ID.
func DeleteSpace(ctx context.Context, client *Client, id string) fwdiag.Diagnostics {
	resp, err := client.API.DeleteSpacesSpaceIdWithResponse(ctx, id)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusNoContent:
		return nil
	case http.StatusNotFound:
		return nil
	default:
		return diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}
