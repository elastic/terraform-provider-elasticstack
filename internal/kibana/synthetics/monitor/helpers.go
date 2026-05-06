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

package monitor

import (
	"strconv"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// mapPtr returns a pointer to v, or nil if v is nil. Unlike schemautil.MapRef,
// the key type is not restricted to string.
func mapPtr[K comparable, V any](v map[K]V) *map[K]V {
	if v == nil {
		return nil
	}
	return &v
}

// slicePtr returns a pointer to v, or nil if v is empty. Unlike schemautil.SliceRef,
// an empty (but non-nil) slice also yields nil.
func slicePtr[T any](v []T) *[]T {
	if len(v) == 0 {
		return nil
	}
	return &v
}

// stringEnumPtr converts a types.String to a typed enum pointer, returning nil for
// null, unknown, or empty-string values.
func stringEnumPtr[T ~string](v types.String) *T {
	if v.IsNull() || v.IsUnknown() || v.ValueString() == "" {
		return nil
	}
	value := T(v.ValueString())
	return &value
}

func int64ToFloat32Ptr(v types.Int64) *float32 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	value := float32(v.ValueInt64())
	return &value
}

func int64ToSyntheticsIcmpMonitorFieldsWait(v types.Int64) *kbapi.SyntheticsIcmpMonitorFields_Wait {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	wait := &kbapi.SyntheticsIcmpMonitorFields_Wait{}
	if err := wait.FromSyntheticsIcmpMonitorFieldsWait0(strconv.FormatInt(v.ValueInt64(), 10)); err != nil {
		return nil
	}
	return wait
}

func int64ToSyntheticsHTTPMonitorFieldsMaxRedirects(v types.Int64) *kbapi.SyntheticsHttpMonitorFields_MaxRedirects {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	maxRedirects := &kbapi.SyntheticsHttpMonitorFields_MaxRedirects{}
	if err := maxRedirects.FromSyntheticsHttpMonitorFieldsMaxRedirects0(strconv.FormatInt(v.ValueInt64(), 10)); err != nil {
		return nil
	}
	return maxRedirects
}
