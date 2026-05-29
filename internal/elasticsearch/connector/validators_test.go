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
	"maps"
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
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
		stringBranchAttrName:      types.StringNull(),
		numberBranchAttrName:      types.NumberNull(),
		boolBranchAttrName:        types.BoolNull(),
		jsonBranchAttrName:        jsontypes.Normalized{StringValue: types.StringNull()},
		secretValueBranchAttrName: types.StringNull(),
	}
}

func TestConfigurationValueBranchValidator_rejectsNoBranchesSet(t *testing.T) {
	t.Parallel()

	wantPath := path.Root("configuration_values").AtMapKey("x")

	var resp validator.ObjectResponse
	configurationValueBranchValidator{}.ValidateObject(context.Background(), validator.ObjectRequest{
		Path:        wantPath,
		ConfigValue: configurationValueObject(t, configurationValueNullFields()),
	}, &resp)

	require.True(t, resp.Diagnostics.HasError(), "expected validation error when no branch is set")
	assertConfigurationValueBranchDiagnostic(t, resp.Diagnostics[0], wantPath, configurationValueBranchErrorSummary, "")
}

func TestConfigurationValueBranchValidator_rejectsNoBranchesSet_diagnosticContent(t *testing.T) {
	t.Parallel()

	wantPath := path.Root("configuration_values").AtMapKey("x")

	var resp validator.ObjectResponse
	configurationValueBranchValidator{}.ValidateObject(context.Background(), validator.ObjectRequest{
		Path:        wantPath,
		ConfigValue: configurationValueObject(t, configurationValueNullFields()),
	}, &resp)

	require.Len(t, resp.Diagnostics, 1)
	assertConfigurationValueBranchDiagnostic(t, resp.Diagnostics[0], wantPath, configurationValueBranchErrorSummary, "must be set")
}

func TestConfigurationValueBranchValidator_rejectsMultipleBranchesSet_diagnosticContent(t *testing.T) {
	t.Parallel()

	wantPath := path.Root("configuration_values").AtMapKey("x")
	attrs := mergeConfigurationValueAttrs(configurationValueNullFields(), map[string]attr.Value{
		stringBranchAttrName: types.StringValue("a"),
		numberBranchAttrName: types.NumberValue(big.NewFloat(1)),
	})

	var resp validator.ObjectResponse
	configurationValueBranchValidator{}.ValidateObject(context.Background(), validator.ObjectRequest{
		Path:        wantPath,
		ConfigValue: configurationValueObject(t, attrs),
	}, &resp)

	require.True(t, resp.Diagnostics.HasError())
	assertConfigurationValueBranchDiagnostic(t, resp.Diagnostics[0], wantPath, configurationValueBranchErrorSummary, "found 2 set")
	detail := resp.Diagnostics[0].Detail()
	require.Contains(t, detail, stringBranchAttrName)
	require.Contains(t, detail, numberBranchAttrName)
}

func TestConfigurationValueBranchValidator_skipsNullOrUnknownObject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value types.Object
	}{
		{name: "null object", value: types.ObjectNull(configurationValueModelAttrTypes())},
		{name: "unknown object", value: types.ObjectUnknown(configurationValueModelAttrTypes())},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var resp validator.ObjectResponse
			configurationValueBranchValidator{}.ValidateObject(context.Background(), validator.ObjectRequest{
				Path:        path.Root("configuration_values").AtMapKey("x"),
				ConfigValue: tc.value,
			}, &resp)

			require.False(t, resp.Diagnostics.HasError(), "diags: %v", resp.Diagnostics)
		})
	}
}

