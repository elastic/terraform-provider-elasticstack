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

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.StringValuable                   = (*NormalizedJSONValue)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*NormalizedJSONValue)(nil)
)

// NormalizedJSONValue is the attr.Value companion for NormalizedJSONType. It
// factors out the jsontypes.Normalized wrapper boilerplate (Type, Equal, and
// the Null/Unknown/Value constructors) shared by custom JSON string types
// that have no need for default-value population; their domain-specific
// StringSemanticEquals logic is defined on the embedding type.
type NormalizedJSONValue struct {
	jsontypes.Normalized
}

// Type returns a NormalizedJSONType.
func (v NormalizedJSONValue) Type(_ context.Context) attr.Type {
	return NormalizedJSONType{}
}

// Equal returns true if the given value is equivalent.
func (v NormalizedJSONValue) Equal(o attr.Value) bool {
	other, ok := o.(NormalizedJSONValue)
	if !ok {
		return false
	}
	return v.Normalized.Equal(other.Normalized)
}

// NewNormalizedJSONNull creates a NormalizedJSONValue with a null value.
func NewNormalizedJSONNull() NormalizedJSONValue {
	return NormalizedJSONValue{Normalized: jsontypes.NewNormalizedNull()}
}

// NewNormalizedJSONUnknown creates a NormalizedJSONValue with an unknown value.
func NewNormalizedJSONUnknown() NormalizedJSONValue {
	return NormalizedJSONValue{Normalized: jsontypes.NewNormalizedUnknown()}
}

// NewNormalizedJSONValue creates a NormalizedJSONValue with a known value.
func NewNormalizedJSONValue(value string) NormalizedJSONValue {
	return NormalizedJSONValue{Normalized: jsontypes.NewNormalizedValue(value)}
}
