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
	"io"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func createCalendarEvent(ctx context.Context, client *clients.ElasticsearchScopedClient, plan CalendarEventTFModel) (CalendarEventTFModel, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	calendarID := plan.CalendarID.ValueString()

	planWire, wdiags := calendarEventWireFromTFModel(&plan)
	diags.Append(wdiags...)
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
	diags.Append(walkMLCalendarEventPages(ctx, typedClient, calendarID, func(events []calendarEventWire) bool {
		for _, e := range events {
			if id := calendarEventWireEventID(&e); id != "" {
				existingIDs[id] = struct{}{}
			}
		}
		return false
	})...)
	if diags.HasError() {
		return plan, diags
	}

	postBody := struct {
		Events []calendarEventWire `json:"events"`
	}{Events: []calendarEventWire{planWire}}
	body, err := json.Marshal(postBody)
	if err != nil {
		diags.AddError("Failed to marshal ML calendar event request", err.Error())
		return plan, diags
	}

	postHTTP, err := typedClient.Ml.PostCalendarEvents(calendarID).Raw(bytes.NewReader(body)).Perform(ctx)
	if err != nil {
		diags.AddError("Failed to create ML calendar event", fmt.Sprintf("Unable to create ML calendar event for calendar %s — %s", calendarID, err.Error()))
		return plan, diags
	}
	defer postHTTP.Body.Close()
	postBodyBytes, readErr := io.ReadAll(postHTTP.Body)
	if readErr != nil {
		diags.AddError("Failed to read ML calendar event response", readErr.Error())
		return plan, diags
	}
	if postHTTP.StatusCode >= 300 {
		diags.AddError("Failed to create ML calendar event", fmt.Sprintf("Unable to create ML calendar event for calendar %s — status %d: %s", calendarID, postHTTP.StatusCode, string(postBodyBytes)))
		return plan, diags
	}

	var eventID string
	var postWire struct {
		Events []calendarEventWire `json:"events"`
	}
	if err := json.Unmarshal(postBodyBytes, &postWire); err != nil {
		diags.AddError("Failed to decode ML calendar event response", fmt.Sprintf("Unable to decode post calendar events response for calendar %s — %s", calendarID, err.Error()))
		return plan, diags
	}
	for i := range postWire.Events {
		ev := postWire.Events[i]
		if calendarEventMatchesPlanWire(ev, planWire) {
			if id := calendarEventWireEventID(&ev); id != "" {
				eventID = id
				break
			}
		}
	}
	if eventID == "" && len(postWire.Events) == 1 {
		eventID = calendarEventWireEventID(&postWire.Events[0])
	}

	if eventID == "" {
		var candidates []calendarEventWire
		diags.Append(walkMLCalendarEventPages(ctx, typedClient, calendarID, func(events []calendarEventWire) bool {
			for _, event := range events {
				id := calendarEventWireEventID(&event)
				if id == "" {
					continue
				}
				if _, existed := existingIDs[id]; !existed {
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
			eventID = calendarEventWireEventID(&candidates[0])
		default:
			for i := range candidates {
				if calendarEventMatchesPlanWire(candidates[i], planWire) {
					eventID = calendarEventWireEventID(&candidates[i])
					break
				}
			}
		}

		if eventID == "" {
			const maxPostRespDiag = 768
			var detail strings.Builder
			fmt.Fprintf(&detail, "Could not determine the new event ID from the API response or calendar events list. ")
			fmt.Fprintf(&detail, "post response had %d event(s). ", len(postWire.Events))
			for i := range postWire.Events {
				fmt.Fprintf(&detail, "[%d] event_id=%q description=%q; ", i, calendarEventWireEventID(&postWire.Events[i]), postWire.Events[i].Description)
			}
			fmt.Fprintf(&detail, "new-event candidates from list: %d. ", len(candidates))
			ps := string(postBodyBytes)
			if len(ps) > maxPostRespDiag {
				ps = ps[:maxPostRespDiag] + "...(truncated)"
			}
			fmt.Fprintf(&detail, "post response excerpt: %s", ps)
			diags.AddError("Failed to identify created event", detail.String())
			return plan, diags
		}
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
