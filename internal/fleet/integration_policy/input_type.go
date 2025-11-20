package integration_policy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.ObjectTypable                    = (*InputType)(nil)
	_ basetypes.ObjectValuableWithSemanticEquals = (*InputValue)(nil)
)

// InputType is a custom type for an individual input that supports semantic equality
type InputType struct {
	basetypes.ObjectType
}

// String returns a human readable string of the type name.
func (t InputType) String() string {
	return "integration_policy.InputType"
}

// ValueType returns the Value type.
func (t InputType) ValueType(ctx context.Context) attr.Value {
	return InputValue{
		ObjectValue: basetypes.NewObjectUnknown(t.AttributeTypes()),
	}
}

// Equal returns true if the given type is equivalent.
func (t InputType) Equal(o attr.Type) bool {
	other, ok := o.(InputType)
	if !ok {
		return false
	}
	return t.ObjectType.Equal(other.ObjectType)
}

// ValueFromObject returns an ObjectValuable type given a basetypes.ObjectValue.
func (t InputType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	return InputValue{
		ObjectValue: in,
	}, nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.
func (t InputType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.ObjectType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	objectValue, ok := attrValue.(basetypes.ObjectValue)
	if !ok {
		return nil, err
	}

	return InputValue{
		ObjectValue: objectValue,
	}, nil
}

// NewInputType creates a new InputType with the given attribute types
func NewInputType(attrTypes map[string]attr.Type) InputType {
	return InputType{
		ObjectType: basetypes.ObjectType{
			AttrTypes: attrTypes,
		},
	}
}
