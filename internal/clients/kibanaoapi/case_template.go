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

// GetCaseTemplate reads the latest version of a Kibana Cases template.
// It returns (nil, nil) when the template does not exist.
func GetCaseTemplate(ctx context.Context, client *Client, spaceID, templateID string) (*kbapi.CasesTemplateV2Response, diag.Diagnostics) {
	resp, err := client.API.GetCaseTemplateDefaultSpaceWithResponse(ctx, templateID, &kbapi.GetCaseTemplateDefaultSpaceParams{}, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return HandleGetRawResponse[kbapi.CasesTemplateV2Response](resp.StatusCode(), resp.Body)
}

// CreateCaseTemplate creates a Kibana Cases template. The dry_run parameter is
// never sent, so a successful response is always the template itself.
func CreateCaseTemplate(ctx context.Context, client *Client, spaceID string, req kbapi.CasesTemplateWriteRequest) (*kbapi.CasesTemplateV2Response, diag.Diagnostics) {
	resp, err := client.API.CreateCaseTemplateDefaultSpaceWithResponse(ctx, &kbapi.CreateCaseTemplateDefaultSpaceParams{}, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return HandleMutateRawResponse[kbapi.CasesTemplateV2Response](resp.StatusCode(), resp.Body)
}

// UpdateCaseTemplate fully replaces a Kibana Cases template, creating a new template version.
func UpdateCaseTemplate(ctx context.Context, client *Client, spaceID, templateID string, req kbapi.CasesTemplateWriteRequest) (*kbapi.CasesTemplateV2Response, diag.Diagnostics) {
	resp, err := client.API.UpdateCaseTemplateDefaultSpaceWithResponse(ctx, templateID, &kbapi.UpdateCaseTemplateDefaultSpaceParams{}, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return HandleMutateRawResponse[kbapi.CasesTemplateV2Response](resp.StatusCode(), resp.Body)
}

// DeleteCaseTemplate soft-deletes all versions of a Kibana Cases template.
// HTTP 204 is success and 404 is treated as already deleted.
func DeleteCaseTemplate(ctx context.Context, client *Client, spaceID, templateID string) diag.Diagnostics {
	resp, err := client.API.DeleteCaseTemplateDefaultSpaceWithResponse(ctx, templateID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusNoContent, http.StatusNotFound)
}
