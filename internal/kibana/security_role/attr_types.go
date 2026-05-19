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

package security_role

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func fieldSecurityAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"grant":  types.SetType{ElemType: types.StringType},
		"except": types.SetType{ElemType: types.StringType},
	}
}

func fieldSecurityObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: fieldSecurityAttrTypes()}
}

func esIndexResourceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"names":          types.SetType{ElemType: types.StringType},
		"privileges":     types.SetType{ElemType: types.StringType},
		"query":          jsontypes.NormalizedType{},
		"field_security": fieldSecurityObjectType(),
	}
}

func esRemoteIndexResourceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"clusters":       types.SetType{ElemType: types.StringType},
		"names":          types.SetType{ElemType: types.StringType},
		"privileges":     types.SetType{ElemType: types.StringType},
		"query":          jsontypes.NormalizedType{},
		"field_security": fieldSecurityObjectType(),
	}
}

func elasticsearchResourceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"cluster":        types.SetType{ElemType: types.StringType},
		"indices":        types.SetType{ElemType: types.ObjectType{AttrTypes: esIndexResourceAttrTypes()}},
		"remote_indices": types.SetType{ElemType: types.ObjectType{AttrTypes: esRemoteIndexResourceAttrTypes()}},
		"run_as":         types.SetType{ElemType: types.StringType},
	}
}

func kibanaFeatureAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":       types.StringType,
		"privileges": types.SetType{ElemType: types.StringType},
	}
}

func kibanaBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"spaces":  types.SetType{ElemType: types.StringType},
		"base":    types.SetType{ElemType: types.StringType},
		"feature": types.SetType{ElemType: types.ObjectType{AttrTypes: kibanaFeatureAttrTypes()}},
	}
}

func kibanaBlockObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: kibanaBlockAttrTypes()}
}
