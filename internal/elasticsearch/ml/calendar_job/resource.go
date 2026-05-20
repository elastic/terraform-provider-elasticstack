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
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
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
	return &calendarJobResource{
		ElasticsearchResource: entitycore.NewElasticsearchResource[TFModel]("ml_calendar_job", entitycore.ElasticsearchResourceOptions[TFModel]{
			Schema: getSchema,
			Read:   readCalendarJob,
			Delete: deleteCalendarJob,
			Create: createCalendarJob,
			Update: updateCalendarJobNoOp,
		}),
	}
}

// NewCalendarJobResource returns the ML calendar–job assignment resource.
func NewCalendarJobResource() resource.Resource {
	return newCalendarJobResource()
}

func (r *calendarJobResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	compID, diags := clients.CompositeIDFromStr(req.ID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	calendarID, jobID, splitDiags := splitCalendarJobResourcePath(compID.ResourceID)
	resp.Diagnostics.Append(splitDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.Client() == nil {
		resp.Diagnostics.AddError("Client not configured", "Provider client is not configured")
		return
	}

	client, connDiags := r.Client().GetElasticsearchClient(ctx, providerschema.ElasticsearchConnectionNullList())
	resp.Diagnostics.Append(connDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	normalized, idDiags := client.ID(ctx, calendarID+"/"+jobID)
	resp.Diagnostics.Append(idDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	idStr := normalized.String()

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("elasticsearch_connection"), providerschema.ElasticsearchConnectionNullList())...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idStr)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("calendar_id"), calendarID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("job_id"), jobID)...)
}
