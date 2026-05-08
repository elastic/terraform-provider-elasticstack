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

	"github.com/elastic/go-elasticsearch/v8/typedapi/ml/postcalendarevents"
	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func createCalendarEvent(ctx context.Context, client *clients.ElasticsearchScopedClient, plan CalendarEventTFModel) (CalendarEventTFModel, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	calendarID := plan.CalendarID.ValueString()

	apiModel, convDiags := plan.toAPIModel(ctx)
	diags.Append(convDiags...)
	if diags.HasError() {
		return plan, diags
	}

	tflog.Debug(ctx, fmt.Sprintf("Creating ML calendar event for calendar: %s", calendarID))

	typedClient, err := client.GetESClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return plan, diags
	}

	existingIDs := make(map[string]struct{})
	preRes, err := typedClient.Ml.GetCalendarEvents(calendarID).Size(10000).Do(ctx)
	if err == nil && preRes != nil {
		for _, e := range preRes.Events {
			if e.EventId != nil {
				existingIDs[*e.EventId] = struct{}{}
			}
		}
	}

	eventPayload := estypes.CalendarEvent{
		Description: apiModel.Description,
		StartTime:   apiModel.StartTime,
		EndTime:     apiModel.EndTime,
	}

	postReq := postcalendarevents.NewRequest()
	postReq.Events = []estypes.CalendarEvent{eventPayload}

	_, err = typedClient.Ml.PostCalendarEvents(calendarID).Request(postReq).Do(ctx)
	if err != nil {
		diags.AddError("Failed to create ML calendar event", fmt.Sprintf("Unable to create ML calendar event for calendar %s — %s", calendarID, err.Error()))
		return plan, diags
	}

	getRes, err := typedClient.Ml.GetCalendarEvents(calendarID).Size(10000).Do(ctx)
	if err != nil {
		diags.AddError("Failed to get calendar events after creation", err.Error())
		return plan, diags
	}

	var eventID string
	for _, event := range getRes.Events {
		if event.EventId == nil {
			continue
		}
		if _, existed := existingIDs[*event.EventId]; !existed {
			eventID = *event.EventId
			break
		}
	}

	if eventID == "" {
		diags.AddError("Failed to identify created event", "Could not find the newly created event in the calendar events list")
		return plan, diags
	}

	compID, sdkDiags := client.ID(ctx, calendarID+"/"+eventID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return plan, diags
	}

	plan.ID = types.StringValue(compID.String())
	plan.EventID = types.StringValue(eventID)
	plan.CalendarID = types.StringValue(calendarID)

	tflog.Debug(ctx, fmt.Sprintf("Successfully created ML calendar event %s for calendar: %s", eventID, calendarID))
	return plan, diags
}
