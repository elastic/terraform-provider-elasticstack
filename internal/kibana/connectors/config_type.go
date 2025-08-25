package connectors

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable = (*ConfigType)(nil)
)

type ConfigType struct {
	jsontypes.NormalizedType
}

// String returns a human readable string of the type name.
func (t ConfigType) String() string {
	return "connectors.ConfigType"
}

// ValueType returns the Value type.
func (t ConfigType) ValueType(ctx context.Context) attr.Value {
	return ConfigValue{}
}

// Equal returns true if the given type is equivalent.
func (t ConfigType) Equal(o attr.Type) bool {
	other, ok := o.(ConfigType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

// ValueFromString returns a StringValuable type given a StringValue.
func (t ConfigType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	var connectorTypeID string
	if utils.IsKnown(in) {
		var configMap map[string]interface{}
		if err := json.Unmarshal([]byte(in.ValueString()), &configMap); err != nil {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic("Failed to unmarshal config value", err.Error()),
			}
		}

		connectorTypeID, _ = configMap[connectorTypeIDKey].(string)
	}

	return ConfigValue{
		Normalized: jsontypes.Normalized{
			StringValue: in,
		},
		connectorTypeID: connectorTypeID,
	}, nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.  This is meant to convert the tftypes.Value into a more convenient Go type
// for the provider to consume the data with.
func (t ConfigType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}
