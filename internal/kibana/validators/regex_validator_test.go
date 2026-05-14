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
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestRegexStringValidatorDescription(t *testing.T) {
	t.Parallel()

	v := regexStringValidator{description: "test description"}
	ctx := context.Background()

	if got := v.Description(ctx); got != "test description" {
		t.Errorf("Description() = %q, want %q", got, "test description")
	}
	if got := v.MarkdownDescription(ctx); got != "test description" {
		t.Errorf("MarkdownDescription() = %q, want %q", got, "test description")
	}
}

func TestRegexStringValidatorValidateString(t *testing.T) {
	t.Parallel()

	matchAlpha := func(s string) (bool, error) {
		for _, c := range s {
			if c < 'a' || c > 'z' {
				return false, nil
			}
		}
		return len(s) > 0, nil
	}
	errMatchFn := func(_ string) (bool, error) {
		return false, fmt.Errorf("match error")
	}

	v := regexStringValidator{
		description: "lowercase letters only",
		errSummary:  "invalid value",
		errDetail:   "must be lowercase letters",
		matchFn:     matchAlpha,
	}

	tests := []struct {
		name      string
		value     types.String
		wantError bool
	}{
		{
			name:      "null value skipped",
			value:     types.StringNull(),
			wantError: false,
		},
		{
			name:      "unknown value skipped",
			value:     types.StringUnknown(),
			wantError: false,
		},
		{
			name:      "valid value",
			value:     types.StringValue("abc"),
			wantError: false,
		},
		{
			name:      "invalid value",
			value:     types.StringValue("ABC"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validator.StringRequest{ConfigValue: tt.value}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)
			if tt.wantError && !resp.Diagnostics.HasError() {
				t.Error("expected error but got none")
			}
			if !tt.wantError && resp.Diagnostics.HasError() {
				t.Errorf("unexpected error: %v", resp.Diagnostics)
			}
		})
	}

	t.Run("match function error triggers error diagnostic", func(t *testing.T) {
		errV := regexStringValidator{
			description: "desc",
			errSummary:  "summary",
			errDetail:   "detail",
			matchFn:     errMatchFn,
		}
		req := validator.StringRequest{ConfigValue: types.StringValue("abc")}
		resp := &validator.StringResponse{}
		errV.ValidateString(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Error("expected error diagnostic when matchFn returns error")
		}
	})
}
