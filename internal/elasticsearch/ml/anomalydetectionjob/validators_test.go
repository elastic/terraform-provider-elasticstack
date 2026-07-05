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

package anomalydetectionjob

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestResultsIndexNameWithoutCustomPrefixValidator_ValidateString(t *testing.T) {
	t.Parallel()

	v := resultsIndexNameWithoutCustomPrefix()

	tests := []struct {
		name      string
		value     types.String
		wantError bool
	}{
		{
			name:  "null value skipped",
			value: types.StringNull(),
		},
		{
			name:  "unknown value skipped",
			value: types.StringUnknown(),
		},
		{
			name:  "shared allowed",
			value: types.StringValue("shared"),
		},
		{
			name:  "user suffix allowed",
			value: types.StringValue("ml-logs-error-count"),
		},
		{
			name:      "custom prefix rejected",
			value:     types.StringValue("custom-ml-logs-error-count"),
			wantError: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), validator.StringRequest{
				Path:        path.Root("results_index_name"),
				ConfigValue: tc.value,
			}, resp)

			hasError := resp.Diagnostics.HasError()
			if hasError != tc.wantError {
				t.Fatalf("ValidateString() hasError = %v, wantError = %v; diagnostics: %v", hasError, tc.wantError, resp.Diagnostics)
			}
		})
	}
}
