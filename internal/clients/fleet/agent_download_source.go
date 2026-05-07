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
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetAgentDownloadSource reads a specific agent binary download source from the API.
func GetAgentDownloadSource(ctx context.Context, client *Client, id string, spaceID string) (*kbapi.GetFleetAgentDownloadSourcesSourceidResponse, diag.Diagnostics) {
	resp, err := client.API.GetFleetAgentDownloadSourcesSourceidWithResponse(ctx, id, kibanautil.SpaceAwarePathRequestEditor(spaceID))
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

// CreateAgentDownloadSource creates a new agent binary download source.
func CreateAgentDownloadSource(
	ctx context.Context,
	client *Client,
	spaceID string,
	req kbapi.PostFleetAgentDownloadSourcesJSONRequestBody,
) (*kbapi.PostFleetAgentDownloadSourcesResponse, diag.Diagnostics) {
	resp, err := client.API.PostFleetAgentDownloadSourcesWithResponse(ctx, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
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

// UpdateAgentDownloadSource updates an existing agent binary download source.
func UpdateAgentDownloadSource(
	ctx context.Context,
	client *Client,
	id string,
	spaceID string,
	req kbapi.PutFleetAgentDownloadSourcesSourceidJSONRequestBody,
) (*kbapi.PutFleetAgentDownloadSourcesSourceidResponse, diag.Diagnostics) {
	resp, err := client.API.PutFleetAgentDownloadSourcesSourceidWithResponse(ctx, id, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
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

// DeleteAgentDownloadSource deletes an existing agent binary download source.
func DeleteAgentDownloadSource(ctx context.Context, client *Client, id string, spaceID string) diag.Diagnostics {
	resp, err := client.API.DeleteFleetAgentDownloadSourcesSourceidWithResponse(ctx, id, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return handleDeleteResponse(resp.StatusCode(), resp.Body)
}

// ListAgentDownloadSources reads all agent binary download sources from the API.
func ListAgentDownloadSources(ctx context.Context, client *Client, spaceID string) (*kbapi.GetFleetAgentDownloadSourcesResponse, diag.Diagnostics) {
	resp, err := client.API.GetFleetAgentDownloadSourcesWithResponse(ctx, kibanautil.SpaceAwarePathRequestEditor(spaceID))
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
