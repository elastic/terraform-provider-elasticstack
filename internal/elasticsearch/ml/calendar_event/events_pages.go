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

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

const mlCalendarEventsPageSize = 1000

// walkMLCalendarEventPages calls fn for each page of calendar events until fn returns true,
// an error occurs, or there are no more events.
func walkMLCalendarEventPages(ctx context.Context, typedClient *elasticsearch.TypedClient, calendarID string, fn func([]types.CalendarEvent) (stop bool)) fwdiags.Diagnostics {
	var diags fwdiags.Diagnostics
	for from := 0; ; from += mlCalendarEventsPageSize {
		res, err := typedClient.Ml.GetCalendarEvents(calendarID).From(from).Size(mlCalendarEventsPageSize).Do(ctx)
		if err != nil {
			var esErr *types.ElasticsearchError
			if from == 0 && errors.As(err, &esErr) && esErr.Status == 404 {
				return diags
			}
			diags.AddError(
				"Failed to list ML calendar events",
				fmt.Sprintf("Unable to list events for calendar %s (offset %d) — %s", calendarID, from, err.Error()),
			)
			return diags
		}
		if len(res.Events) == 0 {
			return diags
		}
		if fn(res.Events) {
			return diags
		}
		if len(res.Events) < mlCalendarEventsPageSize {
			return diags
		}
	}
}
