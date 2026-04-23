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
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/pfresource"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = &WorkflowResource{}
	_ resource.ResourceWithConfigure   = &WorkflowResource{}
	_ resource.ResourceWithImportState = &WorkflowResource{}
)

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &WorkflowResource{}
}

// WorkflowResource manages Kibana Agent Builder workflows.
type WorkflowResource struct {
	orchestrator pfresource.Orchestrator[kbapi.PostWorkflowsWorkflowJSONRequestBody, kbapi.PutWorkflowsWorkflowIdJSONRequestBody, *models.Workflow, *workflowModel]
}

// Configure sets up the resource with the provider client factory.
func (r *WorkflowResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	factory := pfresource.Configure(ctx, req.ProviderData, resp)
	if factory == nil {
		return
	}

	assembly := workflowAssembly{}
	r.orchestrator = pfresource.Orchestrator[kbapi.PostWorkflowsWorkflowJSONRequestBody, kbapi.PutWorkflowsWorkflowIdJSONRequestBody, *models.Workflow, *workflowModel]{
		Factory:  factory,
		Assembly: assembly,
	}
}

// Metadata returns the resource type name.
func (r *WorkflowResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	pfresource.Metadata(req, resp, "kibana_agentbuilder_workflow")
}

// ImportState imports the resource state.
func (r *WorkflowResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	assembly := workflowAssembly{}
	assembly.ImportState(ctx, req, resp)
}

// Create creates a new workflow resource.
func (r *WorkflowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan workflowModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	spaceID := plan.SpaceID.ValueString()
	if spaceID == "" {
		spaceID = defaultSpaceID
	}

	updated, diags := r.orchestrator.Create(ctx, &plan, spaceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, updated)...)
}

// Read reads the current state of the workflow resource.
func (r *WorkflowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state workflowModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	compID, diags := clients.CompositeIDFromStrFw(state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, present, diags := r.orchestrator.Read(ctx, &state, compID.ClusterID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !present {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, updated)...)
}

// Update updates an existing workflow resource.
func (r *WorkflowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan workflowModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	compID, diags := clients.CompositeIDFromStrFw(plan.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, diags := r.orchestrator.Update(ctx, &plan, compID.ClusterID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, updated)...)
}

// Delete deletes the workflow resource.
func (r *WorkflowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state workflowModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	compID, diags := clients.CompositeIDFromStrFw(state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.orchestrator.Delete(ctx, &state, compID.ClusterID)...)
}
