package connectors

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.StringValuable                   = (*ConfigValue)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*ConfigValue)(nil)
	_ xattr.ValidateableAttribute                = (*ConfigValue)(nil)
)

type ConfigValue struct {
	customtypes.JSONWithContextualDefaultsValue
}

// Type returns a ConfigType.
func (v ConfigValue) Type(_ context.Context) attr.Type {
	return ConfigType{
		JSONWithContextualDefaultsType: v.JSONWithContextualDefaultsValue.Type(context.Background()).(customtypes.JSONWithContextualDefaultsType),
	}
}

// Equal returns true if the given value is equivalent.
func (v ConfigValue) Equal(o attr.Value) bool {
	other, ok := o.(ConfigValue)

	if !ok {
		return false
	}

	return v.JSONWithContextualDefaultsValue.Equal(other.JSONWithContextualDefaultsValue)
}

// StringSemanticEquals returns true if the given config object value is semantically equal to the current config object value.
func (v ConfigValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	other, ok := newValuable.(ConfigValue)
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

// NewConfigNull creates a ConfigValue with a null value. Determine whether the value is null via IsNull method.
func NewConfigNull() ConfigValue {
	return ConfigValue{
		JSONWithContextualDefaultsValue: customtypes.NewJSONWithContextualDefaultsNull(),
	}
}

// NewConfigUnknown creates a ConfigValue with an unknown value. Determine whether the value is unknown via IsUnknown method.
func NewConfigUnknown() ConfigValue {
	return ConfigValue{
		JSONWithContextualDefaultsValue: customtypes.NewJSONWithContextualDefaultsUnknown(),
	}
}

// NewConfigValueWithConnectorID creates a ConfigValue with a known value and a connector type ID. Access the value via ValueString method.
func NewConfigValueWithConnectorID(value string, connectorTypeID string) (ConfigValue, diag.Diagnostics) {
	jsonWithContext, diags := customtypes.NewJSONWithContextualDefaultsValue(value, connectorTypeID, kibana_oapi.ConnectorConfigWithDefaults)
	if diags.HasError() {
		return ConfigValue{}, diags
	}

	return ConfigValue{
		JSONWithContextualDefaultsValue: jsonWithContext,
	}, nil
}
