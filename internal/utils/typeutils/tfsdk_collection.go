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
)

// collectionFactory creates a collection type C from a context, element type, and value.
type collectionFactory[C attr.Value] func(context.Context, attr.Type, any) (C, diag.Diagnostics)

// collectionValueFrom converts a []T to a collection type C (types.List or types.Set)
// using the provided factory function.
func collectionValueFrom[T any, C attr.Value](
	ctx context.Context,
	value []T,
	elemType attr.Type,
	p path.Path,
	diags *diag.Diagnostics,
	factory collectionFactory[C],
) C {
	result, d := factory(ctx, elemType, value)
	diags.Append(convertToAttrDiags(d, p)...)
	return result
}

// nonEmptyCollectionOrDefault returns original if slice is empty, otherwise converts
// slice into a collection type C (types.List or types.Set) using the provided factory.
func nonEmptyCollectionOrDefault[T any, C attr.Value](
	ctx context.Context,
	original C,
	elemType attr.Type,
	slice []T,
	factory collectionFactory[C],
) (C, diag.Diagnostics) {
	if len(slice) == 0 {
		return original, nil
	}
	return factory(ctx, elemType, slice)
}
