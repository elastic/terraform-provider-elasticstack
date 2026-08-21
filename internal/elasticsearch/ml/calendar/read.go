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

package calendar

import (
	"context"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/typedapi/ml/getcalendars"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func readCalendar(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state TFModel) (TFModel, bool, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	calendarID := resourceID
	if calendarID == "" {
		diags.AddError("Invalid resource ID", "calendar_id cannot be empty")
		return state, false, diags
	}

	typedClient := client.GetESClient()

	res, found, diags := ml.ReadWithNotFoundAsAbsent(ctx, "ML calendar", calendarID, func() (*getcalendars.Response, error) {
		return typedClient.Ml.GetCalendars().CalendarId(calendarID).Do(ctx)
	})
	if diags.HasError() {
		return state, false, diags
	}
	if !found || len(res.Calendars) == 0 {
		return state, false, nil
	}

	applyTypedCalendarToTFModel(&state, &res.Calendars[0])

	tflog.Debug(ctx, fmt.Sprintf("Successfully read ML calendar: %s", calendarID))
	return state, true, diags
}
