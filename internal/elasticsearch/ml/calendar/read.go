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
	"errors"
	"fmt"

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
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

	tflog.Debug(ctx, fmt.Sprintf("Reading ML calendar: %s", calendarID))

	typedClient := client.GetESClient()

	res, err := typedClient.Ml.GetCalendars().CalendarId(calendarID).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return state, false, nil
		}
		diags.AddError("Failed to get ML calendar", fmt.Sprintf("Unable to get ML calendar: %s — %s", calendarID, err.Error()))
		return state, false, diags
	}

	if len(res.Calendars) == 0 {
		return state, false, nil
	}

	applyTypedCalendarToTFModel(&state, &res.Calendars[0])

	tflog.Debug(ctx, fmt.Sprintf("Successfully read ML calendar: %s", calendarID))
	return state, true, diags
}
