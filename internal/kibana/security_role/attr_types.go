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
		attrGrant:  types.SetType{ElemType: types.StringType},
		attrExcept: types.SetType{ElemType: types.StringType},
	}
}

func fieldSecurityObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: fieldSecurityAttrTypes()}
}

func esIndexResourceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrNames:         types.SetType{ElemType: types.StringType},
		attrPrivileges:    types.SetType{ElemType: types.StringType},
		attrQuery:         jsontypes.NormalizedType{},
		attrFieldSecurity: fieldSecurityObjectType(),
	}
}

func esRemoteIndexResourceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrAllowRestrictedIndices: types.BoolType,
		attrClusters:               types.SetType{ElemType: types.StringType},
		attrNames:                  types.SetType{ElemType: types.StringType},
		attrPrivileges:             types.SetType{ElemType: types.StringType},
		attrQuery:                  jsontypes.NormalizedType{},
		attrFieldSecurity:          fieldSecurityObjectType(),
	}
}

func elasticsearchResourceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrCluster:       types.SetType{ElemType: types.StringType},
		attrIndices:       types.SetType{ElemType: types.ObjectType{AttrTypes: esIndexResourceAttrTypes()}},
		attrRemoteIndices: types.SetType{ElemType: types.ObjectType{AttrTypes: esRemoteIndexResourceAttrTypes()}},
		attrRunAs:         types.SetType{ElemType: types.StringType},
	}
}

func kibanaFeatureAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrName:       types.StringType,
		attrPrivileges: types.SetType{ElemType: types.StringType},
	}
}

func kibanaBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrSpaces:  types.SetType{ElemType: types.StringType},
		attrBase:    types.SetType{ElemType: types.StringType},
		attrFeature: types.SetType{ElemType: types.ObjectType{AttrTypes: kibanaFeatureAttrTypes()}},
	}
}

func kibanaBlockObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: kibanaBlockAttrTypes()}
}
