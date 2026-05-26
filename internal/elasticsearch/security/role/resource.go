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

package role

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                 = newRoleResource()
	_ resource.ResourceWithConfigure    = newRoleResource()
	_ resource.ResourceWithImportState  = newRoleResource()
	_ resource.ResourceWithUpgradeState = newRoleResource()
)

type roleResource struct {
	*entitycore.ElasticsearchResource[Data]
}

func newRoleResource() *roleResource {
	return &roleResource{
		ElasticsearchResource: entitycore.NewElasticsearchResource[Data]("security_role", entitycore.ElasticsearchResourceOptions[Data]{
			Schema: getSchemaFactory,
			Read:   readRole,
			Delete: deleteRole,
			Create: writeRole,
			Update: writeRole,
		}),
	}
}

func NewRoleResource() resource.Resource {
	return newRoleResource()
}

func (r *roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *roleResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			StateUpgrader: v0ToV1,
		},
	}
}

func v0ToV1(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var priorState map[string]any
	err := json.Unmarshal(req.RawState.JSON, &priorState)
	if err != nil {
		resp.Diagnostics.AddError("State Upgrade Error", "Could not unmarshal prior state: "+err.Error())
		return
	}

	if global := priorState[attrGlobal]; global == nil || global == "" {
		delete(priorState, attrGlobal)
	}

	if metadata := priorState[attrMetadata]; metadata == nil || metadata == "" {
		delete(priorState, attrMetadata)
	}

	indices, ok := priorState[blockIndices]
	if ok {
		priorState[blockIndices] = convertV0Indices(indices)
	}

	remoteIndices, ok := priorState[blockRemoteIndices]
	if ok {
		priorState[blockRemoteIndices] = convertV0Indices(remoteIndices)
	}

	stateJSON, err := json.Marshal(priorState)
	if err != nil {
		resp.Diagnostics.AddError("State Upgrade Error", "Could not marshal new state: "+err.Error())
		return
	}
	resp.DynamicValue = &tfprotov6.DynamicValue{
		JSON: stateJSON,
	}
}

func convertV0Indices(indices any) any {
	indicesSlice, ok := indices.([]any)
	if ok {
		for i, index := range indicesSlice {
			indexMap, ok := index.(map[string]any)
			if ok {
				if indexMap[attrQuery] == "" {
					delete(indexMap, attrQuery)
				}
				// Convert field_security from a list to an object
				if fs, ok := indexMap[attrFieldSecurity]; ok {
					fsList, ok := fs.([]any)
					if ok && len(fsList) > 0 {
						indexMap[attrFieldSecurity] = fsList[0]
					} else {
						delete(indexMap, attrFieldSecurity)
					}
				}
				indicesSlice[i] = indexMap
			}
		}
	}
	return indicesSlice
}
