package integration_policy

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable = (*VarsJSONType)(nil)
)

type VarsJSONType struct {
	customtypes.JSONWithContextualDefaultsType
}

// String returns a human readable string of the type name.
func (t VarsJSONType) String() string {
	return "integration_policy.VarsJSONType"
}

// ValueType returns the Value type.
func (t VarsJSONType) ValueType(ctx context.Context) attr.Value {
	return VarsJSONValue{
		JSONWithContextualDefaultsValue: t.JSONWithContextualDefaultsType.ValueType(ctx).(customtypes.JSONWithContextualDefaultsValue),
	}
}

// Equal returns true if the given type is equivalent.
func (t VarsJSONType) Equal(o attr.Type) bool {
	other, ok := o.(VarsJSONType)

	if !ok {
		return false
	}

	return t.JSONWithContextualDefaultsType.Equal(other.JSONWithContextualDefaultsType)
}

// ValueFromString returns a StringValuable type given a StringValue.
func (t VarsJSONType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	val, diags := t.JSONWithContextualDefaultsType.ValueFromString(ctx, in)
	if diags.HasError() {
		return nil, diags
	}

	return VarsJSONValue{
		JSONWithContextualDefaultsValue: val.(customtypes.JSONWithContextualDefaultsValue),
	}, nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.  This is meant to convert the tftypes.Value into a more convenient Go type
// for the provider to consume the data with.
func (t VarsJSONType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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
