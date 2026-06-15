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

package kibanacustomtypes

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

// withUnits returns a copy of v carrying the given units. It exists purely as
// a test helper so individual cases can assert ValidateAttribute behaviour
// without going through ValueFromString.
func withUnits(v AlertingDuration, units AlertingDurationUnits) AlertingDuration {
	v.units = units
	return v
}

func TestAlertingDuration_Type(t *testing.T) {
	require.Equal(t, AlertingDurationType{}, AlertingDuration{}.Type(context.Background()))
	require.Equal(t,
		AlertingDurationType{Units: AlertingDurationUnitsSubDay},
		withUnits(NewAlertingDurationValue("1d"), AlertingDurationUnitsSubDay).Type(context.Background()),
	)
}

func TestAlertingDuration_Equal(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
		val      AlertingDuration
		other    attr.Value
	}{
		{
			name:     "not equal to a non-AlertingDuration value",
			expected: false,
			val:      NewAlertingDurationValue("1d"),
			other:    basetypes.NewStringValue("1d"),
		},
		{
			name:     "not equal when string values differ (even if semantically equal)",
			expected: false,
			val:      NewAlertingDurationValue("1d"),
			other:    NewAlertingDurationValue("24h"),
		},
		{
			name:     "equal when string values match",
			expected: true,
			val:      NewAlertingDurationValue("1d"),
			other:    NewAlertingDurationValue("1d"),
		},
		{
			name:     "equal regardless of units (Equal is logical/string equality)",
			expected: true,
			val:      withUnits(NewAlertingDurationValue("1d"), AlertingDurationUnitsSubDay),
			other:    withUnits(NewAlertingDurationValue("1d"), IntervalFrequencyUnits),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.val.Equal(tt.other))
		})
	}
}

