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

package connector

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func configurationValueObject(t *testing.T, attrs map[string]attr.Value) types.Object {
	t.Helper()

	obj, diags := types.ObjectValue(configurationValueModelAttrTypes(), attrs)
	if diags.HasError() {
		t.Fatalf("building configuration value object: %v", diags)
	}
	return obj
}

func configurationValueNullFields() map[string]attr.Value {
	return map[string]attr.Value{
		"string":       types.StringNull(),
		"number":       types.NumberNull(),
		"bool":         types.BoolNull(),
		"json":         jsontypes.Normalized{StringValue: types.StringNull()},
		"secret_value": types.StringNull(),
	}
}

func TestConfigurationValueBranchValidator_rejectsNoBranchesSet(t *testing.T) {
	t.Parallel()

	var resp validator.ObjectResponse
	configurationValueBranchValidator{}.ValidateObject(context.Background(), validator.ObjectRequest{
		Path:        path.Root("configuration_values").AtMapKey("x"),
		ConfigValue: configurationValueObject(t, configurationValueNullFields()),
	}, &resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected validation error when no branch is set")
	}
}

func TestConfigurationValueBranchValidator_acceptsEachSingleBranch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		attrs map[string]attr.Value
	}{
		{
			name: "string",
			attrs: mergeConfigurationValueAttrs(configurationValueNullFields(), map[string]attr.Value{
				"string": types.StringValue("host"),
			}),
		},
		{
			name: "number",
			attrs: mergeConfigurationValueAttrs(configurationValueNullFields(), map[string]attr.Value{
				"number": types.NumberValue(big.NewFloat(5432)),
			}),
		},
		{
			name: "bool",
			attrs: mergeConfigurationValueAttrs(configurationValueNullFields(), map[string]attr.Value{
				"bool": types.BoolValue(true),
			}),
		},
		{
			name: "json",
			attrs: mergeConfigurationValueAttrs(configurationValueNullFields(), map[string]attr.Value{
				"json": jsontypes.Normalized{StringValue: types.StringValue(`{"k":"v"}`)},
			}),
		},
		{
			name: "secret_value",
			attrs: mergeConfigurationValueAttrs(configurationValueNullFields(), map[string]attr.Value{
				"secret_value": types.StringValue("pw"),
			}),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var resp validator.ObjectResponse
			configurationValueBranchValidator{}.ValidateObject(context.Background(), validator.ObjectRequest{
				Path:        path.Root("configuration_values").AtMapKey("x"),
				ConfigValue: configurationValueObject(t, tc.attrs),
			}, &resp)

			if resp.Diagnostics.HasError() {
				t.Fatalf("expected no error for branch %q, got %v", tc.name, resp.Diagnostics)
			}
		})
	}
}

func TestConfigurationValueBranchValidator_rejectsMultipleBranchesSet(t *testing.T) {
	t.Parallel()

	cases := []map[string]attr.Value{
		mergeConfigurationValueAttrs(configurationValueNullFields(), map[string]attr.Value{
			"string": types.StringValue("a"),
			"number": types.NumberValue(big.NewFloat(1)),
		}),
		mergeConfigurationValueAttrs(configurationValueNullFields(), map[string]attr.Value{
			"bool":         types.BoolValue(true),
			"secret_value": types.StringValue("pw"),
		}),
		mergeConfigurationValueAttrs(configurationValueNullFields(), map[string]attr.Value{
			"string": types.StringValue("a"),
			"bool":   types.BoolValue(false),
			"json":   jsontypes.Normalized{StringValue: types.StringValue(`[]`)},
		}),
	}

	for i, attrs := range cases {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			t.Parallel()

			var resp validator.ObjectResponse
			configurationValueBranchValidator{}.ValidateObject(context.Background(), validator.ObjectRequest{
				Path:        path.Root("configuration_values").AtMapKey("x"),
				ConfigValue: configurationValueObject(t, attrs),
			}, &resp)

			if !resp.Diagnostics.HasError() {
				t.Fatal("expected validation error when multiple branches are set")
			}
		})
	}
}

func TestConfigurationValueBranchValidator_acceptsOneBranchWithUnknowns(t *testing.T) {
	t.Parallel()

	attrs := mergeConfigurationValueAttrs(configurationValueNullFields(), map[string]attr.Value{
		"string":       types.StringValue("host"),
		"number":       types.NumberUnknown(),
		"bool":         types.BoolUnknown(),
		"json":         jsontypes.Normalized{StringValue: types.StringUnknown()},
		"secret_value": types.StringUnknown(),
	})

	var resp validator.ObjectResponse
	configurationValueBranchValidator{}.ValidateObject(context.Background(), validator.ObjectRequest{
		Path:        path.Root("configuration_values").AtMapKey("x"),
		ConfigValue: configurationValueObject(t, attrs),
	}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error when one branch is set and others are unknown, got %v", resp.Diagnostics)
	}
}

func mergeConfigurationValueAttrs(base, overrides map[string]attr.Value) map[string]attr.Value {
	out := make(map[string]attr.Value, len(base))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range overrides {
		out[k] = v
	}
	return out
}
