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

package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestIsURL_ValidateString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		validator   validator.String
		value       types.String
		expectError bool
	}{
		{
			name:        "valid https URL accepted",
			validator:   IsURL("http", "https"),
			value:       types.StringValue("https://s3.example.com"),
			expectError: false,
		},
		{
			name:        "valid http URL with port accepted",
			validator:   IsURL("http", "https"),
			value:       types.StringValue("http://s3.example.com:9000"),
			expectError: false,
		},
		{
			name:        "https:// with no host rejected",
			validator:   IsURL("http", "https"),
			value:       types.StringValue("https://"),
			expectError: true,
		},
		{
			name:        "http:example rejected (no authority)",
			validator:   IsURL("http", "https"),
			value:       types.StringValue("http:example"),
			expectError: true,
		},
		{
			name:        "disallowed scheme rejected before host check",
			validator:   IsURL("http", "https"),
			value:       types.StringValue("ftp://example.com"),
			expectError: true,
		},
		{
			name:        "null value skips validation",
			validator:   IsURL("http", "https"),
			value:       types.StringNull(),
			expectError: false,
		},
		{
			name:        "empty string skips validation",
			validator:   IsURL("http", "https"),
			value:       types.StringValue(""),
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resp := &validator.StringResponse{}
			tc.validator.ValidateString(context.Background(), validator.StringRequest{
				Path:        path.Root("v"),
				ConfigValue: tc.value,
			}, resp)
			if tc.expectError {
				require.True(t, resp.Diagnostics.HasError(), "expected error")
			} else {
				require.False(t, resp.Diagnostics.HasError(), "unexpected error: %s", resp.Diagnostics)
			}
		})
	}
}

func TestIsURI_ValidateString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		validator   validator.String
		value       types.String
		expectError bool
	}{
		{
			name:        "null value skips validation",
			validator:   IsURI(),
			value:       types.StringNull(),
			expectError: false,
		},
		{
			name:        "unknown value skips validation",
			validator:   IsURI(),
			value:       types.StringUnknown(),
			expectError: false,
		},
		{
			name:        "empty string skips validation",
			validator:   IsURI(),
			value:       types.StringValue(""),
			expectError: false,
		},
		{
			name:        "no scheme rejected",
			validator:   IsURI(),
			value:       types.StringValue("example.com"),
			expectError: true,
		},
		{
			name:        "any scheme accepted when no list provided",
			validator:   IsURI(),
			value:       types.StringValue("custom://example.com"),
			expectError: false,
		},
		{
			name:        "http allowed when in list",
			validator:   IsURI("http", "https"),
			value:       types.StringValue("http://s3.example.com:9000"),
			expectError: false,
		},
		{
			name:        "https allowed when in list",
			validator:   IsURI("http", "https"),
			value:       types.StringValue("https://s3.example.com"),
			expectError: false,
		},
		{
			name:        "disallowed scheme rejected",
			validator:   IsURI("http", "https"),
			value:       types.StringValue("ftp://s3.example.com"),
			expectError: true,
		},
		{
			name:        "scheme comparison is case-insensitive",
			validator:   IsURI("http", "https"),
			value:       types.StringValue("HTTPS://example.com"),
			expectError: false,
		},
		{
			name:        "file URI without authority accepted",
			validator:   IsURI("file", "ftp", "http", "https", "jar"),
			value:       types.StringValue("file:/tmp"),
			expectError: false,
		},
		{
			name:        "jar URI accepted",
			validator:   IsURI("file", "ftp", "http", "https", "jar"),
			value:       types.StringValue("jar:file:/tmp/repo.jar!/"),
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resp := &validator.StringResponse{}
			tc.validator.ValidateString(context.Background(), validator.StringRequest{
				Path:        path.Root("v"),
				ConfigValue: tc.value,
			}, resp)
			if tc.expectError {
				require.True(t, resp.Diagnostics.HasError(), "expected error")
			} else {
				require.False(t, resp.Diagnostics.HasError(), "unexpected error: %s", resp.Diagnostics)
			}
		})
	}
}
