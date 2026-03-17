// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package integrationpolicy

import (
	"context"
	"fmt"

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
	return "integrationpolicy.InputsType"
}

// ValueType returns the Value type.
func (t InputsType) ValueType(_ context.Context) attr.Value {
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
func (t InputsType) ValueFromMap(_ context.Context, in basetypes.MapValue) (basetypes.MapValuable, diag.Diagnostics) {
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
		return nil, fmt.Errorf("unexpected type %T, expected basetypes.MapValue", attrValue)
	}

	return InputsValue{
		MapValue: mapValue,
	}, nil
}

// NewInputsType creates a new InputsType with the given element type
func NewInputsType(elemType InputType) InputsType {
	return InputsType{
		MapType: basetypes.MapType{
			ElemType: elemType,
		},
	}
}
