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

// MapRef takes the reference of the given map value. If the value is nil, it returns nil rather
// than a pointer to nil.
func MapRef[K comparable, T any, M ~map[K]T](value M) *M {
	if value == nil {
		return nil
	}
	return &value
}

// SliceRef takes the reference of the given slice value. If the value is nil, it returns nil
// rather than a pointer to nil.
func SliceRef[T any, S ~[]T](value S) *S {
	if value == nil {
		return nil
	}
	return &value
}

// SliceNilIfEmpty takes the reference of the given slice value. If the value is nil or empty, it
// returns nil rather than a pointer to an empty slice.
func SliceNilIfEmpty[T any, S ~[]T](value S) *S {
	if len(value) == 0 {
		return nil
	}
	return &value
}

// Float32Ptr converts a float64 to a *float32.
func Float32Ptr(v float64) *float32 {
	f := float32(v)
	return &f
}

// Deref returns the value referenced by the given pointer. If the pointer is nil, a zero value is
// returned.
func Deref[T any](value *T) T {
	if value == nil {
		var zero T
		return zero
	}
	return *value
}

// DefaultIfNil returns the dereferenced value of the pointer, or the zero value of T if the
// pointer is nil. Deprecated: use Deref instead.
func DefaultIfNil[T any](value *T) T {
	return Deref(value)
}

// NonNilSlice returns an empty slice if s is nil. Guarantees that json.Marshal and terraform
// parameters will not treat the empty slice as null.
func NonNilSlice[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

// Itol converts *int to *int64.
func Itol(value *int) *int64 {
	if value == nil {
		return nil
	}
	return new(int64(*value))
}

// Ltoi converts *int64 to *int.
func Ltoi(value *int64) *int {
	if value == nil {
		return nil
	}
	return new(int(*value))
}

// NonEmptyStringPtr returns a pointer to s, or nil if s is empty.
func NonEmptyStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// DerefOrElse returns the value pointed to by val if it is non-nil and non-empty,
// otherwise it returns def. It is the inverse of NonEmptyStringPtr.
func DerefOrElse(val *string, def string) string {
	if val != nil && *val != "" {
		return *val
	}
	return def
}
