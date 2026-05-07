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
