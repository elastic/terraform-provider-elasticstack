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

package privatelocation

import (
	"context"
	"fmt"
	"math"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Compile-time interface assertions.
var (
	_ basetypes.Float64Typable                    = Float32PrecisionType{}
	_ basetypes.Float64Valuable                   = Float32PrecisionValue{}
	_ basetypes.Float64ValuableWithSemanticEquals = Float32PrecisionValue{}
)

// Float32PrecisionType is a custom Float64 type whose values compare semantically
// equal when they differ only due to float32 storage precision. The Kibana API
// stores geo coordinates as float32; on read they are returned as the float64
// representation of that float32 (e.g. 42.42 → 42.41999816894531). Declaring
// this type for lat/lon tells Terraform to treat such differences as a no-op.
type Float32PrecisionType struct {
	basetypes.Float64Type
}

func (t Float32PrecisionType) String() string {
	return "privatelocation.Float32PrecisionType"
}

func (t Float32PrecisionType) ValueType(_ context.Context) attr.Value {
	return Float32PrecisionValue{}
}

func (t Float32PrecisionType) Equal(o attr.Type) bool {
	_, ok := o.(Float32PrecisionType)
	return ok
}

func (t Float32PrecisionType) ValueFromFloat64(_ context.Context, in basetypes.Float64Value) (basetypes.Float64Valuable, diag.Diagnostics) {
	return Float32PrecisionValue{Float64Value: in}, nil
}

func (t Float32PrecisionType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.Float64Type.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	float64Value, ok := attrValue.(basetypes.Float64Value)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	float64Valuable, diags := t.ValueFromFloat64(ctx, float64Value)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting Float64Value to Float64Valuable: %v", diags)
	}

	return float64Valuable, nil
}

// Float32PrecisionValue is a custom Float64 value that implements semantic equality
// using float32 precision. Two values are considered semantically equal when their
// float32 representations are identical.
type Float32PrecisionValue struct {
	basetypes.Float64Value
}

func NewFloat32PrecisionValue(v float64) Float32PrecisionValue {
	return Float32PrecisionValue{Float64Value: basetypes.NewFloat64Value(v)}
}

func NewFloat32PrecisionNull() Float32PrecisionValue {
	return Float32PrecisionValue{Float64Value: basetypes.NewFloat64Null()}
}

func NewFloat32PrecisionUnknown() Float32PrecisionValue {
	return Float32PrecisionValue{Float64Value: basetypes.NewFloat64Unknown()}
}

func (v Float32PrecisionValue) Type(_ context.Context) attr.Type {
	return Float32PrecisionType{}
}

// Float64SemanticEquals returns true when both values are equivalent under float32
// precision. This prevents spurious plan diffs when the Kibana API returns a
// float32-degraded representation of the user-supplied float64 value.
func (v Float32PrecisionValue) Float64SemanticEquals(_ context.Context, newValuable basetypes.Float64Valuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(Float32PrecisionValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)
		return false, diags
	}

	if v.IsNull() {
		return newValue.IsNull(), diags
	}

	if v.IsUnknown() {
		return newValue.IsUnknown(), diags
	}

	a := v.ValueFloat64()
	b := newValue.ValueFloat64()

	if math.IsNaN(a) || math.IsNaN(b) {
		return false, diags
	}

	return float32(a) == float32(b), diags
}
