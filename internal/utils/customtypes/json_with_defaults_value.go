package customtypes

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.StringValuable                   = (*JSONWithDefaultsValue[any])(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*JSONWithDefaultsValue[any])(nil)
	_ xattr.ValidateableAttribute                = (*JSONWithDefaultsValue[any])(nil)
)

// JSONWithDefaultsValue is a generic value type for JSON attributes that need default values populated
type JSONWithDefaultsValue[TModel any] struct {
	jsontypes.Normalized
	populateDefaults PopulateDefaultsFunc[TModel]
}

// Type returns a JSONWithDefaultsType.
func (v JSONWithDefaultsValue[TModel]) Type(_ context.Context) attr.Type {
	return JSONWithDefaultsType[TModel]{
		populateDefaults: v.populateDefaults,
	}
}

// WithDefaults applies default values to the JSON content
func (v JSONWithDefaultsValue[TModel]) WithDefaults() (JSONWithDefaultsValue[TModel], diag.Diagnostics) {
	var diags diag.Diagnostics

	if v.IsNull() {
		return v, diags
	}

	if v.IsUnknown() {
		return v, diags
	}

	if v.populateDefaults == nil {
		// If no populate defaults function is provided, return as-is
		return v, diags
	}

	var parsedValue TModel
	err := json.Unmarshal([]byte(v.ValueString()), &parsedValue)
	if err != nil {
		diags.AddError("Failed to unmarshal JSON value", err.Error())
		return JSONWithDefaultsValue[TModel]{}, diags
	}

	// Apply defaults
	populatedValue := v.populateDefaults(parsedValue)

	valueWithDefaults, err := json.Marshal(populatedValue)
	if err != nil {
		diags.AddError("Failed to marshal JSON value with defaults", err.Error())
		return JSONWithDefaultsValue[TModel]{}, diags
	}

	return NewJSONWithDefaultsValue(string(valueWithDefaults), v.populateDefaults), diags
}

// StringSemanticEquals returns true if the given value is semantically equal to the current value.
// The comparison will ignore any default values present in one value, but unset in the other.
func (v JSONWithDefaultsValue[TModel]) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(JSONWithDefaultsValue[TModel])
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

	thisWithDefaults, d := v.WithDefaults()
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	thatWithDefaults, d := newValue.WithDefaults()
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	return thisWithDefaults.Normalized.StringSemanticEquals(ctx, thatWithDefaults.Normalized)
}

// NewJSONWithDefaultsNull creates a JSONWithDefaultsValue with a null value.
func NewJSONWithDefaultsNull[TModel any](populateDefaults PopulateDefaultsFunc[TModel]) JSONWithDefaultsValue[TModel] {
	return JSONWithDefaultsValue[TModel]{
		Normalized:       jsontypes.NewNormalizedNull(),
		populateDefaults: populateDefaults,
	}
}

// NewJSONWithDefaultsUnknown creates a JSONWithDefaultsValue with an unknown value.
func NewJSONWithDefaultsUnknown[TModel any](populateDefaults PopulateDefaultsFunc[TModel]) JSONWithDefaultsValue[TModel] {
	return JSONWithDefaultsValue[TModel]{
		Normalized:       jsontypes.NewNormalizedUnknown(),
		populateDefaults: populateDefaults,
	}
}

// NewJSONWithDefaultsValue creates a JSONWithDefaultsValue with a known value.
func NewJSONWithDefaultsValue[TModel any](value string, populateDefaults PopulateDefaultsFunc[TModel]) JSONWithDefaultsValue[TModel] {
	if value == "" {
		return NewJSONWithDefaultsNull(populateDefaults)
	}

	return JSONWithDefaultsValue[TModel]{
		Normalized:       jsontypes.NewNormalizedValue(value),
		populateDefaults: populateDefaults,
	}
}
