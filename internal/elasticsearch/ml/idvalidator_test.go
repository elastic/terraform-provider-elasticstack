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
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPathIDRegexp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id   string
		ok   bool
		note string
	}{
		{"a", true, "single char"},
		{"job1", true, "alphanumeric"},
		{"my.job.id", true, "dots in middle (Elasticsearch job id style)"},
		{"cal_1.a-2", true, "mix of dot underscore hyphen"},
		{"a_b", true, "underscore"},
		{"a-b", true, "hyphen"},
		{"a--b", true, "repeated hyphens"},
		{"a..b", true, "repeated dots"},
		{".ab", false, "must not start with dot"},
		{"ab.", false, "must not end with dot"},
		{"a_b.", false, "must not end with dot"},
		{"_ab", false, "must not start with underscore"},
		{"ab_", false, "must not end with underscore"},
		{"Abc", false, "uppercase rejected"},
		{"a/b", false, "slash rejected"},
		{"", false, "empty"},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%q", tc.id), func(t *testing.T) {
			t.Parallel()
			got := pathIDRegexp.MatchString(tc.id)
			if got != tc.ok {
				t.Fatalf("MatchString(%q) = %v, want %v (%s)", tc.id, got, tc.ok, tc.note)
			}
		})
	}
}

func TestIDValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id      string
		wantErr bool
	}{
		{"job1", false},
		{"my.job.id", false},
		{"a", false},
		{"", true},
		{strings.Repeat("a", 65), true},
		{"Abc", true},
		{"_ab", true},
	}

	v := IDValidator()
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%q", tc.id), func(t *testing.T) {
			t.Parallel()
			req := validator.StringRequest{ConfigValue: types.StringValue(tc.id)}
			var resp validator.StringResponse
			v.ValidateString(context.Background(), req, &resp)
			gotErr := resp.Diagnostics.HasError()
			if gotErr != tc.wantErr {
				t.Fatalf("ValidateString(%q) hasError=%v, want %v (%s)", tc.id, gotErr, tc.wantErr, resp.Diagnostics)
			}
		})
	}
}

func TestIDValidatorWithoutLength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id      string
		wantErr bool
	}{
		{"datafeed-opserv-riskviewxml-customer-transaction-volume-decline-stop", false},
		{strings.Repeat("a", 65), false},
		{"a", false},
		{"", true},
		{"Abc", true},
		{"_ab", true},
	}

	v := IDValidatorWithoutLength()
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%q", tc.id), func(t *testing.T) {
			t.Parallel()
			req := validator.StringRequest{ConfigValue: types.StringValue(tc.id)}
			var resp validator.StringResponse
			v.ValidateString(context.Background(), req, &resp)
			gotErr := resp.Diagnostics.HasError()
			if gotErr != tc.wantErr {
				t.Fatalf("ValidateString(%q) hasError=%v, want %v (%s)", tc.id, gotErr, tc.wantErr, resp.Diagnostics)
			}
		})
	}
}
