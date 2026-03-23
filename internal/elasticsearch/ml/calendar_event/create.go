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
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *calendarEventResource) create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.resourceReady(&resp.Diagnostics) {
		return
	}

	var plan CalendarEventTFModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	calendarID := plan.CalendarID.ValueString()

	apiModel, diags := plan.toAPIModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Creating ML calendar event for calendar: %s", calendarID))

	esClient, err := r.client.GetESClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Elasticsearch client", err.Error())
		return
	}

	// Snapshot existing event IDs so we can identify the new one after creation.
	existingIDs := make(map[string]struct{})
	preRes, err := esClient.ML.GetCalendarEvents(
		calendarID,
		esClient.ML.GetCalendarEvents.WithContext(ctx),
		esClient.ML.GetCalendarEvents.WithSize(10000),
	)
	if err == nil {
		defer preRes.Body.Close()
		var preResp GetCalendarEventsResponse
		if json.NewDecoder(preRes.Body).Decode(&preResp) == nil {
			for _, e := range preResp.Events {
				existingIDs[e.EventID] = struct{}{}
			}
		}
	}

	reqBody := PostCalendarEventsRequest{Events: []CalendarEventAPIModel{*apiModel}}
	body, err := json.Marshal(reqBody)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal calendar event", err.Error())
		return
	}

	res, err := esClient.ML.PostCalendarEvents(calendarID, bytes.NewReader(body), esClient.ML.PostCalendarEvents.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create ML calendar event", err.Error())
		return
	}
	defer res.Body.Close()

	diags = diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to create ML calendar event for calendar: %s", calendarID))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The POST response does not include event_id. GET events to find the newly created one.
	getRes, err := esClient.ML.GetCalendarEvents(
		calendarID,
		esClient.ML.GetCalendarEvents.WithContext(ctx),
		esClient.ML.GetCalendarEvents.WithSize(10000),
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get calendar events after creation", err.Error())
		return
	}
	defer getRes.Body.Close()

	diags = diagutil.CheckErrorFromFW(getRes, fmt.Sprintf("Unable to get ML calendar events for calendar: %s", calendarID))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var getResp GetCalendarEventsResponse
	if err := json.NewDecoder(getRes.Body).Decode(&getResp); err != nil {
		resp.Diagnostics.AddError("Failed to decode calendar events response", err.Error())
		return
	}

	var eventID string
	for _, event := range getResp.Events {
		if _, existed := existingIDs[event.EventID]; !existed {
			eventID = event.EventID
			break
		}
	}

	if eventID == "" {
		resp.Diagnostics.AddError("Failed to identify created event", "Could not find the newly created event in the calendar events list")
		return
	}

	compID, sdkDiags := r.client.ID(ctx, calendarID+"/"+eventID)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(compID.String())
	plan.EventID = types.StringValue(eventID)
	plan.CalendarID = types.StringValue(calendarID)

	found, diags := r.read(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Failed to read created event", fmt.Sprintf("Calendar event with ID %s not found after creation", eventID))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	tflog.Debug(ctx, fmt.Sprintf("Successfully created ML calendar event %s for calendar: %s", eventID, calendarID))
}
