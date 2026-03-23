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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *calendarResource) update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.resourceReady(&resp.Diagnostics) {
		return
	}

	var plan CalendarTFModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state CalendarTFModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	calendarID := state.CalendarID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Updating ML calendar: %s", calendarID))

	esClient, err := r.client.GetESClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Elasticsearch client", err.Error())
		return
	}

	compID, sdkDiags := clients.CompositeIdFromStr(state.ID.ValueString())
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := esClient.ML.GetCalendars(esClient.ML.GetCalendars.WithCalendarID(compID.ResourceId), esClient.ML.GetCalendars.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to get current ML calendar", err.Error())
		return
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		resp.Diagnostics.AddError("Calendar not found", fmt.Sprintf("Calendar %s not found during update", calendarID))
		return
	}

	getDiags := diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to get ML calendar for update: %s", calendarID))
	resp.Diagnostics.Append(getDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var currentResponse struct {
		Calendars []CalendarAPIModel `json:"calendars"`
		Count     int                `json:"count"`
	}
	if err := json.NewDecoder(res.Body).Decode(&currentResponse); err != nil {
		resp.Diagnostics.AddError("Failed to decode calendar response", err.Error())
		return
	}

	if len(currentResponse.Calendars) == 0 {
		resp.Diagnostics.AddError("Calendar not found", fmt.Sprintf("Calendar %s not found during update", calendarID))
		return
	}

	currentCalendar := currentResponse.Calendars[0]

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
			addRes, err := esClient.ML.PutCalendarJob(calendarID, jobID, esClient.ML.PutCalendarJob.WithContext(ctx))
			if err != nil {
				resp.Diagnostics.AddError("Failed to add job to calendar", fmt.Sprintf("Failed to add job %s to calendar %s: %s", jobID, calendarID, err.Error()))
				return
			}
			defer addRes.Body.Close()

			diags = diagutil.CheckErrorFromFW(addRes, fmt.Sprintf("Unable to add job %s to calendar %s", jobID, calendarID))
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}

	for _, jobID := range currentCalendar.JobIDs {
		if _, exists := planJobIDSet[jobID]; !exists {
			removeRes, err := esClient.ML.DeleteCalendarJob(calendarID, jobID, esClient.ML.DeleteCalendarJob.WithContext(ctx))
			if err != nil {
				resp.Diagnostics.AddError("Failed to remove job from calendar", fmt.Sprintf("Failed to remove job %s from calendar %s: %s", jobID, calendarID, err.Error()))
				return
			}
			defer removeRes.Body.Close()

			diags = diagutil.CheckErrorFromFW(removeRes, fmt.Sprintf("Unable to remove job %s from calendar %s", jobID, calendarID))
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}

	found, diags := r.read(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	tflog.Debug(ctx, fmt.Sprintf("Successfully updated ML calendar: %s", calendarID))
}
