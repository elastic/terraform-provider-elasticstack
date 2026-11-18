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

package autofollow

import (
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ccr"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// parseSettingsRaw unmarshals a JSON settings string into a raw message map.
func parseSettingsRaw(settingsRaw string) (map[string]json.RawMessage, diag.Diagnostics) {
	return typeutils.UnmarshalJSONDiag[map[string]json.RawMessage](settingsRaw, "Failed to parse settings_raw")
}

// apiOperation identifies a CCR auto-follow lifecycle API call for testable sequencing.
type apiOperation int

const (
	opPut apiOperation = iota
	opPause
	opResume
)

func (op apiOperation) String() string {
	switch op {
	case opPut:
		return "PutAutoFollowPattern"
	case opPause:
		return "PauseAutoFollowPattern"
	case opResume:
		return "ResumeAutoFollowPattern"
	default:
		return ccr.FormatUnknownOperation(int(op))
	}
}

// updateActiveBranch classifies the active transition for the update state machine.
type updateActiveBranch int

const (
	branchActiveUnchanged updateActiveBranch = iota
	branchActiveToInactive
	branchInactiveToActive
)

func selectUpdateActiveBranch(priorActive, planActive bool) updateActiveBranch {
	switch {
	case priorActive && !planActive:
		return branchActiveToInactive
	case !priorActive && planActive:
		return branchInactiveToActive
	default:
		return branchActiveUnchanged
	}
}

func planCreateOperations(plan Model) []apiOperation {
	ops := []apiOperation{opPut}
	if !plan.Active.ValueBool() {
		ops = append(ops, opPause)
	}
	return ops
}

func planUpdateOperations(prior, plan Model) []apiOperation {
	ops := []apiOperation{opPut}
	switch selectUpdateActiveBranch(prior.Active.ValueBool(), plan.Active.ValueBool()) {
	case branchActiveToInactive:
		ops = append(ops, opPause)
	case branchInactiveToActive:
		ops = append(ops, opResume)
	}
	return ops
}
