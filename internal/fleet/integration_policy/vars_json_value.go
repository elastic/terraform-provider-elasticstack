package integration_policy

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
func (v VarsJSONValue) Type(_ context.Context) attr.Type {
	return VarsJSONType{
		JSONWithContextualDefaultsType: v.JSONWithContextualDefaultsValue.Type(context.Background()).(customtypes.JSONWithContextualDefaultsType),
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

func populateVarsJSONDefaults(ctxVal string, varsJson string) (string, error) {
	if ctxVal == "" {
		return varsJson, nil
	}

	value, ok := knownPackages.Load(ctxVal)
	if !ok {
		return varsJson, nil
	}
	pkg, ok := value.(kbapi.PackageInfo)
	if !ok {
		return varsJson, fmt.Errorf("unexpected package cache value type for key %q", ctxVal)
	}

	pkgVars, diags := varsFromPackageInfo(&pkg)
	if diags.HasError() {
		return varsJson, diagutil.FwDiagsAsError(diags)
	}

	defaults, diags := pkgVars.defaults()
	if diags.HasError() {
		return varsJson, diagutil.FwDiagsAsError(diags)
	}

	var vars map[string]interface{}
	if err := json.Unmarshal([]byte(varsJson), &vars); err != nil {
		return varsJson, err
	}

	var defaultsMap map[string]interface{}
	diags = defaults.Unmarshal(&defaultsMap)
	if diags.HasError() {
		return varsJson, diagutil.FwDiagsAsError(diags)
	}

	for k, v := range defaultsMap {
		if _, ok := vars[k]; !ok {
			vars[k] = v
		}
	}

	varsBytes, err := json.Marshal(vars)
	if err != nil {
		return varsJson, err
	}

	return string(varsBytes), nil
}
