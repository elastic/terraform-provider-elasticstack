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
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/elastic/go-elasticsearch/v9"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

const mlCalendarEventsPageSize = 1000

// mlCalendarEventsPageFetcher loads one page of ML calendar events (used for tests and production).
type mlCalendarEventsPageFetcher interface {
	FetchMLCalendarEventsPage(ctx context.Context, calendarID string, from, size int) ([]calendarEventWire, error)
}

type typedClientCalendarEventsWindowFetcher struct {
	client     *elasticsearch.TypedClient
	start, end string
}

func (f typedClientCalendarEventsWindowFetcher) FetchMLCalendarEventsPage(ctx context.Context, calendarID string, from, size int) ([]calendarEventWire, error) {
	return fetchMLCalendarEventsPage(ctx, f.client, calendarID, from, size, f.start, f.end)
}

func fetchMLCalendarEventsPage(ctx context.Context, client *elasticsearch.TypedClient, calendarID string, from, size int, startRFC3339, endRFC3339 string) ([]calendarEventWire, error) {
	q := client.Ml.GetCalendarEvents(calendarID).From(from).Size(size)
	if startRFC3339 != "" {
		q = q.Start(startRFC3339)
	}
	if endRFC3339 != "" {
		q = q.End(endRFC3339)
	}
	res, err := q.Perform(ctx)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if from == 0 && res.StatusCode == http.StatusNotFound {
		return []calendarEventWire{}, nil
	}
	if res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("unable to list ML calendar events for calendar %s (offset %d): status %d: %s", calendarID, from, res.StatusCode, string(body))
	}
	var envelope struct {
		Events []calendarEventWire `json:"events"`
	}
	if err := json.NewDecoder(res.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("decode ML calendar events for calendar %s: %w", calendarID, err)
	}
	return envelope.Events, nil
}

// walkMLCalendarEventPagesWithWindow calls fn for each page of calendar events until fn returns
// true, an error occurs, or there are no more events. When startRFC3339 and endRFC3339 are empty,
// the full calendar is listed (no start/end query params).
func walkMLCalendarEventPagesWithWindow(
	ctx context.Context,
	typedClient *elasticsearch.TypedClient,
	calendarID, startRFC3339, endRFC3339 string,
	fn func([]calendarEventWire) (stop bool),
) fwdiags.Diagnostics {
	return walkMLCalendarEventPagesWith(ctx, typedClientCalendarEventsWindowFetcher{
		client: typedClient,
		start:  startRFC3339,
		end:    endRFC3339,
	}, calendarID, fn)
}

func walkMLCalendarEventPagesWith(ctx context.Context, fetcher mlCalendarEventsPageFetcher, calendarID string, fn func([]calendarEventWire) (stop bool)) fwdiags.Diagnostics {
	var diags fwdiags.Diagnostics
	for from := 0; ; from += mlCalendarEventsPageSize {
		events, err := fetcher.FetchMLCalendarEventsPage(ctx, calendarID, from, mlCalendarEventsPageSize)
		if err != nil {
			diags.AddError(
				"Failed to list ML calendar events",
				fmt.Sprintf("Unable to list events for calendar %s (offset %d) — %s", calendarID, from, err.Error()),
			)
			return diags
		}
		if len(events) == 0 {
			return diags
		}
		if fn(events) {
			return diags
		}
		if len(events) < mlCalendarEventsPageSize {
			return diags
		}
	}
}
