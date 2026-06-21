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

package entity

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func BlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":            types.StringType,
		attrName:        types.StringType,
		attrType:        types.StringType,
		"sub_type":      types.StringType,
		"source":        types.SetType{ElemType: types.StringType},
		"attributes":    types.ObjectType{AttrTypes: AttributesBlockAttrTypes()},
		"behaviors":     types.ObjectType{AttrTypes: BehaviorsBlockAttrTypes()},
		"lifecycle":     types.ObjectType{AttrTypes: LifecycleBlockAttrTypes()},
		attrRisk:        types.ObjectType{AttrTypes: RiskBlockAttrTypes()},
		"relationships": types.ObjectType{AttrTypes: RelationshipsBlockAttrTypes()},
	}
}

func AttributesBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrAsset:     types.BoolType,
		"managed":     types.BoolType,
		"privileged":  types.BoolType,
		"mfa_enabled": types.BoolType,
	}
}

func BehaviorsBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"brute_force_victim": types.BoolType,
		"new_country_login":  types.BoolType,
		"used_usb_device":    types.BoolType,
	}
}

func LifecycleBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"first_seen":    types.StringType,
		"last_seen":     types.StringType,
		"last_activity": types.StringType,
	}
}

func RiskBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrCalculatedLevel:     types.StringType,
		attrCalculatedScore:     types.Float64Type,
		attrCalculatedScoreNorm: types.Float64Type,
	}
}

func RelationshipsBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"owned_by":               types.SetType{ElemType: types.StringType},
		"owns":                   types.SetType{ElemType: types.StringType},
		"supervised_by":          types.SetType{ElemType: types.StringType},
		"supervises":             types.SetType{ElemType: types.StringType},
		"depends_on":             types.SetType{ElemType: types.StringType},
		"dependent_of":           types.SetType{ElemType: types.StringType},
		"communicates_with":      types.SetType{ElemType: types.StringType},
		"accesses_frequently":    types.SetType{ElemType: types.StringType},
		"accessed_frequently_by": types.SetType{ElemType: types.StringType},
		"accesses_infrequently":  types.SetType{ElemType: types.StringType},
	}
}

func HostBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrName:       types.StringType,
		attrDomain:     types.SetType{ElemType: types.StringType},
		"hostname":     types.SetType{ElemType: types.StringType},
		"id":           types.SetType{ElemType: types.StringType},
		"ip":           types.SetType{ElemType: types.StringType},
		"mac":          types.SetType{ElemType: types.StringType},
		attrType:       types.SetType{ElemType: types.StringType},
		"architecture": types.SetType{ElemType: types.StringType},
		"os":           types.ObjectType{AttrTypes: HostOsBlockAttrTypes()},
		attrRisk:       types.ObjectType{AttrTypes: RiskBlockAttrTypes()},
	}
}

func HostOsBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"family":   types.StringType,
		"full":     types.StringType,
		"kernel":   types.StringType,
		attrName:   types.StringType,
		"platform": types.StringType,
		attrType:   types.StringType,
		"version":  types.StringType,
	}
}

func UserBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrName:    types.StringType,
		attrDomain:  types.SetType{ElemType: types.StringType},
		attrEmail:   types.SetType{ElemType: types.StringType},
		"full_name": types.SetType{ElemType: types.StringType},
		"hash":      types.SetType{ElemType: types.StringType},
		"id":        types.SetType{ElemType: types.StringType},
		"roles":     types.SetType{ElemType: types.StringType},
		attrRisk:    types.ObjectType{AttrTypes: RiskBlockAttrTypes()},
	}
}

func ServiceBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrName: types.StringType,
		attrRisk: types.ObjectType{AttrTypes: RiskBlockAttrTypes()},
	}
}

func OrchestratorBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrName:          types.StringType,
		attrType:          types.StringType,
		"namespace":       types.StringType,
		"cluster_id":      types.StringType,
		"cluster_name":    types.StringType,
		"cluster_version": types.StringType,
		"resource_id":     types.StringType,
		"resource_name":   types.StringType,
		"resource_type":   types.StringType,
	}
}

func CloudBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrProvider:   types.StringType,
		"region":       types.StringType,
		"account_id":   types.StringType,
		"account_name": types.StringType,
		"project_id":   types.StringType,
		"project_name": types.StringType,
		"service_name": types.StringType,
	}
}

func EventBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"category":   types.StringType,
		attrType:     types.StringType,
		"dataset":    types.StringType,
		"kind":       types.StringType,
		"outcome":    types.StringType,
		attrProvider: types.StringType,
		"action":     types.StringType,
		"code":       types.StringType,
		"reference":  types.StringType,
		attrReason:   types.StringType,
		"severity":   types.StringType,
		"timezone":   types.StringType,
		"url":        types.StringType,
		"ingested":   types.StringType,
	}
}

func AssetBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"criticality":          types.StringType,
		"criticality_feedback": types.ObjectType{AttrTypes: AssetCriticalityFeedbackBlockAttrTypes()},
		"owner":                types.ObjectType{AttrTypes: AssetOwnerBlockAttrTypes()},
		attrValue:              types.Float64Type,
	}
}

func AssetCriticalityFeedbackBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"notes":    types.StringType,
		attrReason: types.StringType,
	}
}

func AssetOwnerBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrName:     types.StringType,
		"department": types.StringType,
		attrEmail:    types.StringType,
		"ext":        types.StringType,
	}
}
