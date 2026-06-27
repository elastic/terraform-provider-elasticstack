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

package agentpolicy

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicyIDValidator(t *testing.T) {
	t.Parallel()

	valid255 := strings.Repeat("a", 255)
	tooLong := strings.Repeat("a", 256)

	tests := []struct {
		name      string
		value     types.String
		wantError bool
		wantMatch string
	}{
		{
			name:  "null value",
			value: types.StringNull(),
		},
		{
			name:  "unknown value",
			value: types.StringUnknown(),
		},
		{
			name:      "empty string",
			value:     types.StringValue(""),
			wantError: true,
			wantMatch: "policy_id must be between 1 and 255 characters",
		},
		{
			name:  "valid id",
			value: types.StringValue("my-valid-policy"),
		},
		{
			name:      "length 256",
			value:     types.StringValue(tooLong),
			wantError: true,
			wantMatch: "policy_id must be between 1 and 255 characters",
		},
		{
			name:  "length 255",
			value: types.StringValue(valid255),
		},
		{
			name:      "contains slash",
			value:     types.StringValue("bad/id"),
			wantError: true,
			wantMatch: `policy_id must not contain path separators ("/")`,
		},
		{
			name:      "contains traversal sequence",
			value:     types.StringValue("my..policy"),
			wantError: true,
			wantMatch: `policy_id must not contain traversal sequences ("..")`,
		},
		{
			name:      "bare __proto__",
			value:     types.StringValue("__proto__"),
			wantError: true,
			wantMatch: `policy_id must not contain reserved keys ("__proto__")`,
		},
		{
			name:      "bare constructor",
			value:     types.StringValue("constructor"),
			wantError: true,
			wantMatch: `policy_id must not contain reserved keys ("constructor")`,
		},
		{
			name:      "bare prototype",
			value:     types.StringValue("prototype"),
			wantError: true,
			wantMatch: `policy_id must not contain reserved keys ("prototype")`,
		},
		{
			name:      "contains __proto__ substring",
			value:     types.StringValue("my-__proto__-policy"),
			wantError: true,
			wantMatch: `policy_id must not contain reserved keys ("__proto__")`,
		},
		{
			name:      "contains constructor substring",
			value:     types.StringValue("my-constructor-policy"),
			wantError: true,
			wantMatch: `policy_id must not contain reserved keys ("constructor")`,
		},
		{
			name:      "contains prototype substring",
			value:     types.StringValue("my-prototype-policy"),
			wantError: true,
			wantMatch: `policy_id must not contain reserved keys ("prototype")`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var resp validator.StringResponse
			policyIDValidator{}.ValidateString(context.Background(), validator.StringRequest{
				Path:        path.Root("policy_id"),
				ConfigValue: tc.value,
			}, &resp)

			if tc.wantError {
				require.True(t, resp.Diagnostics.HasError(), "expected validation error")
				assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.wantMatch)
				return
			}

			assert.False(t, resp.Diagnostics.HasError(), "unexpected validation error: %v", resp.Diagnostics)
		})
	}
}
