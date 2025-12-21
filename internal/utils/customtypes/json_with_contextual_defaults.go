package customtypes

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringValuable                   = (*JSONWithContextualDefaultsValue)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*JSONWithContextualDefaultsValue)(nil)
	_ xattr.ValidateableAttribute                = (*JSONWithContextualDefaultsValue)(nil)
	_ basetypes.StringTypable                    = (*JSONWithContextualDefaultsType)(nil)
)

const legacyContextKey = "__tf_provider_connector_type_id"
const ContextKey = "__tf_provider_context"

type JSONWithContextualDefaultsValue struct {
	jsontypes.Normalized
	contextValue     string
	populateDefaults func(contextValue string, value string) (string, error)
}

type JSONWithContextualDefaultsType struct {
	jsontypes.NormalizedType
	populateDefaults func(contextValue string, value string) (string, error)
}

func NewJSONWithContextualDefaultsType(populateDefaults func(contextValue string, value string) (string, error)) JSONWithContextualDefaultsType {
	return JSONWithContextualDefaultsType{
		populateDefaults: populateDefaults,
	}
}

// String returns a human readable string of the type name.
func (t JSONWithContextualDefaultsType) String() string {
	return "customtypes.JSONWithContextType"
}

// ValueType returns the Value type.
func (t JSONWithContextualDefaultsType) ValueType(ctx context.Context) attr.Value {
	return JSONWithContextualDefaultsValue{
		populateDefaults: t.populateDefaults,
	}
}

// Equal returns true if the given type is equivalent.
func (t JSONWithContextualDefaultsType) Equal(o attr.Type) bool {
	other, ok := o.(JSONWithContextualDefaultsType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

// ValueFromString returns a StringValuable type given a StringValue.
func (t JSONWithContextualDefaultsType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	var contextValue string
	if utils.IsKnown(in) {
		var configMap map[string]interface{}
		if err := json.Unmarshal([]byte(in.ValueString()), &configMap); err != nil {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic("Failed to unmarshal config value", err.Error()),
			}
		}

		var ok bool
		contextValue, ok = configMap[ContextKey].(string)
		if !ok {
			contextValue, _ = configMap[legacyContextKey].(string)
		}
	}

	return JSONWithContextualDefaultsValue{
		Normalized: jsontypes.Normalized{
			StringValue: in,
		},
		contextValue:     contextValue,
		populateDefaults: t.populateDefaults,
	}, nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.
func (t JSONWithContextualDefaultsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

// Type returns a JSONWithContextType.
func (v JSONWithContextualDefaultsValue) Type(_ context.Context) attr.Type {
	return JSONWithContextualDefaultsType{
		populateDefaults: v.populateDefaults,
	}
}

// Equal returns true if the given value is equivalent.
func (v JSONWithContextualDefaultsValue) Equal(o attr.Value) bool {
	other, ok := o.(JSONWithContextualDefaultsValue)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (t JSONWithContextualDefaultsValue) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if t.IsNull() || t.IsUnknown() {
		return
	}

	t.Normalized.ValidateAttribute(ctx, req, resp)
}

func (v JSONWithContextualDefaultsValue) SanitizedValue() (string, diag.Diagnostics) {
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

	delete(unsanitizedMap, ContextKey)
	delete(unsanitizedMap, legacyContextKey)
	removeNulls(unsanitizedMap)
	sanitizedValue, err := json.Marshal(unsanitizedMap)
	if err != nil {
		diags.AddError("Failed to marshal sanitized config value", err.Error())
		return "", diags
	}

	return string(sanitizedValue), diags
}

// removeNulls recursively removes all null values from the map
func removeNulls(m map[string]interface{}) {
	for key, value := range m {
		if value == nil {
			delete(m, key)
			continue
		}

		if nestedMap, ok := value.(map[string]interface{}); ok {
			removeNulls(nestedMap)
			continue
		}
	}
}

// StringSemanticEquals returns true if the given config object value is semantically equal to the current config object value.
// The comparison will ignore any default values present in one value, but unset in the other.
func (v JSONWithContextualDefaultsValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(JSONWithContextualDefaultsValue)
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

	contextValue := v.contextValue
	if contextValue == "" {
		contextValue = newValue.contextValue
	}

	if contextValue == "" {
		// We cannot manage default values without a context value.
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

	populateDefaults := v.populateDefaults
	if populateDefaults == nil {
		populateDefaults = newValue.populateDefaults
	}

	if populateDefaults == nil {
		// Fallback to standard comparison if no populateDefaults function is available
		return v.Normalized.StringSemanticEquals(ctx, newValue.Normalized)
	}

	thisWithDefaults, err := populateDefaults(contextValue, thisString)
	if err != nil {
		diags.AddError("Failed to get config with defaults", err.Error())
	}
	thatWithDefaults, err := populateDefaults(contextValue, thatString)
	if err != nil {
		diags.AddError("Failed to get config with defaults", err.Error())
	}

	normalizedWithDefaults := jsontypes.NewNormalizedValue(thisWithDefaults)
	normalizedThatWithDefaults := jsontypes.NewNormalizedValue(thatWithDefaults)
	return normalizedWithDefaults.StringSemanticEquals(ctx, normalizedThatWithDefaults)
}

// NewJSONWithContextualDefaultsNull creates a JSONWithContextualDefaults with a null value.
func NewJSONWithContextualDefaultsNull() JSONWithContextualDefaultsValue {
	return JSONWithContextualDefaultsValue{
		Normalized: jsontypes.NewNormalizedNull(),
	}
}

// NewJSONWithContextualDefaultsUnknown creates a JSONWithContextualDefaults with an unknown value.
func NewJSONWithContextualDefaultsUnknown() JSONWithContextualDefaultsValue {
	return JSONWithContextualDefaultsValue{
		Normalized: jsontypes.NewNormalizedUnknown(),
	}
}

// NewJSONWithContextualDefaultsValue creates a JSONWithContext with a known value and a context value.
func NewJSONWithContextualDefaultsValue(value string, contextValue string, populateDefaults func(contextValue string, value string) (string, error)) (JSONWithContextualDefaultsValue, diag.Diagnostics) {
	if value == "" {
		return NewJSONWithContextualDefaultsNull(), nil
	}

	var configMap map[string]interface{}
	err := json.Unmarshal([]byte(value), &configMap)
	if err != nil {
		return JSONWithContextualDefaultsValue{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Failed to unmarshal config", err.Error()),
		}
	}

	configMap[ContextKey] = contextValue
	jsonBytes, err := json.Marshal(configMap)
	if err != nil {
		return JSONWithContextualDefaultsValue{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Failed to marshal config", err.Error()),
		}
	}

	return JSONWithContextualDefaultsValue{
		Normalized:       jsontypes.NewNormalizedValue(string(jsonBytes)),
		contextValue:     contextValue,
		populateDefaults: populateDefaults,
	}, nil
}
