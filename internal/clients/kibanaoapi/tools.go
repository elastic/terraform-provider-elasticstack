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
	"encoding/json"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetTool reads a specific tool from the API.
func GetTool(ctx context.Context, client *Client, spaceID string, toolID string) (*models.Tool, diag.Diagnostics) {
	resp, err := client.API.GetAgentBuilderToolsToolidWithResponse(ctx, toolID, SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		var tool models.Tool
		if err := json.Unmarshal(resp.Body, &tool); err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		return &tool, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// CreateTool creates a new tool.
func CreateTool(ctx context.Context, client *Client, spaceID string, req kbapi.PostAgentBuilderToolsJSONRequestBody) (*models.Tool, diag.Diagnostics) {
	resp, err := client.API.PostAgentBuilderToolsWithResponse(ctx, req, SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		var tool models.Tool
		if err := json.Unmarshal(resp.Body, &tool); err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		return &tool, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateTool updates an existing tool.
func UpdateTool(ctx context.Context, client *Client, spaceID string, toolID string, req kbapi.PutAgentBuilderToolsToolidJSONRequestBody) (*models.Tool, diag.Diagnostics) {
	resp, err := client.API.PutAgentBuilderToolsToolidWithResponse(ctx, toolID, req, SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		var tool models.Tool
		if err := json.Unmarshal(resp.Body, &tool); err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		return &tool, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// DeleteTool deletes an existing tool.
func DeleteTool(ctx context.Context, client *Client, spaceID string, toolID string) diag.Diagnostics {
	resp, err := client.API.DeleteAgentBuilderToolsToolidWithResponse(ctx, toolID, nil, SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return nil
	default:
		return reportUnknownError(resp.StatusCode(), resp.Body)
	}
}
