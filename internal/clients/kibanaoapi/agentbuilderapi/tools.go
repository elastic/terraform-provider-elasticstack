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

package agentbuilderapi

import (
	"context"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ToolsAPI implements the ResourceAPI interface for Agent Builder tools.
type ToolsAPI struct{}

// Create creates a new tool and returns its ID.
func (t *ToolsAPI) Create(
	ctx context.Context,
	client *kibanaoapi.Client,
	spaceID string,
	req kbapi.PostAgentBuilderToolsJSONRequestBody,
) (string, diag.Diagnostics) {
	resp, err := client.API.PostAgentBuilderToolsWithResponse(ctx, req, kibanaoapi.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return "", diagutil.FrameworkDiagFromError(err)
	}

	tool, diags := handleMutateResponse[models.Tool](resp.StatusCode(), resp.Body)
	if diags.HasError() {
		return "", diags
	}

	return tool.ID, diags
}

// Get retrieves a tool by ID. Returns (nil, false, nil) if not found.
func (t *ToolsAPI) Get(
	ctx context.Context,
	client *kibanaoapi.Client,
	spaceID string,
	toolID string,
) (*models.Tool, bool, diag.Diagnostics) {
	resp, err := client.API.GetAgentBuilderToolsToolidWithResponse(ctx, toolID, kibanaoapi.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, false, diagutil.FrameworkDiagFromError(err)
	}

	return handleGetResponse[models.Tool](resp.StatusCode(), resp.Body)
}

// Update updates an existing tool.
func (t *ToolsAPI) Update(
	ctx context.Context,
	client *kibanaoapi.Client,
	spaceID string,
	toolID string,
	req kbapi.PutAgentBuilderToolsToolidJSONRequestBody,
) diag.Diagnostics {
	resp, err := client.API.PutAgentBuilderToolsToolidWithResponse(ctx, toolID, req, kibanaoapi.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return handleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK)
}

// Delete deletes a tool by ID.
func (t *ToolsAPI) Delete(
	ctx context.Context,
	client *kibanaoapi.Client,
	spaceID string,
	toolID string,
) diag.Diagnostics {
	resp, err := client.API.DeleteAgentBuilderToolsToolidWithResponse(ctx, toolID, nil, kibanaoapi.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return handleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK, http.StatusNotFound)
}
