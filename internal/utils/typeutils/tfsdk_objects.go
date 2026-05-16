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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// StructToObjectType converts a tfsdk naive T1 into an types.Object of T2.
func StructToObjectType[T1 any, T2 any](
	ctx context.Context,
	value *T1,
	attrTypes map[string]attr.Type,
	p path.Path,
	diags *diag.Diagnostics,
	transformee func(item T1, meta ObjectMeta) T2,
) types.Object {
	if value == nil {
		return types.ObjectNull(attrTypes)
	}

	item := TransformObject(ctx, value, p, diags, transformee)
	obj, d := types.ObjectValueFrom(ctx, attrTypes, item)
	diags.Append(convertToAttrDiags(d, p)...)

	return obj
}

// ObjectTypeToStruct converts a types.Object first into a tfsdk aware T1 and transforms
// the result into a T2.
func ObjectTypeToStruct[T1 any, T2 any](ctx context.Context, value types.Object, p path.Path, diags *diag.Diagnostics, transformee func(item T1, meta ObjectMeta) T2) *T2 {
	if !IsKnown(value) {
		return nil
	}

	item := ObjectTypeAs[T1](ctx, value, p, diags)
	if diags.HasError() {
		return nil
	}

	return TransformObject(ctx, item, p, diags, transformee)
}

// ObjectTypeAs converts a types.Object into a tfsdk aware T.
func ObjectTypeAs[T any](ctx context.Context, value types.Object, p path.Path, diags *diag.Diagnostics) *T {
	if !IsKnown(value) {
		return nil
	}

	var item T
	d := value.As(ctx, &item, basetypes.ObjectAsOptions{})
	diags.Append(convertToAttrDiags(d, p)...)
	return &item
}

// ObjectValueFrom converts a tfsdk aware T to a types.Object.
func ObjectValueFrom[T any](ctx context.Context, value *T, attrTypes map[string]attr.Type, p path.Path, diags *diag.Diagnostics) types.Object {
	obj, d := types.ObjectValueFrom(ctx, attrTypes, value)
	diags.Append(convertToAttrDiags(d, p)...)
	return obj
}
