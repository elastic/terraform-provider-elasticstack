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
	"slices"
	"testing"

	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestCalendarIDsFromTFSet_unknownOrNull(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("unknown", func(t *testing.T) {
		t.Parallel()
		out, diags := calendarIDsFromTFSet(ctx, types.SetUnknown(types.StringType))
		if diags.HasError() {
			t.Fatal(diags)
		}
		if out != nil {
			t.Fatalf("expected nil slice, got %#v", out)
		}
	})

	t.Run("null", func(t *testing.T) {
		t.Parallel()
		out, diags := calendarIDsFromTFSet(ctx, types.SetNull(types.StringType))
		if diags.HasError() {
			t.Fatal(diags)
		}
		if out != nil {
			t.Fatalf("expected nil slice, got %#v", out)
		}
	})
}

func TestCalendarIDsContainingJobFromPage(t *testing.T) {
	t.Parallel()
	job := "job-a"
	t.Run("empty page", func(t *testing.T) {
		t.Parallel()
		got := calendarIDsContainingJobFromPage(nil, job)
		if len(got) != 0 {
			t.Fatalf("got %#v", got)
		}
	})
	t.Run("filters and preserves order of append", func(t *testing.T) {
		t.Parallel()
		page := []estypes.Calendar{
			{CalendarId: "cal-1", JobIds: []string{"other"}},
			{CalendarId: "cal-2", JobIds: []string{job, "other"}},
			{CalendarId: "cal-3", JobIds: []string{}},
		}
		got := calendarIDsContainingJobFromPage(page, job)
		if !slices.Equal(got, []string{"cal-2"}) {
			t.Fatalf("got %#v", got)
		}
	})
}
