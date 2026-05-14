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

package lenscommon

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MapOptionalBoolWithSnapshotDefault maps an optional API bool to a Terraform Bool,
// preserving snapshot defaults (e.g. when the API returns false and the user hasn't set it).
func MapOptionalBoolWithSnapshotDefault(current types.Bool, apiValue *bool, snapshotDefault bool) types.Bool {
	switch {
	case apiValue == nil:
		if typeutils.IsKnown(current) {
			return current
		}
		return types.BoolNull()
	case typeutils.IsKnown(current) && *apiValue == snapshotDefault && current.ValueBool() != *apiValue:
		return current
	case !typeutils.IsKnown(current) && *apiValue == snapshotDefault:
		return types.BoolNull()
	default:
		return types.BoolValue(*apiValue)
	}
}

// MapOptionalFloatWithSnapshotDefault maps an optional API float to a Terraform Float64,
// preserving snapshot defaults.
func MapOptionalFloatWithSnapshotDefault(current types.Float64, apiValue *float32, snapshotDefault float64) types.Float64 {
	switch {
	case apiValue == nil:
		if typeutils.IsKnown(current) {
			return current
		}
		return types.Float64Null()
	case typeutils.IsKnown(current) && float64(*apiValue) == snapshotDefault && current.ValueFloat64() != float64(*apiValue):
		return current
	case !typeutils.IsKnown(current) && float64(*apiValue) == snapshotDefault:
		return types.Float64Null()
	default:
		return types.Float64Value(float64(*apiValue))
	}
}
