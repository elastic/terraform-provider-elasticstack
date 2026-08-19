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

	esv8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/ml/getcalendars"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

func readCalendar(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state TFModel) (TFModel, bool, fwdiags.Diagnostics) {
	calendarID := resourceID
	if calendarID == "" {
		var diags fwdiags.Diagnostics
		diags.AddError("Invalid resource ID", "calendar_id cannot be empty")
		return state, false, diags
	}

	return entitycore.SimpleElasticsearchRead[TFModel, getcalendars.Response](
		"ML calendar",
		func(ctx context.Context, typedClient *esv8.TypedClient, calendarID string) (*getcalendars.Response, error) {
			return typedClient.Ml.GetCalendars().CalendarId(calendarID).Do(ctx)
		},
		func(model *TFModel, _ context.Context, res *getcalendars.Response) (bool, fwdiags.Diagnostics) {
			if len(res.Calendars) == 0 {
				return false, nil
			}
			applyTypedCalendarToTFModel(model, &res.Calendars[0])
			return true, nil
		},
	)(ctx, client, calendarID, state)
}
