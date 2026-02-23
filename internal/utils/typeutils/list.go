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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func NonEmptyListOrDefault[T any](ctx context.Context, original types.List, elemType attr.Type, slice []T) (types.List, diag.Diagnostics) {
	if len(slice) == 0 {
		return original, nil
	}

	return types.ListValueFrom(ctx, elemType, slice)
}

// EnsureTypedList converts untyped zero-value lists to properly typed null lists.
// This is commonly needed during import operations where the framework may create
// untyped lists with DynamicPseudoType elements, which causes type conversion errors.
// If the list already has a proper type, it is returned unchanged.
func EnsureTypedList(ctx context.Context, list types.List, elemType attr.Type) types.List {
	// Check if the list has no element type (nil)
	if list.ElementType(ctx) == nil {
		return types.ListNull(elemType)
	}

	// Check if the list has a dynamic pseudo type
	if _, ok := list.ElementType(ctx).(basetypes.DynamicType); ok {
		return types.ListNull(elemType)
	}

	// List is already properly typed, return as-is
	return list
}
