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
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetSkill reads a specific skill from the API.
func GetSkill(ctx context.Context, client *Client, spaceID, skillID string) (*models.Skill, diag.Diagnostics) {
	resp, err := client.API.GetAgentBuilderSkillsSkillidWithResponse(ctx, skillID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return handleGetResponse[models.Skill](resp.StatusCode(), resp.Body)
}

// CreateSkill creates a new skill.
func CreateSkill(ctx context.Context, client *Client, spaceID string, req kbapi.PostAgentBuilderSkillsJSONRequestBody) (*models.Skill, diag.Diagnostics) {
	resp, err := client.API.PostAgentBuilderSkillsWithResponse(ctx, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return handleMutateResponse[models.Skill](resp.StatusCode(), resp.Body)
}

// UpdateSkill updates an existing skill.
func UpdateSkill(ctx context.Context, client *Client, spaceID, skillID string, req kbapi.PutAgentBuilderSkillsSkillidJSONRequestBody) (*models.Skill, diag.Diagnostics) {
	resp, err := client.API.PutAgentBuilderSkillsSkillidWithResponse(ctx, skillID, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return handleMutateResponse[models.Skill](resp.StatusCode(), resp.Body)
}

// DeleteSkill deletes an existing skill. The API also accepts a `force=true`
// query parameter to cascade the deletion through referencing agents; the
// resource does not expose this in v1 so we always send an empty params
// struct and let 409 Conflict flow through as a normal error diagnostic.
func DeleteSkill(ctx context.Context, client *Client, spaceID, skillID string) diag.Diagnostics {
	resp, err := client.API.DeleteAgentBuilderSkillsSkillidWithResponse(ctx, skillID, &kbapi.DeleteAgentBuilderSkillsSkillidParams{}, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK, http.StatusNotFound)
}