func TestConfigurationValueBranchValidator_skipsAllBranchesUnknown(t *testing.T) {
	t.Parallel()

	attrs := map[string]attr.Value{
		stringBranchAttrName:      types.StringUnknown(),
		numberBranchAttrName:      types.NumberUnknown(),
		boolBranchAttrName:        types.BoolUnknown(),
		jsonBranchAttrName:        jsontypes.Normalized{StringValue: types.StringUnknown()},
		secretValueBranchAttrName: types.StringUnknown(),
	}

	var resp validator.ObjectResponse
	configurationValueBranchValidator{}.ValidateObject(context.Background(), validator.ObjectRequest{
		Path:        path.Root("configuration_values").AtMapKey("x"),
		ConfigValue: configurationValueObject(t, attrs),
	}, &resp)

	require.False(t, resp.Diagnostics.HasError(), "diags: %v", resp.Diagnostics)
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
				stringBranchAttrName: types.StringValue("host"),
			}),
		},
		{
			name: "number",
			attrs: mergeConfigurationValueAttrs(configurationValueNullFields(), map[string]attr.Value{
				numberBranchAttrName: types.NumberValue(big.NewFloat(5432)),
			}),
		},
		{
			name: "bool",
			attrs: mergeConfigurationValueAttrs(configurationValueNullFields(), map[string]attr.Value{
				boolBranchAttrName: types.BoolValue(true),
			}),
		},
		{
			name: "json",
			attrs: mergeConfigurationValueAttrs(configurationValueNullFields(), map[string]attr.Value{
				jsonBranchAttrName: jsontypes.Normalized{StringValue: types.StringValue(`{"k":"v"}`)},
			}),
		},
		{
			name: "secret_value",
			attrs: mergeConfigurationValueAttrs(configurationValueNullFields(), map[string]attr.Value{
				secretValueBranchAttrName: types.StringValue("pw"),
			}),
		},
		{
			name: "secret_value unknown during plan",
			attrs: mergeConfigurationValueAttrs(configurationValueNullFields(), map[string]attr.Value{
				secretValueBranchAttrName: types.StringUnknown(),
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
			stringBranchAttrName: types.StringValue("a"),
			numberBranchAttrName: types.NumberValue(big.NewFloat(1)),
		}),
		mergeConfigurationValueAttrs(configurationValueNullFields(), map[string]attr.Value{
			boolBranchAttrName:        types.BoolValue(true),
			secretValueBranchAttrName: types.StringValue("pw"),
		}),
		mergeConfigurationValueAttrs(configurationValueNullFields(), map[string]attr.Value{
			stringBranchAttrName: types.StringValue("a"),
			boolBranchAttrName:   types.BoolValue(false),
			jsonBranchAttrName:   jsontypes.Normalized{StringValue: types.StringValue(`[]`)},
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

func TestConfigurationValueBranchValidator_rejectsTwoUnknownBranches(t *testing.T) {
	t.Parallel()

	attrs := map[string]attr.Value{
		stringBranchAttrName:      types.StringUnknown(),
		numberBranchAttrName:      types.NumberUnknown(),
		boolBranchAttrName:        types.BoolNull(),
		jsonBranchAttrName:        jsontypes.Normalized{StringValue: types.StringNull()},
		secretValueBranchAttrName: types.StringNull(),
	}

	var resp validator.ObjectResponse
	configurationValueBranchValidator{}.ValidateObject(context.Background(), validator.ObjectRequest{
		Path:        path.Root("configuration_values").AtMapKey("x"),
		ConfigValue: configurationValueObject(t, attrs),
	}, &resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected validation error when two branches are unknown")
	}
}

func TestConfigurationValueBranchValidator_acceptsOneBranchWithUnknowns(t *testing.T) {
	t.Parallel()

	attrs := mergeConfigurationValueAttrs(configurationValueNullFields(), map[string]attr.Value{
		stringBranchAttrName:      types.StringValue("host"),
		numberBranchAttrName:      types.NumberUnknown(),
		boolBranchAttrName:        types.BoolUnknown(),
		jsonBranchAttrName:        jsontypes.Normalized{StringValue: types.StringUnknown()},
		secretValueBranchAttrName: types.StringUnknown(),
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

func assertConfigurationValueBranchDiagnostic(
	t *testing.T,
	diagnostic diag.Diagnostic,
	wantPath path.Path,
	wantSummary string,
	wantDetailSubstring string,
) {
	t.Helper()

	dwp, ok := diagnostic.(diag.DiagnosticWithPath)
	require.True(t, ok, "expected attribute diagnostic with Path()")
	require.Equal(t, wantPath, dwp.Path())
	require.Equal(t, wantSummary, diagnostic.Summary())
	if wantDetailSubstring != "" {
		require.Contains(t, diagnostic.Detail(), wantDetailSubstring)
	}
}

func mergeConfigurationValueAttrs(base, overrides map[string]attr.Value) map[string]attr.Value {
	out := make(map[string]attr.Value, len(base))
	maps.Copy(out, base)
	maps.Copy(out, overrides)
	return out
}
