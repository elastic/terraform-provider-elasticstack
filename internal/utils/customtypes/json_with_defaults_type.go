package customtypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable = (*JSONWithDefaultsType[any])(nil)
)

// PopulateDefaultsFunc is a function that takes a parsed model and returns it with defaults populated
type PopulateDefaultsFunc[TModel any] func(model TModel) TModel

// JSONWithDefaultsType is a generic type for JSON attributes that need default values populated
type JSONWithDefaultsType[TModel any] struct {
	jsontypes.NormalizedType
	populateDefaults PopulateDefaultsFunc[TModel]
}

// NewJSONWithDefaultsType creates a new JSONWithDefaultsType with the given PopulateDefaultsFunc
func NewJSONWithDefaultsType[TModel any](populateDefaults PopulateDefaultsFunc[TModel]) JSONWithDefaultsType[TModel] {
	return JSONWithDefaultsType[TModel]{
		NormalizedType:   jsontypes.NormalizedType{},
		populateDefaults: populateDefaults,
	}
}

// String returns a human readable string of the type name.
func (t JSONWithDefaultsType[TModel]) String() string {
	return "customtypes.JSONWithDefaultsType"
}

// ValueType returns the Value type.
func (t JSONWithDefaultsType[TModel]) ValueType(ctx context.Context) attr.Value {
	return JSONWithDefaultsValue[TModel]{
		populateDefaults: t.populateDefaults,
	}
}

// Equal returns true if the given type is equivalent.
func (t JSONWithDefaultsType[TModel]) Equal(o attr.Type) bool {
	other, ok := o.(JSONWithDefaultsType[TModel])

	if !ok {
		return false
	}

	return t.NormalizedType.Equal(other.NormalizedType)
}

// ValueFromString returns a StringValuable type given a StringValue.
func (t JSONWithDefaultsType[TModel]) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return JSONWithDefaultsValue[TModel]{
		Normalized:       jsontypes.Normalized{StringValue: in},
		populateDefaults: t.populateDefaults,
	}, nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.
func (t JSONWithDefaultsType[TModel]) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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
