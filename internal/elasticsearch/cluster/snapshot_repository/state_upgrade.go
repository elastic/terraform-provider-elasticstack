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

package snapshot_repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var typeBlockKeys = [...]string{"fs", "url", "gcs", "azure", "s3", "hdfs"}

// migrateSnapshotRepositoryStateV0ToV1 unwraps list-wrapped nested blocks from
// schema version 0 (SDK) into single objects for Plugin Framework SingleNestedBlock.
func migrateSnapshotRepositoryStateV0ToV1(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	if req.RawState == nil || req.RawState.JSON == nil {
		resp.Diagnostics.AddError("Invalid raw state", "Raw state or JSON is nil")
		return
	}

	var stateMap map[string]any
	err := json.Unmarshal(req.RawState.JSON, &stateMap)
	if err != nil {
		resp.Diagnostics.AddError("State upgrade error", "Could not unmarshal prior state: "+err.Error())
		return
	}

	for _, key := range typeBlockKeys {
		if raw, ok := stateMap[key]; ok {
			u, err := unwrapSingletonList(raw, key)
			if err != nil {
				resp.Diagnostics.AddError("State upgrade error", err.Error())
				return
			}
			if u == nil {
				delete(stateMap, key)
			} else {
				stateMap[key] = u
			}
		}
	}

	stateJSON, err := json.Marshal(stateMap)
	if err != nil {
		resp.Diagnostics.AddError("State upgrade error", "Could not marshal new state: "+err.Error())
		return
	}
	resp.DynamicValue = &tfprotov6.DynamicValue{
		JSON: stateJSON,
	}
}

func unwrapSingletonList(v any, key string) (any, error) {
	if v == nil {
		return nil, nil
	}
	list, ok := v.([]any)
	if !ok {
		// Already an object (or absent); PF SingleNestedBlock state shape.
		return v, nil
	}
	if len(list) == 0 {
		return nil, nil
	}
	if len(list) > 1 {
		return nil, fmt.Errorf("unexpected multi-element array at path %q", key)
	}
	first := list[0]
	obj, ok := first.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected non-object element at path %q", key)
	}
	return obj, nil
}
