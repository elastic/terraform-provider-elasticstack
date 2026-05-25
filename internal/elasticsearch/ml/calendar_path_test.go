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

package ml

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitCalendarResourcePath(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		_, _, diags := SplitCalendarResourcePath("", "<event_id>")
		require.True(t, diags.HasError())
	})

	t.Run("missing slash", func(t *testing.T) {
		_, _, diags := SplitCalendarResourcePath("only-calendar", "<event_id>")
		require.True(t, diags.HasError())
	})

	t.Run("empty calendar_id", func(t *testing.T) {
		_, _, diags := SplitCalendarResourcePath("/event-1", "<event_id>")
		require.True(t, diags.HasError())
	})

	t.Run("empty sub_resource_id", func(t *testing.T) {
		_, _, diags := SplitCalendarResourcePath("cal-1/", "<event_id>")
		require.True(t, diags.HasError())
	})

	t.Run("valid event path", func(t *testing.T) {
		cal, sub, diags := SplitCalendarResourcePath("my-cal/evt-1", "<event_id>")
		require.False(t, diags.HasError())
		assert.Equal(t, "my-cal", cal)
		assert.Equal(t, "evt-1", sub)
	})

	t.Run("valid job path", func(t *testing.T) {
		cal, sub, diags := SplitCalendarResourcePath("my-cal/job-1", "<job_id>")
		require.False(t, diags.HasError())
		assert.Equal(t, "my-cal", cal)
		assert.Equal(t, "job-1", sub)
	})

	t.Run("sub_resource_id with slashes", func(t *testing.T) {
		cal, sub, diags := SplitCalendarResourcePath("my-cal/evt/with/slashes", "<event_id>")
		require.False(t, diags.HasError())
		assert.Equal(t, "my-cal", cal)
		assert.Equal(t, "evt/with/slashes", sub)
	})

	t.Run("error message includes label", func(t *testing.T) {
		_, _, diags := SplitCalendarResourcePath("bad", "<job_id>")
		require.True(t, diags.HasError())
		assert.Contains(t, diags[0].Detail(), "<job_id>")
	})
}
