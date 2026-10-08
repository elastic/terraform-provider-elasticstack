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

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable = (*NormalizedJSONType)(nil)
)

// NormalizedJSONType is the attr.Type companion for NormalizedJSONValue. It
// factors out the jsontypes.NormalizedType wrapper boilerplate (String,
// ValueType, Equal, ValueFromString, ValueFromTerraform) so that custom JSON
// string types which have no need for default-value population can embed it
// instead of re-deriving the same methods by hand, mirroring how
// JSONWithContextualDefaultsType is embedded by VarsJSONType.
type NormalizedJSONType struct {
	jsontypes.NormalizedType
}

// String returns a human readable string of the type name.
func (t NormalizedJSONType) String() string {
	return "customtypes.NormalizedJSONType"
}

// ValueType returns the Value type.
func (t NormalizedJSONType) ValueType(_ context.Context) attr.Value {
	return NormalizedJSONValue{}
}

// Equal returns true if the given type is equivalent.
func (t NormalizedJSONType) Equal(o attr.Type) bool {
	other, ok := o.(NormalizedJSONType)
	if !ok {
		return false
	}
	return t.NormalizedType.Equal(other.NormalizedType)
}

// ValueFromString returns a StringValuable type given a StringValue.
func (t NormalizedJSONType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return NormalizedJSONValue{StringValue: in}, nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.
func (t NormalizedJSONType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.NormalizedType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	norm, ok := attrValue.(jsontypes.Normalized)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	return NormalizedJSONValue{Normalized: norm}, nil
}
