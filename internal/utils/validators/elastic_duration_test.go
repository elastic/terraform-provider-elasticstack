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

package validators_test

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestElasticDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		value       types.String
		expectError bool
		errSummary  string
		errDetail   string
	}{
		{name: "null skipped", value: types.StringNull()},
		{name: "unknown skipped", value: types.StringUnknown()},
		{name: "valid days", value: types.StringValue("7d")},
		{name: "valid hours", value: types.StringValue("12h")},
		{name: "valid minutes", value: types.StringValue("30m")},
		{name: "valid seconds", value: types.StringValue("60s")},
		{name: "valid milliseconds", value: types.StringValue("500ms")},
		{name: "valid microseconds", value: types.StringValue("250micros")},
		{name: "valid nanoseconds", value: types.StringValue("100nanos")},
		{name: "valid fractional", value: types.StringValue("1.5h")},
		{
			name:        "empty string",
			value:       types.StringValue(""),
			expectError: true,
			errSummary:  "Invalid Elastic duration",
			errDetail:   "duration must not be empty",
		},
		{
			name:        "unsupported unit weeks",
			value:       types.StringValue("2w"),
			expectError: true,
			errSummary:  "Invalid Elastic duration",
			errDetail:   `"2w" is not a valid Elastic time-unit duration`,
		},
		{
			name:        "missing leading digit",
			value:       types.StringValue(".5s"),
			expectError: true,
			errSummary:  "Invalid Elastic duration",
		},
		{
			name:        "trailing garbage",
			value:       types.StringValue("30s "),
			expectError: true,
			errSummary:  "Invalid Elastic duration",
		},
		{
			name:        "missing unit",
			value:       types.StringValue("30"),
			expectError: true,
			errSummary:  "Invalid Elastic duration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := validator.StringRequest{
				Path:        path.Root("test"),
				ConfigValue: tt.value,
			}
			resp := &validator.StringResponse{}
			validators.ElasticDuration().ValidateString(context.Background(), req, resp)

			if !tt.expectError {
				require.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %s", resp.Diagnostics)
				return
			}

			require.True(t, resp.Diagnostics.HasError(), "expected an error diagnostic")
			require.Len(t, resp.Diagnostics, 1)
			require.Equal(t, tt.errSummary, resp.Diagnostics[0].Summary())
			if tt.errDetail != "" {
				require.Equal(t, tt.errDetail, resp.Diagnostics[0].Detail())
			}
		})
	}
}

func TestElasticDuration_Description(t *testing.T) {
	t.Parallel()

	v := validators.ElasticDuration()
	require.Equal(t, v.Description(context.Background()), v.MarkdownDescription(context.Background()))
	require.Contains(t, v.Description(context.Background()), "Elastic duration")
}

func TestParseUnitDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		units     string
		wantValue int
		wantUnit  string
		wantErr   bool
	}{
		{name: "valid minutes", input: "5m", units: "mhd", wantValue: 5, wantUnit: "m"},
		{name: "valid hours", input: "3h", units: "mhd", wantValue: 3, wantUnit: "h"},
		{name: "zero is accepted", input: "0m", units: "mhd", wantValue: 0, wantUnit: "m"},
		{name: "multi-digit", input: "150s", units: "smhd", wantValue: 150, wantUnit: "s"},
		{name: "empty string", input: "", units: "mhd", wantErr: true},
		{name: "missing digits", input: "m", units: "mhd", wantErr: true},
		{name: "missing unit", input: "30", units: "mhd", wantErr: true},
		{name: "unsupported unit", input: "30w", units: "mhd", wantErr: true},
		{name: "fractional rejected", input: "1.5m", units: "mhd", wantErr: true},
		{name: "empty units does not panic", input: "5m", units: "", wantErr: true},
		{name: "metacharacter units does not panic", input: "5m", units: "m-h^d", wantValue: 5, wantUnit: "m"},
		{name: "metacharacter units rejects unmatched", input: "5s", units: "m-h^d", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			value, unit, err := validators.ParseUnitDuration(tt.input, tt.units)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantValue, value)
			require.Equal(t, tt.wantUnit, unit)
		})
	}
}

func TestDurationWithUnits(t *testing.T) {
	t.Parallel()

	v := validators.DurationWithUnits("mhd", "must be digits plus m, h, or d")

	tests := []struct {
		name        string
		value       types.String
		expectError bool
	}{
		{name: "null skipped", value: types.StringNull()},
		{name: "unknown skipped", value: types.StringUnknown()},
		{name: "valid", value: types.StringValue("5m")},
		{name: "zero accepted", value: types.StringValue("0d")},
		{name: "invalid unit", value: types.StringValue("5s"), expectError: true},
		{name: "empty", value: types.StringValue(""), expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := validator.StringRequest{
				Path:        path.Root("test"),
				ConfigValue: tt.value,
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			if tt.expectError {
				require.True(t, resp.Diagnostics.HasError())
				return
			}
			require.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %s", resp.Diagnostics)
		})
	}
}
