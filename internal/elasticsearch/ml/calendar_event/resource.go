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

package calendar_event

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                   = (*calendarEventResource)(nil)
	_ resource.ResourceWithConfigure      = (*calendarEventResource)(nil)
	_ resource.ResourceWithImportState    = (*calendarEventResource)(nil)
	_ resource.ResourceWithValidateConfig = (*calendarEventResource)(nil)
	_ resource.ResourceWithModifyPlan     = (*calendarEventResource)(nil)
)

type calendarEventResource struct {
	*entitycore.ElasticsearchResource[CalendarEventTFModel]
}

// newCalendarEventResource wires the envelope with placeholder create callbacks and a no-op update:
// Create is implemented on [*calendarEventResource] because post-create event discovery does not
// fit the generic envelope write path. Do not replace phCreate with a real createFunc without
// migrating that logic.
func newCalendarEventResource() *calendarEventResource {
	placeholder := entitycore.PlaceholderElasticsearchWriteCallback[CalendarEventTFModel]()
	return &calendarEventResource{
		ElasticsearchResource: entitycore.NewElasticsearchResource[CalendarEventTFModel]("ml_calendar_event", entitycore.ElasticsearchResourceOptions[CalendarEventTFModel]{
			Schema: getSchema,
			Read:   readCalendarEvent,
			Delete: deleteCalendarEvent,
			Create: placeholder,
			Update: updateCalendarEventNoOp,
		}),
	}
}

func NewCalendarEventResource() resource.Resource {
	return newCalendarEventResource()
}

func (r *calendarEventResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.Client() == nil {
		resp.Diagnostics.AddError("Client not configured", "Provider client is not configured")
		return
	}

	var plan CalendarEventTFModel
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

	written, callDiags := createCalendarEvent(ctx, client, plan)
	resp.Diagnostics.Append(callDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourcePath := written.CalendarID.ValueString() + "/" + written.EventID.ValueString()
	readModel, found, readDiags := readCalendarEvent(ctx, client, resourcePath, written)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError(
			"Failed to read created event",
			"Calendar event was not found after creation",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &readModel)...)
}

func (r *calendarEventResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The Elasticsearch envelope Read parses `id` as `<cluster_uuid>/<calendar_id>/<event_id>` (first slash
	// separates cluster from resource path; the remainder is split again for calendar vs event).
	// Separate `calendar_id` / `event_id` attributes are populated on refresh.
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
