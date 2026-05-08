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
	"errors"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Update overrides the envelope callback because diffing job_ids requires comparing plan with state.
func (r *calendarResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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

	var state TFModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	calendarID := state.CalendarID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Updating ML calendar: %s", calendarID))

	client, connDiags := r.Client().GetElasticsearchClient(ctx, plan.GetElasticsearchConnection())
	resp.Diagnostics.Append(connDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	typedClient, err := client.GetESClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Elasticsearch client", err.Error())
		return
	}

	res, err := typedClient.Ml.GetCalendars().CalendarId(calendarID).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			resp.Diagnostics.AddError("Calendar not found", fmt.Sprintf("Calendar %s not found during update", calendarID))
			return
		}
		resp.Diagnostics.AddError("Failed to get current ML calendar", err.Error())
		return
	}

	if len(res.Calendars) == 0 {
		resp.Diagnostics.AddError("Calendar not found", fmt.Sprintf("Calendar %s not found during update", calendarID))
		return
	}

	currentCalendar := calendarTypedToAPIModel(&res.Calendars[0])

	currentJobIDSet := make(map[string]struct{})
	for _, id := range currentCalendar.JobIDs {
		currentJobIDSet[id] = struct{}{}
	}

	var planJobIDs []string
	if !plan.JobIDs.IsNull() && !plan.JobIDs.IsUnknown() {
		d := plan.JobIDs.ElementsAs(ctx, &planJobIDs, false)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	planJobIDSet := make(map[string]struct{})
	for _, id := range planJobIDs {
		planJobIDSet[id] = struct{}{}
	}

	for _, jobID := range planJobIDs {
		if _, exists := currentJobIDSet[jobID]; !exists {
			_, err := typedClient.Ml.PutCalendarJob(calendarID, jobID).Do(ctx)
			if err != nil {
				resp.Diagnostics.AddError("Failed to add job to calendar", fmt.Sprintf("Failed to add job %s to calendar %s: %s", jobID, calendarID, err.Error()))
				return
			}
		}
	}

	for _, jobID := range currentCalendar.JobIDs {
		if _, exists := planJobIDSet[jobID]; !exists {
			_, err := typedClient.Ml.DeleteCalendarJob(calendarID, jobID).Do(ctx)
			if err != nil {
				resp.Diagnostics.AddError("Failed to remove job from calendar", fmt.Sprintf("Failed to remove job %s from calendar %s: %s", jobID, calendarID, err.Error()))
				return
			}
		}
	}

	readModel, found, readDiags := readCalendar(ctx, client, calendarID, plan)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &readModel)...)

	tflog.Debug(ctx, fmt.Sprintf("Successfully updated ML calendar: %s", calendarID))
}
