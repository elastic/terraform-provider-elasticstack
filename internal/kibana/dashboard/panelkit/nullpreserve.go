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

package panelkit

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// PreserveString keeps an existing null/unknown string when it is not known; otherwise updates from the API pointer.
func PreserveString(existing types.String, api *string) types.String {
	if !typeutils.IsKnown(existing) {
		return existing
	}
	return types.StringPointerValue(api)
}

// PreserveBool keeps an existing null/unknown bool when it is not known; otherwise updates from the API pointer.
func PreserveBool(existing types.Bool, api *bool) types.Bool {
	if !typeutils.IsKnown(existing) {
		return existing
	}
	return types.BoolPointerValue(api)
}

// PreserveFloat64 keeps an existing null/unknown float when it is not known; otherwise updates from the API pointer.
func PreserveFloat64(existing types.Float64, api *float64) types.Float64 {
	if !typeutils.IsKnown(existing) {
		return existing
	}
	return types.Float64PointerValue(api)
}

// PreserveInt64 keeps an existing null/unknown int64 when it is not known; otherwise updates from the API pointer.
func PreserveInt64(existing types.Int64, api *int64) types.Int64 {
	if !typeutils.IsKnown(existing) {
		return existing
	}
	return types.Int64PointerValue(api)
}

// PreserveFloat32 keeps an existing null/unknown float32 when it is not known; otherwise updates from the API pointer.
// Used for float32 API fields stored as types.Float32 to avoid float32↔float64 round-trip drift.
func PreserveFloat32(existing types.Float32, api *float32) types.Float32 {
	if !typeutils.IsKnown(existing) {
		return existing
	}
	return types.Float32PointerValue(api)
}

// PreserveList keeps an existing null/unknown list when it is not known; otherwise replaces with next.
func PreserveList(existing, next attr.Value) attr.Value {
	if !typeutils.IsKnown(existing) {
		return existing
	}
	return next
}