func TestAlertingDuration_Parse(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{input: "30s", expected: 30 * time.Second},
		{input: "1m", expected: time.Minute},
		{input: "2h", expected: 2 * time.Hour},
		{input: "1d", expected: 24 * time.Hour},
		{input: "24h", expected: 24 * time.Hour},
		{input: "1w", expected: 7 * 24 * time.Hour},
		{input: "7d", expected: 7 * 24 * time.Hour},
		{input: "168h", expected: 7 * 24 * time.Hour},
		{input: "1M", expected: 30 * 24 * time.Hour},
		{input: "1y", expected: 365 * 24 * time.Hour},
		{input: "", wantErr: true},
		{input: "d", wantErr: true},
		{input: "1", wantErr: true},
		{input: "0d", wantErr: true},
		{input: "01d", wantErr: true},
		{input: "1.5d", wantErr: true},
		{input: "-1d", wantErr: true},
		{input: "1x", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseAlertingDuration(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestAlertingDuration_ValidateAttribute(t *testing.T) {
	tests := []struct {
		name     string
		value    AlertingDuration
		wantDiag bool
	}{
		// Defaults: zero-value units accepts every parser-supported unit.
		{name: "null is valid", value: NewAlertingDurationNull()},
		{name: "unknown is valid", value: NewAlertingDurationUnknown()},
		{name: "1d with default units", value: NewAlertingDurationValue("1d")},
		{name: "1w with default units", value: NewAlertingDurationValue("1w")},
		{name: "garbage is invalid", value: NewAlertingDurationValue("nope"), wantDiag: true},
		{name: "missing unit is invalid", value: NewAlertingDurationValue("1"), wantDiag: true},
		{name: "zero is invalid", value: NewAlertingDurationValue("0d"), wantDiag: true},

		// AlertingDurationUnitsSubDay (s, m, h, d): s/m/h/d pass, w/M/y reject.
		{name: "subday accepts 30s", value: withUnits(NewAlertingDurationValue("30s"), AlertingDurationUnitsSubDay)},
		{name: "subday accepts 1d", value: withUnits(NewAlertingDurationValue("1d"), AlertingDurationUnitsSubDay)},
		{name: "subday rejects 1w", value: withUnits(NewAlertingDurationValue("1w"), AlertingDurationUnitsSubDay), wantDiag: true},
		{name: "subday rejects 1M", value: withUnits(NewAlertingDurationValue("1M"), AlertingDurationUnitsSubDay), wantDiag: true},

		// IntervalFrequencyUnits (d, w, M, y): d/w/M/y pass, s/m/h reject.
		{name: "interval accepts 1d", value: withUnits(NewAlertingDurationValue("1d"), IntervalFrequencyUnits)},
		{name: "interval accepts 1w", value: withUnits(NewAlertingDurationValue("1w"), IntervalFrequencyUnits)},
		{name: "interval accepts 1y", value: withUnits(NewAlertingDurationValue("1y"), IntervalFrequencyUnits)},
		{name: "interval rejects 30s", value: withUnits(NewAlertingDurationValue("30s"), IntervalFrequencyUnits), wantDiag: true},
		{name: "interval rejects 1h", value: withUnits(NewAlertingDurationValue("1h"), IntervalFrequencyUnits), wantDiag: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := xattr.ValidateAttributeResponse{}
			tt.value.ValidateAttribute(
				context.Background(),
				xattr.ValidateAttributeRequest{Path: path.Root("duration")},
				&resp,
			)
			require.Equal(t, tt.wantDiag, resp.Diagnostics.HasError(), "diags: %v", resp.Diagnostics)
		})
	}
}

func TestAlertingDuration_StringSemanticEquals(t *testing.T) {
	tests := []struct {
		name      string
		val       AlertingDuration
		other     basetypes.StringValuable
		wantEqual bool
		wantErr   bool
	}{
		{
			name:    "non-AlertingDuration operand produces error diag",
			val:     NewAlertingDurationValue("1d"),
			other:   basetypes.NewStringValue("1d"),
			wantErr: true,
		},
		{name: "null == null", val: NewAlertingDurationNull(), other: NewAlertingDurationNull(), wantEqual: true},
		{name: "null != unknown", val: NewAlertingDurationNull(), other: NewAlertingDurationUnknown()},
		{name: "null != value", val: NewAlertingDurationNull(), other: NewAlertingDurationValue("1d")},
		{name: "unknown == unknown", val: NewAlertingDurationUnknown(), other: NewAlertingDurationUnknown(), wantEqual: true},
		{name: "unknown != value", val: NewAlertingDurationUnknown(), other: NewAlertingDurationValue("1d")},
		{name: "identical strings are equal", val: NewAlertingDurationValue("1d"), other: NewAlertingDurationValue("1d"), wantEqual: true},
		{name: "1d == 24h", val: NewAlertingDurationValue("1d"), other: NewAlertingDurationValue("24h"), wantEqual: true},
		{name: "24h == 1d (reverse direction)", val: NewAlertingDurationValue("24h"), other: NewAlertingDurationValue("1d"), wantEqual: true},
		{name: "1w == 7d", val: NewAlertingDurationValue("1w"), other: NewAlertingDurationValue("7d"), wantEqual: true},
		{name: "1w == 168h", val: NewAlertingDurationValue("1w"), other: NewAlertingDurationValue("168h"), wantEqual: true},
		{name: "60s == 1m", val: NewAlertingDurationValue("60s"), other: NewAlertingDurationValue("1m"), wantEqual: true},
		{name: "30s != 30m", val: NewAlertingDurationValue("30s"), other: NewAlertingDurationValue("30m")},
		{name: "1M != 1m", val: NewAlertingDurationValue("1M"), other: NewAlertingDurationValue("1m")},
		{name: "invalid lhs produces error", val: NewAlertingDurationValue("bad"), other: NewAlertingDurationValue("1d"), wantErr: true},
		{name: "invalid rhs produces error", val: NewAlertingDurationValue("1d"), other: NewAlertingDurationValue("bad"), wantErr: true},
		{
			name:      "values with different unit sets compare on parsed duration only",
			val:       withUnits(NewAlertingDurationValue("1d"), AlertingDurationUnitsSubDay),
			other:     withUnits(NewAlertingDurationValue("24h"), IntervalFrequencyUnits),
			wantEqual: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			equal, diags := tt.val.StringSemanticEquals(context.Background(), tt.other)
			require.Equal(t, tt.wantErr, diags.HasError(), "diags: %v", diags)
			require.Equal(t, tt.wantEqual, equal)
		})
	}
}

func TestAlertingDurationUnits_DescribeAndEqual(t *testing.T) {
	require.Equal(t, "`s`, `m`, `h`, or `d`", AlertingDurationUnitsSubDay.Describe())
	require.Equal(t, "`d`, `w`, `M`, or `y`", IntervalFrequencyUnits.Describe())
	require.Equal(t, "`s`, `m`, `h`, `d`, `w`, `M`, or `y`", AllAlertingDurationUnits.Describe())

	require.True(t, AlertingDurationUnitsSubDay.Equal(newAlertingDurationUnits('d', 'h', 'm', 's')))
	require.False(t, AlertingDurationUnitsSubDay.Equal(IntervalFrequencyUnits))

	// Allowed honours the configured set; zero-value falls back to "everything".
	require.True(t, AlertingDurationUnitsSubDay.Allowed('d'))
	require.False(t, AlertingDurationUnitsSubDay.Allowed('w'))
	require.True(t, AlertingDurationUnits{}.Allowed('y'))
}

func TestAlertingDurationType_Equal(t *testing.T) {
	require.True(t,
		AlertingDurationType{Units: AlertingDurationUnitsSubDay}.Equal(AlertingDurationType{Units: AlertingDurationUnitsSubDay}),
	)
	require.False(t,
		AlertingDurationType{Units: AlertingDurationUnitsSubDay}.Equal(AlertingDurationType{Units: IntervalFrequencyUnits}),
	)
	require.False(t, AlertingDurationType{}.Equal(basetypes.StringType{}))
}
