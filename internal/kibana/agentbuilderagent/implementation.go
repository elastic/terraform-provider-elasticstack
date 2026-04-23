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

package agentbuilderagent

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi/agentbuilderapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/pfresource"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	minKibanaAgentBuilderAPIVersion = version.Must(version.NewVersion("9.3.0"))
)

// agentAssembly binds the schema, model, and API for the agent resource.
type agentAssembly struct{}

func (a agentAssembly) TypeNameSuffix() string {
	return "kibana_agentbuilder_agent"
}

func (a agentAssembly) API() pfresource.ResourceAPI[kbapi.PostAgentBuilderAgentsJSONRequestBody, kbapi.PutAgentBuilderAgentsIdJSONRequestBody, *models.Agent] {
	return &agentbuilderapi.AgentsAPI{}
}

func (a agentAssembly) NewModel() *agentModel {
	return &agentModel{}
}

func (a agentAssembly) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	pfresource.ImportStateCompositeID(ctx, req, resp, "id", "space_id")
}

// Ensure agentModel implements the required interfaces.
var _ pfresource.KibanaConnectionModel = (*agentModel)(nil)
var _ pfresource.IDModel = (*agentModel)(nil)
var _ pfresource.SpaceIDModel = (*agentModel)(nil)
var _ pfresource.ModelContract[kbapi.PostAgentBuilderAgentsJSONRequestBody, kbapi.PutAgentBuilderAgentsIdJSONRequestBody, *models.Agent] = (*agentModel)(nil)

// GetKibanaConnection returns the kibana_connection attribute.
func (m *agentModel) GetKibanaConnection() types.List {
	return m.KibanaConnection
}

// GetID returns the id attribute.
func (m *agentModel) GetID() types.String {
	return m.ID
}

// SetID sets the id attribute.
func (m *agentModel) SetID(id types.String) {
	m.ID = id
}

// GetSpaceID returns the space_id attribute.
func (m *agentModel) GetSpaceID() types.String {
	return m.SpaceID
}

// SetSpaceID sets the space_id attribute.
func (m *agentModel) SetSpaceID(spaceID types.String) {
	m.SpaceID = spaceID
}

// VersionRequirement returns the minimum Kibana version requirement.
func (m *agentModel) VersionRequirement() pfresource.VersionRequirement {
	return pfresource.VersionRequirement{
		MinimumVersion: minKibanaAgentBuilderAPIVersion,
		ErrorSummary:   "Unsupported server version",
		ErrorDetail:    "Agent Builder agents require Elastic Stack v9.3.0 or later.",
	}
}

// ToCreateRequest converts the model to a create request.
func (m *agentModel) ToCreateRequest(ctx context.Context) (kbapi.PostAgentBuilderAgentsJSONRequestBody, diag.Diagnostics) {
	return m.toAPICreateModel(ctx)
}

// ToUpdateRequest converts the model to an update request.
func (m *agentModel) ToUpdateRequest(ctx context.Context) (kbapi.PutAgentBuilderAgentsIdJSONRequestBody, diag.Diagnostics) {
	return m.toAPIUpdateModel(ctx)
}

// PopulateFromRemote populates the model from the remote API response.
func (m *agentModel) PopulateFromRemote(ctx context.Context, spaceID string, remote *models.Agent) diag.Diagnostics {
	return m.populateFromAPI(ctx, spaceID, remote)
}
