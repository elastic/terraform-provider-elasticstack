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

package integrationpolicy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.StringValuable                   = (*VarsJSONValue)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*VarsJSONValue)(nil)
	_ xattr.ValidateableAttribute                = (*VarsJSONValue)(nil)
)

type VarsJSONValue struct {
	customtypes.JSONWithContextualDefaultsValue
}

// Type returns a VarsJSONType.
func (v VarsJSONValue) Type(ctx context.Context) attr.Type {
	return VarsJSONType{
		JSONWithContextualDefaultsType: v.JSONWithContextualDefaultsValue.Type(ctx).(customtypes.JSONWithContextualDefaultsType),
	}
}

// Equal returns true if the given value is equivalent.
func (v VarsJSONValue) Equal(o attr.Value) bool {
	other, ok := o.(VarsJSONValue)

	if !ok {
		return false
	}

	return v.JSONWithContextualDefaultsValue.Equal(other.JSONWithContextualDefaultsValue)
}

// StringSemanticEquals returns true if the given config object value is semantically equal to the current vars object value.
func (v VarsJSONValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	other, ok := newValuable.(VarsJSONValue)
	if !ok {
		var diags diag.Diagnostics
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)

		return false, diags
	}

	return v.JSONWithContextualDefaultsValue.StringSemanticEquals(ctx, other.JSONWithContextualDefaultsValue)
}

// NewVarsJSONNull creates a VarsJSONValue with a null value. Determine whether the value is null via IsNull method.
func NewVarsJSONNull() VarsJSONValue {
	return VarsJSONValue{
		JSONWithContextualDefaultsValue: customtypes.NewJSONWithContextualDefaultsNull(),
	}
}

// NewVarsJSONUnknown creates a VarsJSONValue with an unknown value. Determine whether the value is unknown via IsUnknown method.
func NewVarsJSONUnknown() VarsJSONValue {
	return VarsJSONValue{
		JSONWithContextualDefaultsValue: customtypes.NewJSONWithContextualDefaultsUnknown(),
	}
}

// NewVarsJSONWithIntegration creates a VarsJSONValue with a known value and a integration context. Access the value via ValueString method.
func NewVarsJSONWithIntegration(value string, name, version string) (VarsJSONValue, diag.Diagnostics) {
	integrationContext := getPackageCacheKey(name, version)
	jsonWithContext, diags := customtypes.NewJSONWithContextualDefaultsValue(value, integrationContext, populateVarsJSONDefaults)
	if diags.HasError() {
		return VarsJSONValue{}, diags
	}

	return VarsJSONValue{
		JSONWithContextualDefaultsValue: jsonWithContext,
	}, nil
}

func populateVarsJSONDefaults(ctxVal string, varsJSON string) (string, error) {
	if ctxVal == "" {
		return varsJSON, nil
	}

	value, ok := knownPackages.Load(ctxVal)
	if !ok {
		return varsJSON, nil
	}
	pkg, ok := value.(kbapi.PackageInfo)
	if !ok {
		return varsJSON, fmt.Errorf("unexpected package cache value type for key %q", ctxVal)
	}

	pkgVars, diags := varsFromPackageInfo(&pkg)
	if diags.HasError() {
		return varsJSON, diagutil.FwDiagsAsError(diags)
	}

	defaults, diags := pkgVars.defaults()
	if diags.HasError() {
		return varsJSON, diagutil.FwDiagsAsError(diags)
	}

	var vars map[string]any
	if err := json.Unmarshal([]byte(varsJSON), &vars); err != nil {
		return varsJSON, err
	}

	var defaultsMap map[string]any
	diags = defaults.Unmarshal(&defaultsMap)
	if diags.HasError() {
		return varsJSON, diagutil.FwDiagsAsError(diags)
	}

	for k, v := range defaultsMap {
		if _, ok := vars[k]; !ok {
			vars[k] = v
		}
	}

	varsBytes, err := json.Marshal(vars)
	if err != nil {
		return varsJSON, err
	}

	return string(varsBytes), nil
}
