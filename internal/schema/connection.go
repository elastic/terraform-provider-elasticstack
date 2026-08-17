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

package schema

import (
	"fmt"
	"maps"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	fwschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// KibanaConnectionNullList returns a properly-typed null list value for the
// kibana_connection block. Use this when building a state struct from scratch
// (e.g., in ImportState or state upgraders) so the framework can match the
// list element type against the schema instead of encountering a zero-value.
func KibanaConnectionNullList() types.List {
	return types.ListNull(KibanaConnectionObjectType())
}

// KibanaConnectionObjectType returns the object type for kibana_connection list
// elements. Managed and ephemeral resources use the same connection block shape.
func KibanaConnectionObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: kibanaConnectionBlockObjectAttrTypes(),
	}
}

// ElasticsearchConnectionNullList returns a properly-typed null list value for the
// elasticsearch_connection block. Use when building state in ImportState so the
// framework list element type matches the resource schema.
func ElasticsearchConnectionNullList() types.List {
	return types.ListNull(ElasticsearchConnectionObjectType())
}

// ElasticsearchConnectionObjectType returns the object type for elasticsearch_connection
// list elements. Managed and ephemeral resources use the same connection block shape.
func ElasticsearchConnectionObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: elasticsearchConnectionBlockObjectAttrTypes(),
	}
}

func GetEsFWConnectionBlock() fwschema.Block {
	return esConnectionBlockSpec().fwBlock()
}

func GetKbFWConnectionBlock() fwschema.Block {
	return kbConnectionBlockSpec().fwBlock()
}

func GetFleetFWConnectionBlock() fwschema.Block {
	return fleetConnectionBlockSpec().fwBlock()
}

var (
	// elasticsearchConnectionBlockObjectAttrTypesVal is the canonical
	// attribute-type map for the elasticsearch_connection block, derived once
	// from esConnectionBlockSpec(). Deriving from the spec (rather than the
	// built schema, with a hand-maintained fallback) keeps the object type and
	// the schema from drifting apart.
	elasticsearchConnectionBlockObjectAttrTypesVal = sync.OnceValue(func() map[string]attr.Type {
		return esConnectionBlockSpec().attrTypes()
	})

	// kibanaConnectionBlockObjectAttrTypesVal is the canonical attribute-type
	// map for the kibana_connection block, derived once from
	// kbConnectionBlockSpec().
	kibanaConnectionBlockObjectAttrTypesVal = sync.OnceValue(func() map[string]attr.Type {
		return kbConnectionBlockSpec().attrTypes()
	})
)

// copyAttrTypes returns a shallow copy of src so callers cannot mutate the
// cached attribute-type map shared across calls.
func copyAttrTypes(src map[string]attr.Type) map[string]attr.Type {
	out := make(map[string]attr.Type, len(src))
	maps.Copy(out, src)
	return out
}

func elasticsearchConnectionBlockObjectAttrTypes() map[string]attr.Type {
	return copyAttrTypes(elasticsearchConnectionBlockObjectAttrTypesVal())
}

func kibanaConnectionBlockObjectAttrTypes() map[string]attr.Type {
	return copyAttrTypes(kibanaConnectionBlockObjectAttrTypesVal())
}

func fwNestedBlockAttributesToAttrTypes(attrs map[string]fwschema.Attribute) (map[string]attr.Type, error) {
	out := make(map[string]attr.Type, len(attrs))
	for name, a := range attrs {
		t, err := fwAttributeToAttrType(name, a)
		if err != nil {
			return nil, err
		}
		out[name] = t
	}
	return out, nil
}

func fwAttributeToAttrType(name string, a fwschema.Attribute) (attr.Type, error) {
	switch a := a.(type) {
	case fwschema.StringAttribute:
		if a.CustomType != nil {
			return a.CustomType, nil
		}
		return types.StringType, nil
	case fwschema.BoolAttribute:
		if a.CustomType != nil {
			return a.CustomType, nil
		}
		return types.BoolType, nil
	case fwschema.ListAttribute:
		if a.CustomType != nil {
			return a.CustomType, nil
		}
		if a.ElementType == nil {
			return nil, fmt.Errorf("attribute %q: ListAttribute missing ElementType", name)
		}
		return types.ListType{ElemType: a.ElementType}, nil
	case fwschema.MapAttribute:
		if a.CustomType != nil {
			return a.CustomType, nil
		}
		if a.ElementType == nil {
			return nil, fmt.Errorf("attribute %q: MapAttribute missing ElementType", name)
		}
		return types.MapType{ElemType: a.ElementType}, nil
	default:
		return nil, fmt.Errorf("attribute %q: unsupported framework attribute type %T (extend fwAttributeToAttrType)", name, a)
	}
}
