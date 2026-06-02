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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StringSetOrNull converts a []string to a types.Set. Returns a null string
// set when src is empty so that unset attributes are stored as null in state.
func StringSetOrNull(ctx context.Context, src []string) (types.Set, diag.Diagnostics) {
	if len(src) == 0 {
		return types.SetNull(types.StringType), nil
	}
	return types.SetValueFrom(ctx, types.StringType, src)
}

// NonEmptySetOrDefault returns the original set if slice is empty,
// otherwise converts slice into a types.Set.
func NonEmptySetOrDefault[T any](ctx context.Context, original types.Set, elemType attr.Type, slice []T) (types.Set, diag.Diagnostics) {
	if len(slice) == 0 {
		return original, nil
	}

	return types.SetValueFrom(ctx, elemType, slice)
}

// SetTypeAs converts a types.Set into a tfsdk aware []T.
func SetTypeAs[T any](ctx context.Context, value types.Set, p path.Path, diags *diag.Diagnostics) []T {
	return elementsAs[[]T](ctx, value, p, diags)
}

// SetValueFrom converts a tfsdk aware []T to a types.Set.
func SetValueFrom[T any](ctx context.Context, value []T, elemType attr.Type, p path.Path, diags *diag.Diagnostics) types.Set {
	list, d := types.SetValueFrom(ctx, elemType, value)
	diags.Append(convertToAttrDiags(d, p)...)
	return list
}

// StringSetElements extracts the string values from a types.Set of strings
// without requiring a context.Context. Returns nil for null/unknown sets and
// appends an error diagnostic for non-string or unknown elements.
func StringSetElements(set types.Set, diags *diag.Diagnostics) []string {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}
	elems := make([]string, 0, len(set.Elements()))
	for _, elem := range set.Elements() {
		str, ok := elem.(types.String)
		if !ok || str.IsUnknown() {
			if !ok {
				diags.AddError("Invalid set element type", "expected types.String")
			} else {
				diags.AddError("Unknown set element", "set elements cannot be unknown")
			}
			continue
		}
		elems = append(elems, str.ValueString())
	}
	return elems
}
