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

package customtypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable = NormalizedYamlType{}
)

// NormalizedYamlType is a custom type for YAML attributes that performs
// semantic equality comparison, ignoring insignificant whitespace and
// key ordering differences.
type NormalizedYamlType struct {
	basetypes.StringType
}

// String returns a human readable string of the type name.
func (t NormalizedYamlType) String() string {
	return "customtypes.NormalizedYamlType"
}

// ValueType returns the Value type.
func (t NormalizedYamlType) ValueType(_ context.Context) attr.Value {
	return NormalizedYamlValue{}
}

// Equal returns true if the given type is equivalent.
func (t NormalizedYamlType) Equal(o attr.Type) bool {
	_, ok := o.(NormalizedYamlType)
	return ok
}

// ValueFromString returns a StringValuable type given a StringValue.
func (t NormalizedYamlType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return NormalizedYamlValue{StringValue: in}, nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.
func (t NormalizedYamlType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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
