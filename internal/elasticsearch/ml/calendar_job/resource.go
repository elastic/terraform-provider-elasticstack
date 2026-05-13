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

package calendar_job

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = (*calendarJobResource)(nil)
	_ resource.ResourceWithConfigure   = (*calendarJobResource)(nil)
	_ resource.ResourceWithImportState = (*calendarJobResource)(nil)
)

type calendarJobResource struct {
	*entitycore.ElasticsearchResource[TFModel]
}

func newCalendarJobResource() *calendarJobResource {
	phCreate, phUpdate := entitycore.PlaceholderElasticsearchWriteCallbacks[TFModel]()
	return &calendarJobResource{
		ElasticsearchResource: entitycore.NewElasticsearchResource(
			entitycore.ComponentElasticsearch,
			"ml_calendar_job",
			getSchema,
			readCalendarJob,
			deleteCalendarJob,
			phCreate,
			phUpdate,
		),
	}
}

// NewCalendarJobResource returns the ML calendar–job assignment resource.
func NewCalendarJobResource() resource.Resource {
	return newCalendarJobResource()
}

func (r *calendarJobResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.Client() == nil {
		resp.Diagnostics.AddError("Client not configured", "Provider client is not configured")
		return
	}

	var plan TFModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, connDiags := r.Client().GetElasticsearchClient(ctx, plan.GetElasticsearchConnection())
	resp.Diagnostics.Append(connDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	written, callDiags := createCalendarJob(ctx, client, plan.GetResourceID().ValueString(), plan)
	resp.Diagnostics.Append(callDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	readModel, found, readDiags := readCalendarJob(ctx, client, plan.GetResourceID().ValueString(), written)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError(
			"Failed to read calendar job assignment",
			"Assignment was not found after create",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &readModel)...)
}

func (r *calendarJobResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"Calendar job assignments do not support in-place updates. Changing calendar_id or job_id requires replacement.",
	)
}

func (r *calendarJobResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	compID, diags := clients.CompositeIDFromStrFw(req.ID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	calendarID, jobID, splitDiags := splitCalendarJobResourcePath(compID.ResourceID)
	resp.Diagnostics.Append(splitDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("calendar_id"), calendarID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("job_id"), jobID)...)
}
