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

type Elementable interface {
	attr.Value
	ElementsAs(ctx context.Context, target any, allowUnhandled bool) diag.Diagnostics
}

type ListMeta struct {
	Index int
	Path  path.Path
	Diags *diag.Diagnostics
}

type MapMeta struct {
	Key   string
	Path  path.Path
	Diags *diag.Diagnostics
}

type ObjectMeta struct {
	Path  path.Path
	Diags *diag.Diagnostics
}

func elementsAs[T any](ctx context.Context, value Elementable, p path.Path, diags *diag.Diagnostics) T {
	var result T
	if !IsKnown(value) {
		return result
	}

	d := value.ElementsAs(ctx, &result, false)
	diags.Append(convertToAttrDiags(d, p)...)
	return result
}

// TransformObject converts T1 to T2 via the transformee.
func TransformObject[T1 any, T2 any](_ context.Context, value *T1, p path.Path, diags *diag.Diagnostics, transformee func(item T1, meta ObjectMeta) T2) *T2 {
	if value == nil {
		return nil
	}

	result := transformee(*value, ObjectMeta{Path: p, Diags: diags})
	return &result
}

// TransformMap converts map[string]T1 to map[string]T2 via the iteratee.
func TransformMap[T1 any, T2 any](_ context.Context, value map[string]T1, p path.Path, diags *diag.Diagnostics, iteratee func(item T1, meta MapMeta) T2) map[string]T2 {
	if value == nil {
		return nil
	}

	elems := make(map[string]T2, len(value))
	for k, v := range value {
		elems[k] = iteratee(v, MapMeta{Key: k, Path: p.AtMapKey(k), Diags: diags})
	}

	return elems
}

// TransformSlice converts []T1 to []T2 via the iteratee.
func TransformSlice[T1 any, T2 any](_ context.Context, value []T1, p path.Path, diags *diag.Diagnostics, iteratee func(item T1, meta ListMeta) T2) []T2 {
	if value == nil {
		return nil
	}

	elems := make([]T2, len(value))
	for i, v := range value {
		elems[i] = iteratee(v, ListMeta{Index: i, Path: p.AtListIndex(i), Diags: diags})
	}

	return elems
}

// TransformSliceToMap converts []T1 to map[string]T2 via the iteratee.
func TransformSliceToMap[T1 any, T2 any](_ context.Context, value []T1, p path.Path, diags *diag.Diagnostics, iteratee func(item T1, meta ListMeta) (key string, elem T2)) map[string]T2 {
	if value == nil {
		return nil
	}

	elems := make(map[string]T2, len(value))
	for i, v := range value {
		k, v := iteratee(v, ListMeta{Index: i, Path: p.AtListIndex(i), Diags: diags})
		elems[k] = v
	}

	return elems
}

// TransformMapToSlice converts map[string]T1 to []T2 via the iteratee.
func TransformMapToSlice[T1 any, T2 any](_ context.Context, value map[string]T1, p path.Path, diags *diag.Diagnostics, iteratee func(item T1, meta MapMeta) T2) []T2 {
	if value == nil {
		return nil
	}

	elems := make([]T2, 0, len(value))
	for k, v := range value {
		v := iteratee(v, MapMeta{Key: k, Path: p.AtMapKey(k), Diags: diags})
		elems = append(elems, v)
	}

	return elems
}

func convertToAttrDiags(diags diag.Diagnostics, path path.Path) diag.Diagnostics {
	var nd diag.Diagnostics
	for _, d := range diags {
		switch d.Severity() {
		case diag.SeverityError:
			nd.AddAttributeError(path, d.Summary(), d.Detail())
		case diag.SeverityWarning:
			nd.AddAttributeWarning(path, d.Summary(), d.Detail())
		default:
			nd.Append(d)
		}
	}
	return nd
}
