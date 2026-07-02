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
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMLDurationRegexp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		ok    bool
		note  string
	}{
		{"15m", true, "minutes"},
		{"1h", true, "hours"},
		{"150s", true, "seconds"},
		{"60s", true, "seconds short"},
		{"2m", true, "minutes short"},
		{"1d", true, "days"},
		{"2h", true, "hours"},
		{"100n", true, "nanos"},
		{"5u", true, "micros"},
		{"", false, "empty"},
		{"m", false, "no leading digits"},
		{"1.5m", false, "decimal not allowed"},
		{"1ms", false, "multi-char unit not allowed"},
		{"1H", false, "uppercase unit"},
		{"1 m", false, "space not allowed"},
		{"-1m", false, "negative"},
		{"0m", true, "zero value"},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%q", tc.input), func(t *testing.T) {
			t.Parallel()
			got := mlDurationRegexp.MatchString(tc.input)
			if got != tc.ok {
				t.Fatalf("MatchString(%q) = %v, want %v (%s)", tc.input, got, tc.ok, tc.note)
			}
		})
	}
}

func TestMLDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input   string
		wantErr bool
	}{
		{"15m", false},
		{"1h", false},
		{"150s", false},
		{"1d", false},
		{"100n", false},
		{"5u", false},
		{"", true},
		{"m", true},
		{"1.5m", true},
		{"1ms", true},
	}

	v := Duration()
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%q", tc.input), func(t *testing.T) {
			t.Parallel()
			req := validator.StringRequest{ConfigValue: types.StringValue(tc.input)}
			var resp validator.StringResponse
			v.ValidateString(context.Background(), req, &resp)
			gotErr := resp.Diagnostics.HasError()
			if gotErr != tc.wantErr {
				t.Fatalf("ValidateString(%q) hasError=%v, want %v (%s)", tc.input, gotErr, tc.wantErr, resp.Diagnostics)
			}
		})
	}
}
