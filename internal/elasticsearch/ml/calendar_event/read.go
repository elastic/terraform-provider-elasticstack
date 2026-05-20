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
	"strings"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func splitCalendarEventResourcePath(resourcePath string) (calendarID, eventID string, diags fwdiags.Diagnostics) {
	parts := strings.SplitN(resourcePath, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		diags.AddError("Invalid ID format", "Expected resource segment format: <calendar_id>/<event_id>")
		return "", "", diags
	}
	return parts[0], parts[1], diags
}

func parseCalendarEventFullCompositeID(id string) (calendarID, eventID string, diags fwdiags.Diagnostics) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 {
		diags.AddError("Invalid ID format", "Expected format: <cluster_uuid>/<calendar_id>/<event_id>")
		return "", "", diags
	}
	return splitCalendarEventResourcePath(parts[1])
}

func calendarEventWireWindowRFC3339(w calendarEventWire) (start string, end string, ok bool) {
	startAny, err := rawJSONToAny(w.StartTime)
	if err != nil {
		return "", "", false
	}
	endAny, err := rawJSONToAny(w.EndTime)
	if err != nil {
		return "", "", false
	}
	startMillis, ok1 := calendarEventAnyTimeToUnixMilli(startAny)
	endMillis, ok2 := calendarEventAnyTimeToUnixMilli(endAny)
	if !ok1 || !ok2 {
		return "", "", false
	}
	return time.UnixMilli(startMillis).UTC().Format(time.RFC3339), time.UnixMilli(endMillis).UTC().Format(time.RFC3339), true
}

func calendarEventReadWindowRFC3339(state CalendarEventTFModel) (start string, end string, ok bool) {
	if state.StartTime.IsNull() || state.StartTime.IsUnknown() || state.EndTime.IsNull() || state.EndTime.IsUnknown() {
		return "", "", false
	}
	st, d := state.StartTime.ValueRFC3339Time()
	if d.HasError() {
		return "", "", false
	}
	et, d := state.EndTime.ValueRFC3339Time()
	if d.HasError() {
		return "", "", false
	}
	return st.UTC().Format(time.RFC3339), et.UTC().Format(time.RFC3339), true
}

func readCalendarEvent(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state CalendarEventTFModel) (CalendarEventTFModel, bool, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	calendarID, eventID, splitDiags := splitCalendarEventResourcePath(resourceID)
	diags.Append(splitDiags...)
	if diags.HasError() {
		return state, false, diags
	}

	tflog.Debug(ctx, fmt.Sprintf("Reading ML calendar event %s from calendar: %s", eventID, calendarID))

	typedClient, err := client.GetESClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return state, false, diags
	}

	tryWalk := func(startRFC3339, endRFC3339 string) (matched bool, walkDiags fwdiags.Diagnostics) {
		var inner fwdiags.Diagnostics
		walk := func(events []calendarEventWire) bool {
			for _, event := range events {
				if calendarEventWireEventID(&event) != eventID {
					continue
				}
				apiModel, convDiags := wireEventToAPIModel(&event)
				inner.Append(convDiags...)
				if inner.HasError() {
					return true
				}
				inner.Append(state.fromAPIModel(ctx, apiModel)...)
				if inner.HasError() {
					return true
				}
				matched = true
				tflog.Debug(ctx, fmt.Sprintf("Successfully read ML calendar event %s from calendar: %s", eventID, calendarID))
				return true
			}
			return false
		}
		if startRFC3339 != "" && endRFC3339 != "" {
			inner.Append(walkMLCalendarEventPagesWithWindow(ctx, typedClient, calendarID, startRFC3339, endRFC3339, walk)...)
		} else {
			inner.Append(walkMLCalendarEventPages(ctx, typedClient, calendarID, walk)...)
		}
		return matched, inner
	}

	windowStart, windowEnd, haveWindow := calendarEventReadWindowRFC3339(state)
	if haveWindow {
		found, wdiags := tryWalk(windowStart, windowEnd)
		diags.Append(wdiags...)
		if diags.HasError() {
			return state, false, diags
		}
		if found {
			return state, true, diags
		}
	}

	found, wdiags := tryWalk("", "")
	diags.Append(wdiags...)
	if diags.HasError() {
		return state, false, diags
	}
	if !found {
		return state, false, nil
	}
	return state, true, diags
}
