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
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *calendarEventResource) delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.resourceReady(&resp.Diagnostics) {
		return
	}

	var state CalendarEventTFModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	calendarID, eventID, parseDiags := parseCompositeID(state.ID.ValueString())
	resp.Diagnostics.Append(parseDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Deleting ML calendar event %s from calendar: %s", eventID, calendarID))

	esClient, err := r.client.GetESClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Elasticsearch client", err.Error())
		return
	}

	res, err := esClient.ML.DeleteCalendarEvent(calendarID, eventID, esClient.ML.DeleteCalendarEvent.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete ML calendar event", err.Error())
		return
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		tflog.Debug(ctx, fmt.Sprintf("ML calendar event %s already deleted from calendar: %s", eventID, calendarID))
		return
	}

	diags = diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to delete ML calendar event %s from calendar: %s", eventID, calendarID))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully deleted ML calendar event %s from calendar: %s", eventID, calendarID))
}
