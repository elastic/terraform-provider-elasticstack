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

package typeutils_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatStrictDateTime(t *testing.T) {
	t.Parallel()

	ts := time.Date(2024, 3, 15, 10, 30, 45, 123000000, time.UTC)
	got := typeutils.FormatStrictDateTime(ts)
	require.Equal(t, "2024-03-15T10:30:45.123Z", got)
}

func TestTimeToStringValue(t *testing.T) {
	t.Parallel()

	ts := time.Date(2024, 3, 15, 10, 30, 45, 123000000, time.UTC)
	got := typeutils.TimeToStringValue(ts)
	require.Equal(t, types.StringValue("2024-03-15T10:30:45.123Z"), got)
}

func TestElasticDateTimeToMillis(t *testing.T) {
	t.Parallel()

	ms := time.Date(2026, 7, 1, 15, 30, 0, 0, time.FixedZone("EDT", -4*3600)).UnixMilli()

	t.Run("nil", func(t *testing.T) {
		_, ok := typeutils.ElasticDateTimeToMillis(nil)
		assert.False(t, ok)
	})

	t.Run("float64", func(t *testing.T) {
		got, ok := typeutils.ElasticDateTimeToMillis(float64(ms))
		require.True(t, ok)
		assert.Equal(t, ms, got)
	})

	t.Run("int64", func(t *testing.T) {
		got, ok := typeutils.ElasticDateTimeToMillis(ms)
		require.True(t, ok)
		assert.Equal(t, ms, got)
	})

	t.Run("int", func(t *testing.T) {
		got, ok := typeutils.ElasticDateTimeToMillis(int(ms))
		require.True(t, ok)
		assert.Equal(t, ms, got)
	})

	t.Run("uint64", func(t *testing.T) {
		got, ok := typeutils.ElasticDateTimeToMillis(uint64(ms))
		require.True(t, ok)
		assert.Equal(t, ms, got)
	})

	t.Run("json.Number int", func(t *testing.T) {
		n := json.Number(fmt.Sprintf("%d", ms))
		got, ok := typeutils.ElasticDateTimeToMillis(n)
		require.True(t, ok)
		assert.Equal(t, ms, got)
	})

	t.Run("json.Number float", func(t *testing.T) {
		n := json.Number(fmt.Sprintf("%.1f", float64(ms)))
		got, ok := typeutils.ElasticDateTimeToMillis(n)
		require.True(t, ok)
		assert.Equal(t, ms, got)
	})

	t.Run("json.Number invalid", func(t *testing.T) {
		_, ok := typeutils.ElasticDateTimeToMillis(json.Number("not-a-number"))
		assert.False(t, ok)
	})

	t.Run("RFC3339 string with offset", func(t *testing.T) {
		got, ok := typeutils.ElasticDateTimeToMillis("2026-07-01T15:30:00-04:00")
		require.True(t, ok)
		assert.Equal(t, ms, got)
	})

	t.Run("empty string", func(t *testing.T) {
		_, ok := typeutils.ElasticDateTimeToMillis("")
		assert.False(t, ok)
	})

	t.Run("invalid string", func(t *testing.T) {
		_, ok := typeutils.ElasticDateTimeToMillis("not-a-date")
		assert.False(t, ok)
	})

	t.Run("DateTime wrapping float64", func(t *testing.T) {
		// estypes.DateTime is `type DateTime any`; its concrete value is unwrapped by
		// Go's interface boxing, so DateTime(float64) is handled by the float64 case.
		got, ok := typeutils.ElasticDateTimeToMillis(estypes.DateTime(float64(ms)))
		require.True(t, ok)
		assert.Equal(t, ms, got)
	})

	t.Run("unsupported type bool", func(t *testing.T) {
		_, ok := typeutils.ElasticDateTimeToMillis(true)
		assert.False(t, ok)
	})

	t.Run("unsupported type map", func(t *testing.T) {
		_, ok := typeutils.ElasticDateTimeToMillis(map[string]any{"k": "v"})
		assert.False(t, ok)
	})
}

func TestElasticDateTimeToStringValue(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		input   any
		wantNil bool
		want    string
	}{
		{name: "nil", input: nil, wantNil: true},
		{name: "zero float64", input: estypes.DateTime(float64(0)), wantNil: true},
		{name: "epoch millis float64", input: estypes.DateTime(float64(1717243200000)), want: "2024-06-01T12:00:00.000Z"},
		{name: "RFC3339 string", input: estypes.DateTime("2024-06-01T12:00:00Z"), want: "2024-06-01T12:00:00.000Z"},
		{name: "empty string", input: estypes.DateTime(""), wantNil: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := typeutils.ElasticDateTimeToStringValue(tc.input)
			if tc.wantNil {
				assert.True(t, got.IsNull())
			} else {
				assert.False(t, got.IsNull())
				assert.Equal(t, tc.want, got.ValueString())
			}
		})
	}
}
