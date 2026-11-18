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

func TestMLDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input   string
		wantErr bool
	}{
		{"15m", false},
		{"1h", false},
		{"150s", false},
		{"60s", false},
		{"2m", false},
		{"2h", false},
		{"1d", false},
		{"100n", false},
		{"5u", false},
		{"0m", false},
		{"", true},
		{"m", true},
		{"1.5m", true},
		{"1ms", true},
		{"1H", true},
		{"1 m", true},
		{"-1m", true},
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
