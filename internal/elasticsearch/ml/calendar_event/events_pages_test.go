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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubEventsPager struct {
	fetch func(ctx context.Context, calendarID string, from, size int) ([]calendarEventWire, error)
}

func (s stubEventsPager) FetchMLCalendarEventsPage(ctx context.Context, calendarID string, from, size int) ([]calendarEventWire, error) {
	return s.fetch(ctx, calendarID, from, size)
}

func TestWalkMLCalendarEventPagesWith_firstPageEmptyNoDiag(t *testing.T) {
	ctx := context.Background()
	p := stubEventsPager{fetch: func(_ context.Context, _ string, from, _ int) ([]calendarEventWire, error) {
		require.Equal(t, 0, from)
		return []calendarEventWire{}, nil
	}}
	var calls int
	diags := walkMLCalendarEventPagesWith(ctx, p, "cal", func([]calendarEventWire) bool {
		calls++
		return false
	})
	assert.False(t, diags.HasError())
	assert.Equal(t, 0, calls)
}

func TestWalkMLCalendarEventPagesWith_firstPageNon404Error(t *testing.T) {
	ctx := context.Background()
	p := stubEventsPager{fetch: func(_ context.Context, _ string, from, _ int) ([]calendarEventWire, error) {
		require.Equal(t, 0, from)
		return nil, fmt.Errorf("network down")
	}}
	diags := walkMLCalendarEventPagesWith(ctx, p, "cal", func([]calendarEventWire) bool { return false })
	require.True(t, diags.HasError())
	assert.Contains(t, diags.Errors()[0].Summary(), "Failed to list ML calendar events")
}

func TestWalkMLCalendarEventPagesWith_stopAfterFirstPage(t *testing.T) {
	ctx := context.Background()
	id := "only-one"
	p := stubEventsPager{fetch: func(_ context.Context, _ string, from, _ int) ([]calendarEventWire, error) {
		require.Equal(t, 0, from)
		return []calendarEventWire{{EventID: &id}}, nil
	}}
	var saw int
	diags := walkMLCalendarEventPagesWith(ctx, p, "cal", func(ev []calendarEventWire) bool {
		saw += len(ev)
		return true
	})
	assert.False(t, diags.HasError())
	assert.Equal(t, 1, saw)
}

func TestWalkMLCalendarEventPagesWith_multiPage(t *testing.T) {
	ctx := context.Background()
	p := stubEventsPager{fetch: func(_ context.Context, _ string, from, size int) ([]calendarEventWire, error) {
		require.Equal(t, mlCalendarEventsPageSize, size)
		switch from {
		case 0:
			out := make([]calendarEventWire, mlCalendarEventsPageSize)
			for i := range out {
				s := fmt.Sprintf("id-%d", i)
				out[i] = calendarEventWire{EventID: &s}
			}
			return out, nil
		case mlCalendarEventsPageSize:
			s := "last-page"
			return []calendarEventWire{{EventID: &s}}, nil
		default:
			return nil, fmt.Errorf("unexpected from=%d", from)
		}
	}}
	var total int
	diags := walkMLCalendarEventPagesWith(ctx, p, "cal", func(ev []calendarEventWire) bool {
		total += len(ev)
		return false
	})
	assert.False(t, diags.HasError())
	assert.Equal(t, mlCalendarEventsPageSize+1, total)
}

func TestWalkMLCalendarEventPagesWith_secondPageError(t *testing.T) {
	ctx := context.Background()
	p := stubEventsPager{fetch: func(_ context.Context, _ string, from, _ int) ([]calendarEventWire, error) {
		if from == 0 {
			out := make([]calendarEventWire, mlCalendarEventsPageSize)
			for i := range out {
				s := fmt.Sprintf("id-%d", i)
				out[i] = calendarEventWire{EventID: &s}
			}
			return out, nil
		}
		return nil, fmt.Errorf("boom on page 2")
	}}
	diags := walkMLCalendarEventPagesWith(ctx, p, "cal", func([]calendarEventWire) bool { return false })
	require.True(t, diags.HasError())
	assert.Contains(t, diags.Errors()[0].Detail(), "boom on page 2")
}
