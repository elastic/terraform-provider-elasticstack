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

package dataview

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

// TestFieldAttrElemType_matchesSchema guards against the documented drift hazard from
// hardcoding getFieldAttrElemType() instead of deriving from the schema (which would recurse
// because schema's field_attrs CustomType is constructed from getFieldAttrElemType itself).
//
// If a future rename of `custom_label` or `count` happens in the schema's NestedAttributeObject,
// this test will fail loudly so the helper is updated in lock-step.
func TestFieldAttrElemType_matchesSchema(t *testing.T) {
	t.Parallel()

	dataViewAttr, ok := getSchema().Attributes["data_view"].(schema.SingleNestedAttribute)
	require.True(t, ok, "expected data_view to be SingleNestedAttribute")

	fieldAttrsAttr, ok := dataViewAttr.Attributes["field_attrs"].(schema.MapNestedAttribute)
	require.True(t, ok, "expected data_view.field_attrs to be MapNestedAttribute")

	schemaElem := schemaNestedObjectElemType(fieldAttrsAttr.NestedObject)
	require.Equal(t, schemaElem, getFieldAttrElemType(),
		"getFieldAttrElemType() drifted from the schema's field_attrs nested object; "+
			"update both sides together")
}

// schemaNestedObjectElemType constructs the ObjectType implied by a MapNestedAttribute's
// NestedAttributeObject so we can compare it against the helper without recursing through the
// schema's CustomType (which is itself built from getFieldAttrElemType).
func schemaNestedObjectElemType(no schema.NestedAttributeObject) attr.Type {
	attrTypes := make(map[string]attr.Type, len(no.Attributes))
	for name, a := range no.Attributes {
		attrTypes[name] = a.GetType()
	}
	return types.ObjectType{AttrTypes: attrTypes}
}
