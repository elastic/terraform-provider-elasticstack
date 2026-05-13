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

	var found bool
	pageDiags := walkMLCalendarEventPages(ctx, typedClient, calendarID, func(events []calendarEventWire) bool {
		for _, event := range events {
			if calendarEventWireEventID(&event) != eventID {
				continue
			}
			apiModel, convDiags := wireEventToAPIModel(&event)
			diags.Append(convDiags...)
			if diags.HasError() {
				return true
			}
			diags.Append(state.fromAPIModel(ctx, apiModel)...)
			if diags.HasError() {
				return true
			}
			found = true
			tflog.Debug(ctx, fmt.Sprintf("Successfully read ML calendar event %s from calendar: %s", eventID, calendarID))
			return true
		}
		return false
	})
	diags.Append(pageDiags...)
	if diags.HasError() {
		return state, false, diags
	}
	if !found {
		return state, false, nil
	}
	return state, true, diags
}
