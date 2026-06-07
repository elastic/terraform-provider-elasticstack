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

func parseSettingsRaw(settingsRaw string) (map[string]json.RawMessage, diag.Diagnostics) {
	var settings map[string]json.RawMessage
	if err := json.Unmarshal([]byte(settingsRaw), &settings); err != nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic("Failed to parse settings_raw", err.Error()),
		}
	}
	return settings, nil
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
		FollowIndexPattern:  typeutils.OptStringPtr(model.FollowIndexPattern),
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

	if v, d := ccr.OptIntFromInt64("max_outstanding_read_requests", model.MaxOutstandingReadRequests); d.HasError() {
		diags.Append(d...)
	} else if v != nil {
		req.MaxOutstandingReadRequests = v
	}
	if v, d := ccr.OptIntFromInt64("max_outstanding_write_requests", model.MaxOutstandingWriteRequests); d.HasError() {
		diags.Append(d...)
	} else if v != nil {
		req.MaxOutstandingWriteRequests = v
	}
	if v, d := ccr.OptIntFromInt64("max_read_request_operation_count", model.MaxReadRequestOperationCount); d.HasError() {
		diags.Append(d...)
	} else if v != nil {
		req.MaxReadRequestOperationCount = v
	}
	if v := ccr.ByteSizeFromString(model.MaxReadRequestSize); v != nil {
		req.MaxReadRequestSize = v
	}
	if v := ccr.DurationFromCustomType(model.MaxRetryDelay); v != nil {
		req.MaxRetryDelay = v
	}
	if v, d := ccr.OptIntFromInt64("max_write_buffer_count", model.MaxWriteBufferCount); d.HasError() {
		diags.Append(d...)
	} else if v != nil {
		req.MaxWriteBufferCount = v
	}
	if v := ccr.ByteSizeFromString(model.MaxWriteBufferSize); v != nil {
		req.MaxWriteBufferSize = v
	}
	if v, d := ccr.OptIntFromInt64("max_write_request_operation_count", model.MaxWriteRequestOperationCount); d.HasError() {
		diags.Append(d...)
	} else if v != nil {
		req.MaxWriteRequestOperationCount = v
	}
	if v := ccr.ByteSizeFromString(model.MaxWriteRequestSize); v != nil {
		req.MaxWriteRequestSize = v
	}
	if v := ccr.DurationFromCustomType(model.ReadPollTimeout); v != nil {
		req.ReadPollTimeout = v
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
	// Preserve prior state for write-only tuning parameters; the auto-follow
	// GET API does not return them so summary carries zero values.
	model.MaxOutstandingReadRequests = prior.MaxOutstandingReadRequests
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
