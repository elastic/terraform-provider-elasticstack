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

package slm

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestExpandWildcardsValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	v := expandWildcardsValidator{}

	cases := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"single valid", "open", false},
		{"multiple valid", "open,hidden", false},
		{"with spaces", "open, hidden", false},
		{"invalid value", "invalid", true},
		{"mixed valid and invalid", "open,invalid", true},
		{"empty string", "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := validator.StringRequest{
				ConfigValue: types.StringValue(tc.value),
			}
			var resp validator.StringResponse
			v.ValidateString(ctx, req, &resp)
			if tc.wantErr {
				require.True(t, resp.Diagnostics.HasError(), "expected error for value %q", tc.value)
			} else {
				require.False(t, resp.Diagnostics.HasError(), "unexpected error: %s", resp.Diagnostics)
			}
		})
	}
}

func TestExpandWildcardsValidator_NullAndUnknown(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	v := expandWildcardsValidator{}

	t.Run("null", func(t *testing.T) {
		t.Parallel()
		req := validator.StringRequest{
			ConfigValue: types.StringNull(),
		}
		var resp validator.StringResponse
		v.ValidateString(ctx, req, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("unknown", func(t *testing.T) {
		t.Parallel()
		req := validator.StringRequest{
			ConfigValue: types.StringUnknown(),
		}
		var resp validator.StringResponse
		v.ValidateString(ctx, req, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})
}
