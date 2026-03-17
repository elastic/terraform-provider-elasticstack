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

package apikey

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestRequiresTypeValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		typeValue   string
		attrValue   string
		expectError bool
	}

	testCases := []testCase{
		{
			name:        "role_descriptors with type=rest should be valid",
			typeValue:   "rest",
			attrValue:   `{"role": {"cluster": ["all"]}}`,
			expectError: false,
		},
		{
			name:        "role_descriptors with type=cross_cluster should be invalid",
			typeValue:   "cross_cluster",
			attrValue:   `{"role": {"cluster": ["all"]}}`,
			expectError: true,
		},
		{
			name:        "null role_descriptors with type=cross_cluster should be valid",
			typeValue:   "cross_cluster",
			attrValue:   "",
			expectError: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Create test config values
			configValues := map[string]tftypes.Value{
				"type": tftypes.NewValue(tftypes.String, testCase.typeValue),
			}

			if testCase.attrValue != "" {
				configValues["role_descriptors"] = tftypes.NewValue(tftypes.String, testCase.attrValue)
			} else {
				configValues["role_descriptors"] = tftypes.NewValue(tftypes.String, nil)
			}

			config := tfsdk.Config{
				Raw: tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"type":             tftypes.String,
					"role_descriptors": tftypes.String,
				}}, configValues),
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"type":             schema.StringAttribute{},
						"role_descriptors": schema.StringAttribute{},
					},
				},
			}

			var configValue types.String
			if testCase.attrValue != "" {
				configValue = types.StringValue(testCase.attrValue)
			} else {
				configValue = types.StringNull()
			}

			request := validator.StringRequest{
				Path:        path.Root("role_descriptors"),
				ConfigValue: configValue,
				Config:      config,
			}

			response := &validator.StringResponse{}
			requiresType("rest").ValidateString(context.Background(), request, response)

			if testCase.expectError && !response.Diagnostics.HasError() {
				t.Errorf("Expected error but got none")
			}

			if !testCase.expectError && response.Diagnostics.HasError() {
				t.Errorf("Expected no error but got: %v", response.Diagnostics)
			}
		})
	}
}
