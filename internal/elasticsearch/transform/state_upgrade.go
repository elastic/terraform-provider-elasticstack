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

package transform

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/stateutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// migrateStateV0ToV1 unwraps singleton-list nested blocks (source, destination,
// retention_policy, sync, and the inner time blocks within retention_policy and
// sync) into single objects. The schema previously modeled these as
// ListNestedBlock with SizeBetween(1,1) or SizeAtMost(1) and is now
// SingleNestedBlock. The aliases block remains a list and is left unchanged.
func migrateStateV0ToV1(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	stateMap := stateutil.UnmarshalStateMap(req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, key := range [...]string{attrSource, attrDestination, attrRetentionPolicy, attrSync} {
		resp.Diagnostics.Append(stateutil.CollapseListPath(stateMap, key, key)...)
	}
	for _, parent := range [...]string{attrRetentionPolicy, attrSync} {
		parentObj, ok := stateMap[parent].(map[string]any)
		if !ok {
			continue
		}
		resp.Diagnostics.Append(stateutil.CollapseListPath(parentObj, attrTime, parent+"."+attrTime)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// The SDK provider stored unset JSON string attributes as "" rather than
	// null. The Plugin Framework jsontypes.NormalizedType rejects empty strings,
	// so normalise them to nil before marshalling the upgraded state.
	stateutil.NullifyEmptyString(stateMap, "metadata", "pivot", "latest")
	if src, ok := stateMap[attrSource].(map[string]any); ok {
		stateutil.NullifyEmptyString(src, attrQuery, "runtime_mappings")
	}
	if dst, ok := stateMap[attrDestination].(map[string]any); ok {
		stateutil.NullifyEmptyString(dst, "pipeline")
	}

	stateutil.MarshalStateMap(stateMap, resp)
}
