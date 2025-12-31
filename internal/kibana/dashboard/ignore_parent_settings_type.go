package dashboard

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.ObjectTypable = (*IgnoreParentSettingsType)(nil)
)

// IgnoreParentSettingsType is a custom type for the ignore_parent_settings nested attribute
type IgnoreParentSettingsType struct {
	basetypes.ObjectType
}

// String returns a human readable string of the type name.
func (t IgnoreParentSettingsType) String() string {
	return "dashboard.IgnoreParentSettingsType"
}

// ValueType returns the Value type.
func (t IgnoreParentSettingsType) ValueType(ctx context.Context) attr.Value {
	return IgnoreParentSettingsValue{
		ObjectValue: types.ObjectNull(t.AttrTypes),
	}
}

// Equal returns true if the given type is equivalent.
func (t IgnoreParentSettingsType) Equal(o attr.Type) bool {
	other, ok := o.(IgnoreParentSettingsType)
	if !ok {
		return false
	}
	return t.ObjectType.Equal(other.ObjectType)
}

// ValueFromObject returns an IgnoreParentSettingsValue given an ObjectValue.
func (t IgnoreParentSettingsType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	return IgnoreParentSettingsValue{
		ObjectValue: in,
	}, nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.
func (t IgnoreParentSettingsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.ObjectType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	objectValue, ok := attrValue.(basetypes.ObjectValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	objectValuable, diags := t.ValueFromObject(ctx, objectValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting ObjectValue to ObjectValuable: %v", diags)
	}

	return objectValuable, nil
}

// NewIgnoreParentSettingsType creates a new IgnoreParentSettingsType with the appropriate attribute types
func NewIgnoreParentSettingsType() IgnoreParentSettingsType {
	return IgnoreParentSettingsType{
		ObjectType: types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"ignore_filters":     types.BoolType,
				"ignore_query":       types.BoolType,
				"ignore_timerange":   types.BoolType,
				"ignore_validations": types.BoolType,
			},
		},
	}
}
