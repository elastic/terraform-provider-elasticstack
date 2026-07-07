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
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v9/typedapi/ml/postcalendarevents"
	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/go-version"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// mlCalendarEventOptionalAPIFieldsMinElasticsearch is the minimum Elasticsearch version that accepts
// skip_result, skip_model_update, and force_time_shift on the post calendar events API.
var mlCalendarEventOptionalAPIFieldsMinElasticsearch = version.Must(version.NewVersion("8.16.0"))

func createCalendarEvent(ctx context.Context, client *clients.ElasticsearchScopedClient, plan CalendarEventTFModel) (CalendarEventTFModel, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	calendarID := plan.CalendarID.ValueString()

	planWire, wdiags := calendarEventWireFromTFModel(&plan)
	diags.Append(wdiags...)
	if diags.HasError() {
		return plan, diags
	}

	if postCalendarEventWireNeedsRawPOSTBody(planWire) {
		supported, vdiags := client.EnforceMinVersion(ctx, mlCalendarEventOptionalAPIFieldsMinElasticsearch)
		diags.Append(vdiags...)
		if diags.HasError() {
			return plan, diags
		}
		if !supported {
			diags.AddError(
				"ML calendar event optional scheduling fields not supported",
				fmt.Sprintf("skip_result, skip_model_update, and force_time_shift require Elasticsearch %s or newer "+
					"(or Elasticsearch Serverless). Omit these arguments or upgrade Elasticsearch.",
					mlCalendarEventOptionalAPIFieldsMinElasticsearch.String()),
			)
			return plan, diags
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("Creating ML calendar event for calendar: %s", calendarID))

	typedClient := client.GetESClient()

	var postWire []calendarEventWire
	var postBodyBytes []byte

	if postCalendarEventWireNeedsRawPOSTBody(planWire) {
		postBody := struct {
			Events []calendarEventWire `json:"events"`
		}{Events: []calendarEventWire{planWire}}
		var marshalErr error
		postBodyBytes, marshalErr = json.Marshal(postBody)
		if marshalErr != nil {
			diags.AddError("Failed to marshal ML calendar event request", marshalErr.Error())
			return plan, diags
		}

		postHTTP, err := typedClient.Ml.PostCalendarEvents(calendarID).Raw(bytes.NewReader(postBodyBytes)).Perform(ctx)
		if err != nil {
			diags.AddError("Failed to create ML calendar event", fmt.Sprintf("Unable to create ML calendar event for calendar %s — %s", calendarID, err.Error()))
			return plan, diags
		}
		defer postHTTP.Body.Close()
		postBodyBytes, err = io.ReadAll(postHTTP.Body)
		if err != nil {
			diags.AddError("Failed to read ML calendar event response", err.Error())
			return plan, diags
		}
		if postHTTP.StatusCode >= 300 {
			diags.AddError("Failed to create ML calendar event", fmt.Sprintf("Unable to create ML calendar event for calendar %s — status %d: %s", calendarID, postHTTP.StatusCode, string(postBodyBytes)))
			return plan, diags
		}

		var envelope struct {
			Events []calendarEventWire `json:"events"`
		}
		if err := json.Unmarshal(postBodyBytes, &envelope); err != nil {
			diags.AddError("Failed to decode ML calendar event response", fmt.Sprintf("Unable to decode post calendar events response for calendar %s — %s", calendarID, err.Error()))
			return plan, diags
		}
		postWire = envelope.Events
	} else {
		ev, convDiags := calendarEventTypesFromWireForPOST(planWire)
		diags.Append(convDiags...)
		if diags.HasError() {
			return plan, diags
		}

		req := postcalendarevents.NewRequest()
		req.Events = []estypes.CalendarEvent{ev}

		resp, err := typedClient.Ml.PostCalendarEvents(calendarID).Request(req).Do(ctx)
		if err != nil {
			if esErr, ok := errors.AsType[*estypes.ElasticsearchError](err); ok {
				diags.AddError("Failed to create ML calendar event", fmt.Sprintf("Unable to create ML calendar event for calendar %s — %s", calendarID, esErr.Error()))
				return plan, diags
			}
			diags.AddError("Failed to create ML calendar event", fmt.Sprintf("Unable to create ML calendar event for calendar %s — %s", calendarID, err.Error()))
			return plan, diags
		}

		postWire = make([]calendarEventWire, 0, len(resp.Events))
		for i := range resp.Events {
			w, convErr := calendarEventWireFromTypesCalendarEvent(resp.Events[i])
			if convErr != nil {
				diags.AddError("Failed to decode ML calendar event response", convErr.Error())
				return plan, diags
			}
			postWire = append(postWire, w)
		}
		var marshalErr error
		postBodyBytes, marshalErr = json.Marshal(postWire)
		if marshalErr != nil {
			postBodyBytes = nil
		}
	}

	var eventID string
	for i := range postWire {
		ev := postWire[i]
		if calendarEventMatchesPlanWire(ev, planWire) {
			if id := calendarEventWireEventID(&ev); id != "" {
				eventID = id
				break
			}
		}
	}
	if eventID == "" && len(postWire) == 1 {
		eventID = calendarEventWireEventID(&postWire[0])
	}

	if eventID == "" {
		windowStart, windowEnd, haveWindow := calendarEventWireWindowRFC3339(planWire)
		if !haveWindow {
			diags.AddError(
				"Failed to identify created event",
				"Could not derive start_time and end_time for listing calendar events after create.",
			)
			return plan, diags
		}

		var matches []calendarEventWire
		diags.Append(walkMLCalendarEventPagesWithWindow(ctx, typedClient, calendarID, windowStart, windowEnd, func(events []calendarEventWire) bool {
			for _, event := range events {
				id := calendarEventWireEventID(&event)
				if id == "" {
					continue
				}
				if calendarEventMatchesPlanWire(event, planWire) {
					matches = append(matches, event)
				}
			}
			return false
		})...)
		if diags.HasError() {
			return plan, diags
		}
		switch len(matches) {
		case 1:
			eventID = calendarEventWireEventID(&matches[0])
		case 0:
			break
		default:
			diags.AddError(
				"Failed to identify created event",
				fmt.Sprintf("Found %d calendar events matching this configuration after create; cannot determine which is the new event_id. "+
					"If duplicate events exist, remove extras or use a unique description.", len(matches)),
			)
			return plan, diags
		}
	}

	if eventID == "" {
		const maxPostRespDiag = 768
		var detail strings.Builder
		fmt.Fprintf(&detail, "Could not determine the new event ID from the API response or calendar events list. ")
		fmt.Fprintf(&detail, "post response had %d event(s). ", len(postWire))
		for i := range postWire {
			fmt.Fprintf(&detail, "[%d] event_id=%q description=%q; ", i, calendarEventWireEventID(&postWire[i]), postWire[i].Description)
		}
		fmt.Fprintf(&detail, "no calendar event in the cluster matched the planned description and time window after listing. ")
		if len(postBodyBytes) > 0 {
			ps := string(postBodyBytes)
			if len(ps) > maxPostRespDiag {
				ps = ps[:maxPostRespDiag] + "...(truncated)"
			}
			fmt.Fprintf(&detail, "post response excerpt: %s", ps)
		}
		diags.AddError("Failed to identify created event", detail.String())
		return plan, diags
	}

	compID, idDiags := client.ID(ctx, calendarID+"/"+eventID)
	diags.Append(idDiags...)
	if diags.HasError() {
		return plan, diags
	}

	plan.ID = types.StringValue(compID.String())
	plan.EventID = types.StringValue(eventID)
	plan.CalendarID = types.StringValue(calendarID)

	tflog.Debug(ctx, fmt.Sprintf("Successfully created ML calendar event %s for calendar: %s", eventID, calendarID))
	return plan, diags
}
