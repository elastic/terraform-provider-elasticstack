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

package securityexceptionitem

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/stateutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// UpgradeState migrates exception item resource state to the current schema
// version (1). v0 -> v1 adds the computed created_at/created_by attributes to
// each stored comment and the optional required_comments attribute. Each
// existing comment's id and comment are preserved, the new audit attributes
// are added as null, and required_comments stays null: v0 state holds the
// full observed Kibana log, not the subset the user intended to require, so
// nothing is inferred into required_comments.
func (r *ExceptionItemResource) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {StateUpgrader: migrateExceptionItemV0ToV1},
	}
}

func migrateExceptionItemV0ToV1(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	stateutil.SetDefaultState(req, resp)

	state := stateutil.UnmarshalStateMap(req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	upgradeCommentsV0ToV1(state)

	stateutil.MarshalStateMap(state, resp)
}

func upgradeCommentsV0ToV1(state map[string]any) {
	comments, ok := state["comments"].([]any)
	if !ok {
		return
	}
	for _, raw := range comments {
		comment, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		comment["created_at"] = nil
		comment["created_by"] = nil
	}
}
