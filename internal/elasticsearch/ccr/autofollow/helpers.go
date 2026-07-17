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
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/typedapi/ccr/putautofollowpattern"
	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ccr"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
		return fmt.Sprintf("apiOperation(%d)", op)
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

func buildPutAutoFollowPatternRequest(ctx context.Context, model Model) (*putautofollowpattern.Request, diag.Diagnostics) {
	var diags diag.Diagnostics

	leaderPatterns := typeutils.ListTypeToSliceString(ctx, model.LeaderIndexPatterns, path.Root("leader_index_patterns"), &diags)
	if diags.HasError() {
		return nil, diags
	}

	req := &putautofollowpattern.Request{
		RemoteCluster:       model.RemoteCluster.ValueString(),
		LeaderIndexPatterns: leaderPatterns,
		FollowIndexPattern:  typeutils.OptionalString(model.FollowIndexPattern),
	}

	if typeutils.IsKnown(model.LeaderIndexExclusionPatterns) {
		req.LeaderIndexExclusionPatterns = typeutils.ListTypeToSliceString(
			ctx,
			model.LeaderIndexExclusionPatterns,
			path.Root("leader_index_exclusion_patterns"),
			&diags,
		)
		if diags.HasError() {
			return nil, diags
		}
	}

	if typeutils.IsKnown(model.SettingsRaw) {
		settings, settingsDiags := parseSettingsRaw(model.SettingsRaw.ValueString())
		diags.Append(settingsDiags...)
		if diags.HasError() {
			return nil, diags
		}
		req.Settings = settings
	}

	// max_outstanding_read_requests is Optional+Computed and mapped from the GET
	// API, which omits the field when it was never set (decoding to 0). The
	// ApplyToPutAutoFollowRequest helper skips non-positive values so the
	// Computed zero echoed back on update is not rejected by Elasticsearch.
	tuning := ccr.TuningParams{
		MaxOutstandingReadRequests:    model.MaxOutstandingReadRequests,
		MaxOutstandingWriteRequests:   model.MaxOutstandingWriteRequests,
		MaxReadRequestOperationCount:  model.MaxReadRequestOperationCount,
		MaxReadRequestSize:            model.MaxReadRequestSize,
		MaxRetryDelay:                 model.MaxRetryDelay,
		MaxWriteBufferCount:           model.MaxWriteBufferCount,
		MaxWriteBufferSize:            model.MaxWriteBufferSize,
		MaxWriteRequestOperationCount: model.MaxWriteRequestOperationCount,
		MaxWriteRequestSize:           model.MaxWriteRequestSize,
		ReadPollTimeout:               model.ReadPollTimeout,
	}
	diags.Append(ccr.ApplyToPutAutoFollowRequest(tuning, req)...)
	if diags.HasError() {
		return nil, diags
	}

	return req, diags
}

func mapAutoFollowPatternToModel(ctx context.Context, summary *estypes.AutoFollowPatternSummary, prior Model) (Model, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := prior
	model.Active = types.BoolValue(summary.Active)
	model.RemoteCluster = types.StringValue(summary.RemoteCluster)
	model.LeaderIndexPatterns = typeutils.ListValueFrom(
		ctx,
		summary.LeaderIndexPatterns,
		types.StringType,
		path.Root("leader_index_patterns"),
		&diags,
	)
	model.LeaderIndexExclusionPatterns = typeutils.ListValueFrom(
		ctx,
		summary.LeaderIndexExclusionPatterns,
		types.StringType,
		path.Root("leader_index_exclusion_patterns"),
		&diags,
	)
	model.FollowIndexPattern = types.StringPointerValue(summary.FollowIndexPattern)
	// max_outstanding_read_requests is the only tuning parameter the auto-follow
	// GET API returns, so it is mapped from the API response (Computed). The
	// remaining tuning parameters are write-only and absent from the summary, so
	// prior state is preserved for them.
	model.MaxOutstandingReadRequests = types.Int64Value(int64(summary.MaxOutstandingReadRequests))
	model.MaxOutstandingWriteRequests = prior.MaxOutstandingWriteRequests
	model.MaxReadRequestOperationCount = prior.MaxReadRequestOperationCount
	model.MaxReadRequestSize = prior.MaxReadRequestSize
	model.MaxRetryDelay = prior.MaxRetryDelay
	model.MaxWriteBufferCount = prior.MaxWriteBufferCount
	model.MaxWriteBufferSize = prior.MaxWriteBufferSize
	model.MaxWriteRequestOperationCount = prior.MaxWriteRequestOperationCount
	model.MaxWriteRequestSize = prior.MaxWriteRequestSize
	model.ReadPollTimeout = prior.ReadPollTimeout
	model.SettingsRaw = prior.SettingsRaw

	return model, diags
}
