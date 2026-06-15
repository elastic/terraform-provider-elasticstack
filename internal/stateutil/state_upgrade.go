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

package stateutil

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// UnmarshalStateMap decodes req.RawState.JSON into a map[string]any for use in
// state upgrade functions. Returns nil and sets a diagnostic error on failure.
func UnmarshalStateMap(req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) map[string]any {
	if req.RawState == nil || req.RawState.JSON == nil {
		resp.Diagnostics.AddError("Invalid raw state", "Raw state or JSON is nil")
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(req.RawState.JSON, &m); err != nil {
		resp.Diagnostics.AddError("State upgrade error", "Could not unmarshal prior state: "+err.Error())
		return nil
	}
	return m
}

// MarshalStateMap encodes m as JSON and assigns it to resp.DynamicValue.
func MarshalStateMap(m map[string]any, resp *resource.UpgradeStateResponse) {
	stateJSON, err := json.Marshal(m)
	if err != nil {
		resp.Diagnostics.AddError("State upgrade error", "Could not marshal new state: "+err.Error())
		return
	}
	resp.DynamicValue = &tfprotov6.DynamicValue{JSON: stateJSON}
}
