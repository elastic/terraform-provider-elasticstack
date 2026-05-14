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

package contracttest

import (
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type leafPaths struct {
	// Optional semantics verified by null-preservation checks. Collection-level paths
	// (List/ListNested/Map Optional) plus scalars nested under SingleNested only — never
	// individual elements inside Optional ListNested rows.
	optional [][]string
	// Required leaves for fixture-presence validation and shallow ValidatePanelConfig passes.
	required [][]string
}

func collectLeafPaths(root schema.SingleNestedAttribute) leafPaths {
	var out leafPaths
	for name, a := range root.Attributes {
		walkAttributes(a, []string{name}, false, &out)
	}
	return out
}

func schemaAttributeAt(root schema.SingleNestedAttribute, path []string) (schema.Attribute, bool) {
	var cur schema.Attribute = root
	for _, seg := range path {
		next, ok := schemaAttrChild(cur, seg)
		if !ok {
			return nil, false
		}
		cur = next
	}
	return cur, true
}

func schemaAttrChild(parent schema.Attribute, name string) (schema.Attribute, bool) {
	switch p := parent.(type) {
	case schema.SingleNestedAttribute:
		child, ok := p.Attributes[name]
		return child, ok
	case schema.ListNestedAttribute:
		child, ok := p.NestedObject.Attributes[name]
		return child, ok
	default:
		return nil, false
	}
}

func walkAttributes(a schema.Attribute, parent []string, insideListNested bool, out *leafPaths) {
	switch at := a.(type) {
	case schema.StringAttribute:
		recordScalarLeaf(parent, at.Required, at.Optional, at.Computed, insideListNested, out)
	case schema.BoolAttribute:
		recordScalarLeaf(parent, at.Required, at.Optional, at.Computed, insideListNested, out)
	case schema.Float64Attribute:
		recordScalarLeaf(parent, at.Required, at.Optional, at.Computed, insideListNested, out)
	case schema.Int64Attribute:
		recordScalarLeaf(parent, at.Required, at.Optional, at.Computed, insideListNested, out)

	case schema.ListAttribute:
		recordLeafCollection(parent, at.Required, at.Optional, at.Computed, insideListNested, out)

	case schema.MapAttribute:
		recordLeafCollection(parent, at.Required, at.Optional, at.Computed, insideListNested, out)

	case schema.ListNestedAttribute:
		listPath := slices.Clone(parent)
		recordLeafCollection(parent, at.Required, at.Optional, at.Computed, insideListNested, out)
		for childName, child := range at.NestedObject.Attributes {
			walkAttributes(child, append(slices.Clone(listPath), childName), true, out)
		}

	case schema.SingleNestedAttribute:
		if at.Computed {
			return
		}
		for childName, child := range at.Attributes {
			walkAttributes(child, append(slices.Clone(parent), childName), insideListNested, out)
		}
	}
}

func recordScalarLeaf(path []string, required, optional, computed, insideListNested bool, out *leafPaths) {
	if computed || len(path) == 0 {
		return
	}
	path = slices.Clone(path)
	if required {
		out.required = append(out.required, path)
		return
	}
	if optional && !required && !insideListNested {
		out.optional = append(out.optional, path)
	}
}

func recordLeafCollection(path []string, required, optional, computed, insideListNested bool, out *leafPaths) {
	if computed || len(path) == 0 {
		return
	}
	path = slices.Clone(path)
	if required {
		out.required = append(out.required, path)
	}
	if optional && !required && !insideListNested {
		out.optional = append(out.optional, path)
	}
}
