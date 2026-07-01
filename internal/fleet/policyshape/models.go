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

package policyshape

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// InputModel is the Go representation of a single element of the top-level
// `inputs` map attribute. It backs InputValue's semantic-equality logic and
// is the shape callers decode/encode individual inputs through when
// converting to/from the Fleet API.
//
// Condition and Vars are user-configurable; Defaults is populated from Fleet
// package metadata (see PackageInfoToDefaults) and is used only to compute
// semantic equality against unset user values — it is not sent to the API.
type InputModel struct {
	Enabled   types.Bool           `tfsdk:"enabled"`
	Condition types.String         `tfsdk:"condition"`
	Vars      jsontypes.Normalized `tfsdk:"vars"`
	Defaults  types.Object         `tfsdk:"defaults"` // > InputDefaultsModel
	Streams   types.Map            `tfsdk:"streams"`  // > InputStreamModel
}

// InputStreamModel is the Go representation of a single stream nested under
// an input.
type InputStreamModel struct {
	Enabled   types.Bool           `tfsdk:"enabled"`
	Condition types.String         `tfsdk:"condition"`
	Vars      jsontypes.Normalized `tfsdk:"vars"`
}

// InputDefaultsModel captures the package-supplied default values for an
// input, computed from Fleet package metadata (policy templates / data
// streams) by PackageInfoToDefaults. It has no Condition counterpart: the
// Fleet package registry does not publish default condition expressions.
type InputDefaultsModel struct {
	Vars    jsontypes.Normalized                `tfsdk:"vars"`
	Streams map[string]InputDefaultsStreamModel `tfsdk:"streams"`
}

// InputDefaultsStreamModel is the per-stream counterpart of InputDefaultsModel.
type InputDefaultsStreamModel struct {
	Enabled types.Bool           `tfsdk:"enabled"`
	Vars    jsontypes.Normalized `tfsdk:"vars"`
}
