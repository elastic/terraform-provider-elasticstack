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
func (t InputType) ValueType(_ context.Context) attr.Value {
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
func (t InputType) ValueFromObject(_ context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
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
