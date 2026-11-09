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

package elasticdefendintegrationpolicy

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---- attr types shared by two or more OS-specific policy types ----

func malwareFullAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrMode:        types.StringType,
		attrBlocklist:   types.BoolType,
		attrOnWriteScan: types.BoolType,
		attrNotifyUser:  types.BoolType,
	}
}

func protectionModeAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrMode:      types.StringType,
		attrSupported: types.BoolType,
	}
}

func memoryProtectionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrMode:       types.StringType,
		attrSupported:  types.BoolType,
		attrCustomYara: types.BoolType,
	}
}

func deviceControlAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrEnabled:    types.BoolType,
		attrUsbStorage: types.StringType,
	}
}

func behaviorProtectionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrMode:              types.StringType,
		attrSupported:         types.BoolType,
		attrReputationService: types.BoolType,
	}
}

func loggingAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrFile: types.StringType,
	}
}

func policyAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		policyOSWindows: types.ObjectType{AttrTypes: windowsAttrTypes()},
		policyOSMac:     types.ObjectType{AttrTypes: macAttrTypes()},
		policyOSLinux:   types.ObjectType{AttrTypes: linuxAttrTypes()},
	}
}
