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

package typeutils

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MapToNormalizedType marshals a map[string]T into a jsontypes.Normalized.
func MapToNormalizedType[T any](value map[string]T, p path.Path, diags *diag.Diagnostics) jsontypes.Normalized {
	if value == nil {
		return jsontypes.NewNormalizedNull()
	}

	bytes, err := json.Marshal(value)
	if err != nil {
		diags.AddAttributeError(p, "marshal failure", err.Error())
		return jsontypes.NewNormalizedNull()
	}

	return jsontypes.NewNormalizedValue(string(bytes))
}

// NormalizedTypeToMap unmarshals a jsontypes.Normalized to a map[string]T.
func NormalizedTypeToMap[T any](value jsontypes.Normalized, p path.Path, diags *diag.Diagnostics) map[string]T {
	if !IsKnown(value) {
		return nil
	}

	var dest map[string]T
	d := value.Unmarshal(&dest)
	diags.Append(convertToAttrDiags(d, p)...)
	return dest
}

// MapToMapType converts a tfsdk naive map[string]T1 into an types.Map of map[string]T2.
// This handles both structs and simple types to attr.Values.
func MapToMapType[T1 any, T2 any](ctx context.Context, value map[string]T1, elemType attr.Type, p path.Path, diags *diag.Diagnostics, iteratee func(item T1, meta MapMeta) T2) types.Map {
	if value == nil {
		return types.MapNull(elemType)
	}

	elems := TransformMap(ctx, value, p, diags, iteratee)
	mapping, d := types.MapValueFrom(ctx, elemType, elems)
	diags.Append(convertToAttrDiags(d, p)...)

	return mapping
}

// MapTypeToMap converts a types.Map first into a tfsdk aware map[string]T1 and transforms
// the result into a map[string]T2.
func MapTypeToMap[T1 any, T2 any](ctx context.Context, value types.Map, p path.Path, diags *diag.Diagnostics, iteratee func(item T1, meta MapMeta) T2) map[string]T2 {
	if !IsKnown(value) {
		return nil
	}

	elems := MapTypeAs[T1](ctx, value, p, diags)
	if diags.HasError() {
		return nil
	}

	return TransformMap(ctx, elems, p, diags, iteratee)
}

// MapTypeAs converts a types.Map into a tfsdk aware map[string]T.
func MapTypeAs[T any](ctx context.Context, value types.Map, p path.Path, diags *diag.Diagnostics) map[string]T {
	return elementsAs[map[string]T](ctx, value, p, diags)
}

// MapValueFrom converts a tfsdk aware map[string]T to a types.Map.
func MapValueFrom[T any](ctx context.Context, value map[string]T, elemType attr.Type, p path.Path, diags *diag.Diagnostics) types.Map {
	mapping, d := types.MapValueFrom(ctx, elemType, value)
	diags.Append(convertToAttrDiags(d, p)...)
	return mapping
}
