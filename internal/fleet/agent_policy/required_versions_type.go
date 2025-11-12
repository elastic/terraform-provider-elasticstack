package agent_policy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.SetTypable = (*RequiredVersionsType)(nil)
)

// RequiredVersionsType is a custom set type that enforces uniqueness based on version only.
type RequiredVersionsType struct {
	basetypes.SetType
}

// String returns a human readable string of the type name.
func (t RequiredVersionsType) String() string {
	return "agent_policy.RequiredVersionsType"
}

// ValueType returns the Value type.
func (t RequiredVersionsType) ValueType(ctx context.Context) attr.Value {
	return RequiredVersionsValue{
		SetValue: basetypes.NewSetUnknown(t.ElemType),
	}
}

// Equal returns true if the given type is equivalent.
func (t RequiredVersionsType) Equal(o attr.Type) bool {
	other, ok := o.(RequiredVersionsType)
	if !ok {
		return false
	}
	return t.SetType.Equal(other.SetType)
}

// ValueFromSet returns a SetValuable type given a SetValue.
func (t RequiredVersionsType) ValueFromSet(ctx context.Context, in basetypes.SetValue) (basetypes.SetValuable, diag.Diagnostics) {
	return RequiredVersionsValue{
		SetValue: in,
	}, nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.
func (t RequiredVersionsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.SetType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	setValue, ok := attrValue.(basetypes.SetValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	setValuable, diags := t.ValueFromSet(ctx, setValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting SetValue to SetValuable: %v", diags)
	}

	return setValuable, nil
}
