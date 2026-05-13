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

package anomalydetectionjob

import (
	"context"
	"errors"
	"fmt"
	"slices"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

const getCalendarsPageSize = 500

// listCalendarIDsForJob returns sorted calendar IDs that include jobID in their job_ids.
func listCalendarIDsForJob(ctx context.Context, typed *elasticsearch.TypedClient, jobID string) ([]string, error) {
	var out []string
	from := 0
	for {
		res, err := typed.Ml.GetCalendars().CalendarId("*").From(from).Size(getCalendarsPageSize).Do(ctx)
		if err != nil {
			return nil, fmt.Errorf("ml get calendars: %w", err)
		}
		if len(res.Calendars) == 0 {
			break
		}
		for _, cal := range res.Calendars {
			if slices.Contains(cal.JobIds, jobID) {
				out = append(out, cal.CalendarId)
			}
		}
		if len(res.Calendars) < getCalendarsPageSize {
			break
		}
		from += getCalendarsPageSize
	}
	slices.Sort(out)
	out = slices.Compact(out)
	return out, nil
}

// syncJobCalendars applies calendar–job membership: adds jobID to calendars in desired but not
// previous, and removes jobID from calendars in previous but not desired.
func syncJobCalendars(ctx context.Context, typed *elasticsearch.TypedClient, jobID string, desired, previous []string) error {
	prev := append([]string(nil), previous...)
	want := append([]string(nil), desired...)

	for _, cal := range prev {
		if slices.Contains(want, cal) {
			continue
		}
		_, err := typed.Ml.DeleteCalendarJob(cal, jobID).Do(ctx)
		if err != nil {
			var esErr *types.ElasticsearchError
			if errors.As(err, &esErr) && esErr.Status == 404 {
				continue
			}
			return fmt.Errorf("remove job %q from calendar %q: %w", jobID, cal, err)
		}
	}

	for _, cal := range want {
		if slices.Contains(prev, cal) {
			continue
		}
		_, err := typed.Ml.PutCalendarJob(cal, jobID).Do(ctx)
		if err != nil {
			return fmt.Errorf("add job %q to calendar %q: %w", jobID, cal, err)
		}
	}
	return nil
}
