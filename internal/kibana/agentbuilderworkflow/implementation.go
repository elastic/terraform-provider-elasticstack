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

package agentbuilderworkflow

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
	// Workflow API is GA from 9.4.x onwards
	minKibanaAgentBuilderAPIVersion = version.Must(version.NewVersion("9.4.0-SNAPSHOT"))
	defaultSpaceID                  = "default"
)

// workflowAssembly binds the schema, model, and API for the workflow resource.
type workflowAssembly struct{}

func (a workflowAssembly) TypeNameSuffix() string {
	return "kibana_agentbuilder_workflow"
}

func (a workflowAssembly) API() pfresource.ResourceAPI[kbapi.PostWorkflowsWorkflowJSONRequestBody, kbapi.PutWorkflowsWorkflowIdJSONRequestBody, *models.Workflow] {
	return &agentbuilderapi.WorkflowsAPI{}
}

func (a workflowAssembly) NewModel() *workflowModel {
	return &workflowModel{}
}

func (a workflowAssembly) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	pfresource.ImportStatePassthroughID(ctx, "id", req, resp)
}

// Ensure workflowModel implements the required interfaces.
var _ pfresource.KibanaConnectionModel = (*workflowModel)(nil)
var _ pfresource.IDModel = (*workflowModel)(nil)
var _ pfresource.SpaceIDModel = (*workflowModel)(nil)
var _ pfresource.ModelContract[kbapi.PostWorkflowsWorkflowJSONRequestBody, kbapi.PutWorkflowsWorkflowIdJSONRequestBody, *models.Workflow] = (*workflowModel)(nil)

// GetKibanaConnection returns the kibana_connection attribute.
func (m *workflowModel) GetKibanaConnection() types.List {
	return m.KibanaConnection
}

// GetID returns the id attribute.
func (m *workflowModel) GetID() types.String {
	return m.ID
}

// SetID sets the id attribute.
func (m *workflowModel) SetID(id types.String) {
	m.ID = id
}

// GetSpaceID returns the space_id attribute.
func (m *workflowModel) GetSpaceID() types.String {
	return m.SpaceID
}

// SetSpaceID sets the space_id attribute.
func (m *workflowModel) SetSpaceID(spaceID types.String) {
	m.SpaceID = spaceID
}

// VersionRequirement returns the minimum Kibana version requirement.
func (m *workflowModel) VersionRequirement() pfresource.VersionRequirement {
	return pfresource.VersionRequirement{
		MinimumVersion: minKibanaAgentBuilderAPIVersion,
		ErrorSummary:   "Unsupported server version",
		ErrorDetail:    "Agent Builder workflows require Elastic Stack v9.4.0 or later.",
	}
}

// ToCreateRequest converts the model to a create request.
func (m *workflowModel) ToCreateRequest(_ context.Context) (kbapi.PostWorkflowsWorkflowJSONRequestBody, diag.Diagnostics) {
	return m.toAPICreateModel(), nil
}

// ToUpdateRequest converts the model to an update request.
func (m *workflowModel) ToUpdateRequest(_ context.Context) (kbapi.PutWorkflowsWorkflowIdJSONRequestBody, diag.Diagnostics) {
	return m.toAPIUpdateModel(), nil
}

// PopulateFromRemote populates the model from the remote API response.
func (m *workflowModel) PopulateFromRemote(_ context.Context, spaceID string, remote *models.Workflow) diag.Diagnostics {
	// Restore space_id before populating so the model has the correct context
	m.SpaceID = types.StringValue(spaceID)
	m.populateFromAPI(remote)
	return nil
}
