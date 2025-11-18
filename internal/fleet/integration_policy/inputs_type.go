package integration_policy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.MapTypable                    = (*InputsType)(nil)
	_ basetypes.MapValuableWithSemanticEquals = (*InputsValue)(nil)
)

// InputsType is a custom type for the inputs map that supports semantic equality
type InputsType struct {
	basetypes.MapType
}

// String returns a human readable string of the type name.
func (t InputsType) String() string {
	return "integration_policy.InputsType"
}

// ValueType returns the Value type.
func (t InputsType) ValueType(ctx context.Context) attr.Value {
	return InputsValue{
		MapValue: basetypes.NewMapUnknown(t.ElementType()),
	}
}

// Equal returns true if the given type is equivalent.
func (t InputsType) Equal(o attr.Type) bool {
	other, ok := o.(InputsType)
	if !ok {
		return false
	}
	return t.MapType.Equal(other.MapType)
}

// ValueFromMap returns a MapValuable type given a basetypes.MapValue.
func (t InputsType) ValueFromMap(ctx context.Context, in basetypes.MapValue) (basetypes.MapValuable, diag.Diagnostics) {
	return InputsValue{
		MapValue: in,
	}, nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.
func (t InputsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.MapType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	mapValue, ok := attrValue.(basetypes.MapValue)
	if !ok {
		return nil, err
	}

	return InputsValue{
		MapValue: mapValue,
	}, nil
}

// NewInputsType creates a new InputsType with the given element type
func NewInputsType(elemType attr.Type) InputsType {
	return InputsType{
		MapType: basetypes.MapType{
			ElemType: elemType,
		},
	}
}
