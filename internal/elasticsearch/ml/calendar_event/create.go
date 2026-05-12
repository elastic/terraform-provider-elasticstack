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
	"time"

	"github.com/elastic/go-elasticsearch/v8/typedapi/ml/postcalendarevents"
	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func calendarEventDateTimeToUnixMilli(dt estypes.DateTime) (int64, bool) {
	switch v := any(dt).(type) {
	case float64:
		return int64(v), true
	case int64:
		return v, true
	case int:
		return int64(v), true
	case string:
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return 0, false
		}
		return t.UnixMilli(), true
	default:
		return 0, false
	}
}

func calendarEventMatchesPlan(e estypes.CalendarEvent, description string, startMs, endMs int64) bool {
	if e.Description != description {
		return false
	}
	sm, ok1 := calendarEventDateTimeToUnixMilli(e.StartTime)
	em, ok2 := calendarEventDateTimeToUnixMilli(e.EndTime)
	return ok1 && ok2 && sm == startMs && em == endMs
}

func createCalendarEvent(ctx context.Context, client *clients.ElasticsearchScopedClient, plan CalendarEventTFModel) (CalendarEventTFModel, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	calendarID := plan.CalendarID.ValueString()

	apiModel, convDiags := plan.toAPIModel(ctx)
	diags.Append(convDiags...)
	if diags.HasError() {
		return plan, diags
	}

	startTime, d := plan.StartTime.ValueRFC3339Time()
	diags.Append(d...)
	if diags.HasError() {
		return plan, diags
	}
	endTime, d := plan.EndTime.ValueRFC3339Time()
	diags.Append(d...)
	if diags.HasError() {
		return plan, diags
	}
	planDesc := plan.Description.ValueString()
	planStartMs := startTime.UnixMilli()
	planEndMs := endTime.UnixMilli()

	tflog.Debug(ctx, fmt.Sprintf("Creating ML calendar event for calendar: %s", calendarID))

	typedClient, err := client.GetESClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return plan, diags
	}

	existingIDs := make(map[string]struct{})
	diags.Append(walkMLCalendarEventPages(ctx, typedClient, calendarID, func(events []estypes.CalendarEvent) bool {
		for _, e := range events {
			if e.EventId != nil && *e.EventId != "" {
				existingIDs[*e.EventId] = struct{}{}
			}
		}
		return false
	})...)
	if diags.HasError() {
		return plan, diags
	}

	eventPayload := estypes.CalendarEvent{
		Description: apiModel.Description,
		StartTime:   apiModel.StartTime,
		EndTime:     apiModel.EndTime,
	}

	postReq := postcalendarevents.NewRequest()
	postReq.Events = []estypes.CalendarEvent{eventPayload}

	postRes, err := typedClient.Ml.PostCalendarEvents(calendarID).Request(postReq).Do(ctx)
	if err != nil {
		diags.AddError("Failed to create ML calendar event", fmt.Sprintf("Unable to create ML calendar event for calendar %s — %s", calendarID, err.Error()))
		return plan, diags
	}

	var eventID string
	if postRes != nil {
		for _, e := range postRes.Events {
			if e.EventId != nil && *e.EventId != "" {
				eventID = *e.EventId
				break
			}
		}
	}

	if eventID == "" {
		var candidates []estypes.CalendarEvent
		diags.Append(walkMLCalendarEventPages(ctx, typedClient, calendarID, func(events []estypes.CalendarEvent) bool {
			for _, event := range events {
				if event.EventId == nil || *event.EventId == "" {
					continue
				}
				if _, existed := existingIDs[*event.EventId]; !existed {
					candidates = append(candidates, event)
				}
			}
			return false
		})...)
		if diags.HasError() {
			return plan, diags
		}

		switch len(candidates) {
		case 0:
			break
		case 1:
			eventID = *candidates[0].EventId
		default:
			for i := range candidates {
				if calendarEventMatchesPlan(candidates[i], planDesc, planStartMs, planEndMs) {
					eventID = *candidates[i].EventId
					break
				}
			}
		}
	}

	if eventID == "" {
		diags.AddError("Failed to identify created event", "Could not determine the new event ID from the API response or calendar events list")
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
