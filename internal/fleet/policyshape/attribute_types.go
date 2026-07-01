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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Attribute name constants shared by every caller building an inputs/streams
// schema on top of this package.
const (
	AttrEnabled   = "enabled"
	AttrCondition = "condition"
	AttrVars      = "vars"
	AttrDefaults  = "defaults"
	AttrStreams   = "streams"
)

// StreamAttributeTypes returns the attribute types map for a single stream
// object: enabled, condition, and vars.
func StreamAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		AttrEnabled:   types.BoolType,
		AttrCondition: types.StringType,
		AttrVars:      jsontypes.NormalizedType{},
	}
}

// StreamType returns the attr.Type of a single stream object (see
// StreamAttributeTypes).
func StreamType() attr.Type {
	return types.ObjectType{AttrTypes: StreamAttributeTypes()}
}

// InputDefaultsAttributeTypes returns the attribute types map for the
// package-computed `defaults` object nested under an input: vars, plus a map
// of per-stream defaults (enabled + vars; streams have no default condition).
func InputDefaultsAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		AttrVars: jsontypes.NormalizedType{},
		AttrStreams: types.MapType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					AttrEnabled: types.BoolType,
					AttrVars:    jsontypes.NormalizedType{},
				},
			},
		},
	}
}

// InputDefaultsType returns the attr.Type of the `defaults` object nested
// under an input (see InputDefaultsAttributeTypes).
func InputDefaultsType() attr.Type {
	return types.ObjectType{AttrTypes: InputDefaultsAttributeTypes()}
}

// InputAttributeTypes returns the attribute types map for a single input
// object: enabled, condition, vars, the package-computed defaults object, and
// the nested streams map.
func InputAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		AttrEnabled:   types.BoolType,
		AttrCondition: types.StringType,
		AttrVars:      jsontypes.NormalizedType{},
		AttrDefaults:  InputDefaultsType(),
		AttrStreams: types.MapType{
			ElemType: StreamType(),
		},
	}
}

// InputElementType returns the InputType used as the element type of the
// top-level inputs map (see InputAttributeTypes).
func InputElementType() InputType {
	return NewInputType(InputAttributeTypes())
}
