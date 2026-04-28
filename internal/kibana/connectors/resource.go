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

package connectors

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                 = newResource()
	_ resource.ResourceWithConfigure    = newResource()
	_ resource.ResourceWithImportState  = newResource()
	_ resource.ResourceWithUpgradeState = newResource()

	MinVersionSupportingPreconfiguredIDs = version.Must(version.NewVersion("8.8.0"))
)

type Resource struct {
	*entitycore.ResourceBase
}

func newResource() *Resource {
	return &Resource{
		ResourceBase: entitycore.NewResourceBase(entitycore.ComponentKibana, "action_connector"),
	}
}

// NewResource returns a configured resource for provider registration.
func NewResource() resource.Resource {
	return newResource()
}

func (r *Resource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), request.ID)...)
}

func (r *Resource) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {StateUpgrader: upgradeV0},
	}
}

// The schema between V0 and V1 is mostly the same, however config saved ""
// values to the state when null values were in the config. jsontypes.Normalized
// correctly states this is invalid JSON.
func upgradeV0(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var state map[string]any

	removeEmptyString := func(state map[string]any, key string) map[string]any {
		value, ok := state[key]
		if !ok {
			return state
		}

		valueString, ok := value.(string)
		if !ok || valueString != "" {
			return state
		}

		delete(state, key)
		return state
	}

	err := json.Unmarshal(req.RawState.JSON, &state)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal state", err.Error())
		return
	}

	state = removeEmptyString(state, "config")
	state = removeEmptyString(state, "secrets")

	stateBytes, err := json.Marshal(state)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal state", err.Error())
		return
	}

	resp.DynamicValue = &tfprotov6.DynamicValue{
		JSON: stateBytes,
	}
}
