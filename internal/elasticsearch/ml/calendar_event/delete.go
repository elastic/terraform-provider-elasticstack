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
	"errors"
	"fmt"

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func deleteCalendarEvent(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, _ CalendarEventTFModel) fwdiags.Diagnostics {
	var diags fwdiags.Diagnostics

	calendarID, eventID, splitDiags := ml.SplitCalendarResourcePath(resourceID, "<event_id>")
	diags.Append(splitDiags...)
	if diags.HasError() {
		return diags
	}

	tflog.Debug(ctx, fmt.Sprintf("Deleting ML calendar event %s from calendar: %s", eventID, calendarID))

	typedClient := client.GetESClient()

	_, err := typedClient.Ml.DeleteCalendarEvent(calendarID, eventID).Do(ctx)
	if err != nil {
		if esErr, ok := errors.AsType[*types.ElasticsearchError](err); ok && esErr.Status == 404 {
			tflog.Debug(ctx, fmt.Sprintf("ML calendar event %s already deleted from calendar: %s", eventID, calendarID))
			return diags
		}
		diags.AddError("Failed to delete ML calendar event", fmt.Sprintf("Unable to delete ML calendar event %s from calendar %s — %s", eventID, calendarID, err.Error()))
		return diags
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully deleted ML calendar event %s from calendar: %s", eventID, calendarID))
	return diags
}
