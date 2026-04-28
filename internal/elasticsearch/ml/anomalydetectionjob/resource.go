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

package anomalydetectionjob

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = newAnomalyDetectionJobResource()
	_ resource.ResourceWithConfigure   = newAnomalyDetectionJobResource()
	_ resource.ResourceWithImportState = newAnomalyDetectionJobResource()
)

type anomalyDetectionJobResource struct {
	*entitycore.ResourceBase
}

func newAnomalyDetectionJobResource() *anomalyDetectionJobResource {
	return &anomalyDetectionJobResource{
		ResourceBase: entitycore.NewResourceBase(entitycore.ComponentElasticsearch, "ml_anomaly_detection_job"),
	}
}

func NewAnomalyDetectionJobResource() resource.Resource {
	return newAnomalyDetectionJobResource()
}

func (r *anomalyDetectionJobResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.create(ctx, req, resp)
}

func (r *anomalyDetectionJobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TFModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, diags := r.read(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *anomalyDetectionJobResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.update(ctx, req, resp)
}

func (r *anomalyDetectionJobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.delete(ctx, req, resp)
}

// resourceReady checks if the client is ready for API calls
func (r *anomalyDetectionJobResource) resourceReady(diags *fwdiags.Diagnostics) bool {
	if r.Client() == nil {
		diags.AddError("Client not configured", "Provider client is not configured")
		return false
	}
	return true
}

func (r *anomalyDetectionJobResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import is intentionally sparse: only IDs are set. Everything else is populated by Read().
	compID, diags := clients.CompositeIDFromStrFw(req.ID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("job_id"), compID.ResourceID)...)
}
