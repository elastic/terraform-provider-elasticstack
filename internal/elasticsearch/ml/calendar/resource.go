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

package calendar

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = (*calendarResource)(nil)
	_ resource.ResourceWithConfigure   = (*calendarResource)(nil)
	_ resource.ResourceWithImportState = (*calendarResource)(nil)
	_ resource.ResourceWithModifyPlan  = (*calendarResource)(nil)
)

type calendarResource struct {
	*entitycore.ElasticsearchResource[TFModel]
}

func newCalendarResource() *calendarResource {
	return &calendarResource{
		ElasticsearchResource: entitycore.NewElasticsearchResource[TFModel]("ml_calendar", entitycore.ElasticsearchResourceOptions[TFModel]{
			Schema: getSchema,
			Read:   readCalendar,
			Delete: deleteCalendar,
			Create: createCalendar,
			// Calendar definition changes (notably `description`) use RequiresReplace so
			// Terraform runs delete+create. ML put calendar is create-only on Elasticsearch
			// 8.0.x, so an in-place PUT would return "calendar already exists". Job
			// associations live on `elasticstack_elasticsearch_ml_calendar_job`.
			Update: entitycore.NoOpElasticsearchWriteCallback[TFModel](),
		}),
	}
}

func NewCalendarResource() resource.Resource {
	return newCalendarResource()
}

func (r *calendarResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	compID, diags := clients.CompositeIDFromStr(req.ID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("calendar_id"), compID.ResourceID)...)
}
