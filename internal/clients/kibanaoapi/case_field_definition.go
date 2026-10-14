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

// GetCaseFieldDefinition reads a field definition from the Kibana Cases field library.
// It returns (nil, nil) when the field definition does not exist.
func GetCaseFieldDefinition(ctx context.Context, client *Client, spaceID, fieldDefinitionID string) (*kbapi.CasesFieldDefinitionResponse, diag.Diagnostics) {
	resp, err := client.API.GetCaseFieldDefinitionDefaultSpaceWithResponse(ctx, fieldDefinitionID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return HandleGetRawResponse[kbapi.CasesFieldDefinitionResponse](resp.StatusCode(), resp.Body)
}

// CreateCaseFieldDefinition creates a field definition in the Kibana Cases field library.
func CreateCaseFieldDefinition(ctx context.Context, client *Client, spaceID string, req kbapi.CasesFieldDefinitionWriteRequest) (*kbapi.CasesFieldDefinitionResponse, diag.Diagnostics) {
	resp, err := client.API.CreateCaseFieldDefinitionDefaultSpaceWithResponse(ctx, &kbapi.CreateCaseFieldDefinitionDefaultSpaceParams{}, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return HandleMutateRawResponse[kbapi.CasesFieldDefinitionResponse](resp.StatusCode(), resp.Body)
}

// UpdateCaseFieldDefinition updates a field definition in the Kibana Cases field library.
func UpdateCaseFieldDefinition(ctx context.Context, client *Client, spaceID, fieldDefinitionID string, req kbapi.CasesFieldDefinitionPutRequest) (*kbapi.CasesFieldDefinitionResponse, diag.Diagnostics) {
	resp, err := client.API.UpdateCaseFieldDefinitionDefaultSpaceWithResponse(ctx, fieldDefinitionID, &kbapi.UpdateCaseFieldDefinitionDefaultSpaceParams{}, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return HandleMutateRawResponse[kbapi.CasesFieldDefinitionResponse](resp.StatusCode(), resp.Body)
}

// DeleteCaseFieldDefinition deletes a field definition from the Kibana Cases
// field library. HTTP 204 is success and 404 is treated as already deleted;
// any other status (including 409 when the field is still referenced) is an error.
func DeleteCaseFieldDefinition(ctx context.Context, client *Client, spaceID, fieldDefinitionID string) diag.Diagnostics {
	resp, err := client.API.DeleteCaseFieldDefinitionDefaultSpaceWithResponse(ctx, fieldDefinitionID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusNoContent, http.StatusNotFound)
}
