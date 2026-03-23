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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// parseCompositeID extracts calendarID and eventID from the composite ID
// format: <cluster_uuid>/<calendar_id>/<event_id>
func parseCompositeID(id string) (calendarID, eventID string, diags fwdiags.Diagnostics) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 {
		diags.AddError("Invalid ID format", "Expected format: <cluster_uuid>/<calendar_id>/<event_id>")
		return
	}

	resourceParts := strings.SplitN(parts[1], "/", 2)
	if len(resourceParts) != 2 {
		diags.AddError("Invalid ID format", "Expected format: <cluster_uuid>/<calendar_id>/<event_id>")
		return
	}

	return resourceParts[0], resourceParts[1], diags
}

func (r *calendarEventResource) read(ctx context.Context, model *CalendarEventTFModel) (bool, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	if !r.resourceReady(&diags) {
		return false, diags
	}

	calendarID, eventID, parseDiags := parseCompositeID(model.ID.ValueString())
	diags.Append(parseDiags...)
	if diags.HasError() {
		return false, diags
	}

	tflog.Debug(ctx, fmt.Sprintf("Reading ML calendar event %s from calendar: %s", eventID, calendarID))

	esClient, err := r.client.GetESClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return false, diags
	}

	res, err := esClient.ML.GetCalendarEvents(
		calendarID,
		esClient.ML.GetCalendarEvents.WithContext(ctx),
		esClient.ML.GetCalendarEvents.WithSize(10000),
	)
	if err != nil {
		diags.AddError("Failed to get ML calendar events", err.Error())
		return false, diags
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return false, nil
	}

	getDiags := diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to get ML calendar events for calendar: %s", calendarID))
	diags.Append(getDiags...)
	if diags.HasError() {
		return false, diags
	}

	var response GetCalendarEventsResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		diags.AddError("Failed to decode calendar events response", err.Error())
		return false, diags
	}

	for _, event := range response.Events {
		if event.EventID == eventID {
			diags.Append(model.fromAPIModel(ctx, &event)...)
			if diags.HasError() {
				return false, diags
			}
			tflog.Debug(ctx, fmt.Sprintf("Successfully read ML calendar event %s from calendar: %s", eventID, calendarID))
			return true, diags
		}
	}

	return false, nil
}
