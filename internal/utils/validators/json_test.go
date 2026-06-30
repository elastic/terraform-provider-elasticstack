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

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStringIsJSONObject(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		validator StringIsJSONObject
		value     string
		wantError bool
	}{
		{
			name:      "empty object with zero-value passes",
			validator: StringIsJSONObject{},
			value:     "{}",
			wantError: false,
		},
		{
			name:      "empty object with NonEmpty fails",
			validator: StringIsJSONObject{NonEmpty: true},
			value:     "{}",
			wantError: true,
		},
		{
			name:      "array fails",
			validator: StringIsJSONObject{},
			value:     "[]",
			wantError: true,
		},
		{
			name:      "string fails",
			validator: StringIsJSONObject{},
			value:     `"hello"`,
			wantError: true,
		},
		{
			name:      "number fails",
			validator: StringIsJSONObject{},
			value:     "123",
			wantError: true,
		},
		{
			name:      "null with zero-value fails",
			validator: StringIsJSONObject{},
			value:     "null",
			wantError: true,
		},
		{
			name:      "null with NonEmpty fails as non-object",
			validator: StringIsJSONObject{NonEmpty: true},
			value:     "null",
			wantError: true,
		},
		{
			name:      "valid object with keys and zero-value passes",
			validator: StringIsJSONObject{},
			value:     `{"properties":{"title":{"type":"keyword"}}}`,
			wantError: false,
		},
		{
			name:      "valid object with keys and NonEmpty passes",
			validator: StringIsJSONObject{NonEmpty: true},
			value:     `{"properties":{"title":{"type":"keyword"}}}`,
			wantError: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := validator.StringRequest{
				ConfigValue: types.StringValue(tc.value),
			}
			var resp validator.StringResponse
			tc.validator.ValidateString(context.Background(), req, &resp)

			hasError := resp.Diagnostics.HasError()
			if hasError != tc.wantError {
				t.Errorf("ValidateString(%q) error = %v, wantError = %v; diagnostics: %v", tc.value, hasError, tc.wantError, resp.Diagnostics)
			}
		})
	}
}
