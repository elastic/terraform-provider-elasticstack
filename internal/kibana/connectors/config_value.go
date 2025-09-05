package connectors

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
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
	jsontypes.Normalized
	connectorTypeID string
}

// Type returns a ConfigType.
func (v ConfigValue) Type(_ context.Context) attr.Type {
	return ConfigType{}
}

// Equal returns true if the given value is equivalent.
func (v ConfigValue) Equal(o attr.Value) bool {
	other, ok := o.(ConfigValue)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (t ConfigValue) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if t.IsNull() || t.IsUnknown() {
		return
	}

	t.Normalized.ValidateAttribute(ctx, req, resp)
}

func (v ConfigValue) SanitizedValue() (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	if v.IsNull() {
		return "", diags
	}

	if v.IsUnknown() {
		return "", diags
	}

	var unsanitizedMap map[string]interface{}
	err := json.Unmarshal([]byte(v.ValueString()), &unsanitizedMap)
	if err != nil {
		diags.AddError("Failed to unmarshal config value", err.Error())
		return "", diags
	}

	delete(unsanitizedMap, connectorTypeIDKey)
	sanitizedValue, err := json.Marshal(unsanitizedMap)
	if err != nil {
		diags.AddError("Failed to marshal sanitized config value", err.Error())
		return "", diags
	}

	return string(sanitizedValue), diags
}

// StringSemanticEquals returns true if the given config object value is semantically equal to the current config object value.
// The comparison will ignore any default values present in one value, but unset in the other.
func (v ConfigValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(ConfigValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)

		return false, diags
	}

	if v.IsNull() {
		return newValue.IsNull(), diags
	}

	if v.IsUnknown() {
		return newValue.IsUnknown(), diags
	}

	connectorTypeID := v.connectorTypeID
	if connectorTypeID == "" {
		connectorTypeID = newValue.connectorTypeID
	}

	if connectorTypeID == "" {
		// We cannot manage default values without a connector type ID.
		return v.Normalized.StringSemanticEquals(ctx, newValue.Normalized)
	}

	thisString, diags := v.SanitizedValue()
	if diags.HasError() {
		return false, diags
	}
	thatString, diags := newValue.SanitizedValue()
	if diags.HasError() {
		return false, diags
	}

	thisWithDefaults, err := kibana_oapi.ConnectorConfigWithDefaults(connectorTypeID, thisString)
	if err != nil {
		diags.AddError("Failed to get connector config with defaults", err.Error())
	}
	thatWithDefaults, err := kibana_oapi.ConnectorConfigWithDefaults(connectorTypeID, thatString)
	if err != nil {
		diags.AddError("Failed to get connector config with defaults", err.Error())
	}

	normalizedWithDefaults := jsontypes.NewNormalizedValue(thisWithDefaults)
	normalizedThatWithDefaults := jsontypes.NewNormalizedValue(thatWithDefaults)
	return normalizedWithDefaults.StringSemanticEquals(ctx, normalizedThatWithDefaults)
}

// NewConfigNull creates a ConfigValue with a null value. Determine whether the value is null via IsNull method.
func NewConfigNull() ConfigValue {
	return ConfigValue{
		Normalized: jsontypes.NewNormalizedNull(),
	}
}

// NewConfigUnknown creates a ConfigValue with an unknown value. Determine whether the value is unknown via IsUnknown method.
func NewConfigUnknown() ConfigValue {
	return ConfigValue{
		Normalized: jsontypes.NewNormalizedUnknown(),
	}
}

const connectorTypeIDKey = "__tf_provider_connector_type_id"

// NewConfigValueWithConnectorID creates a ConfigValue with a known value and a connector type ID. Access the value via ValueString method.
func NewConfigValueWithConnectorID(value string, connectorTypeID string) (ConfigValue, diag.Diagnostics) {
	if value == "" {
		return NewConfigNull(), nil
	}

	var configMap map[string]interface{}
	err := json.Unmarshal([]byte(value), &configMap)
	if err != nil {
		return ConfigValue{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Failed to unmarshal config", err.Error()),
		}
	}

	configMap[connectorTypeIDKey] = connectorTypeID
	jsonBytes, err := json.Marshal(configMap)
	if err != nil {
		return ConfigValue{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Failed to marshal config", err.Error()),
		}
	}

	return ConfigValue{
		Normalized:      jsontypes.NewNormalizedValue(string(jsonBytes)),
		connectorTypeID: connectorTypeID,
	}, nil
}
