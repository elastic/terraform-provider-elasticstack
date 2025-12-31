package dashboard

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.ObjectValuable = (*IgnoreParentSettingsValue)(nil)
)

// IgnoreParentSettingsValue is a custom value type for the ignore_parent_settings nested attribute
type IgnoreParentSettingsValue struct {
	basetypes.ObjectValue
}

// Type returns an IgnoreParentSettingsType.
func (v IgnoreParentSettingsValue) Type(ctx context.Context) attr.Type {
	return NewIgnoreParentSettingsType()
}

// Equal returns true if the given value is equivalent.
func (v IgnoreParentSettingsValue) Equal(o attr.Value) bool {
	other, ok := o.(IgnoreParentSettingsValue)
	if !ok {
		return false
	}
	return v.ObjectValue.Equal(other.ObjectValue)
}

// ToModel converts the custom value to the internal model struct
func (v IgnoreParentSettingsValue) ToModel(ctx context.Context) (*ignoreParentSettingsModel, error) {
	if v.IsNull() || v.IsUnknown() {
		return nil, nil
	}

	var model ignoreParentSettingsModel
	diags := v.As(ctx, &model, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, fmt.Errorf("failed to convert to model: %v", diags)
	}

	return &model, nil
}

// NewIgnoreParentSettingsValueNull creates a null IgnoreParentSettingsValue
func NewIgnoreParentSettingsValueNull() IgnoreParentSettingsValue {
	return IgnoreParentSettingsValue{
		ObjectValue: types.ObjectNull(NewIgnoreParentSettingsType().AttrTypes),
	}
}

// NewIgnoreParentSettingsValueUnknown creates an unknown IgnoreParentSettingsValue
func NewIgnoreParentSettingsValueUnknown() IgnoreParentSettingsValue {
	return IgnoreParentSettingsValue{
		ObjectValue: types.ObjectUnknown(NewIgnoreParentSettingsType().AttrTypes),
	}
}

// NewIgnoreParentSettingsValue creates a new IgnoreParentSettingsValue from a model
func NewIgnoreParentSettingsValue(ctx context.Context, model *ignoreParentSettingsModel) (IgnoreParentSettingsValue, error) {
	if model == nil {
		return NewIgnoreParentSettingsValueNull(), nil
	}

	objValue, diags := types.ObjectValueFrom(ctx, NewIgnoreParentSettingsType().AttrTypes, model)
	if diags.HasError() {
		return IgnoreParentSettingsValue{}, fmt.Errorf("failed to create value: %v", diags)
	}

	return IgnoreParentSettingsValue{
		ObjectValue: objValue,
	}, nil
}
