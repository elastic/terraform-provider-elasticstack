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

package followerindex

import (
	"encoding/json"
	"fmt"
	"maps"
	"strings"

	"github.com/elastic/go-elasticsearch/v8/typedapi/ccr/follow"
	"github.com/elastic/go-elasticsearch/v8/typedapi/ccr/resumefollow"
	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ccr"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// apiOperation identifies a CCR lifecycle API call for testable sequencing.
type apiOperation int

const (
	opPause apiOperation = iota
	opUpdateSettings
	opResume
	opClose
	opUnfollow
	opDeleteIndex
	opOpenIndex
)

func (op apiOperation) String() string {
	switch op {
	case opPause:
		return "PauseFollowerIndex"
	case opUpdateSettings:
		return "UpdateIndexSettings"
	case opResume:
		return "ResumeFollowerIndex"
	case opClose:
		return "CloseIndex"
	case opUnfollow:
		return "UnfollowIndex"
	case opDeleteIndex:
		return "DeleteIndex"
	case opOpenIndex:
		return "OpenIndex"
	default:
		return ccr.FormatUnknownOperation(int(op))
	}
}

// updateBranch classifies the status transition for the update state machine.
type updateBranch int

const (
	branchActiveActive updateBranch = iota
	branchActivePaused
	branchPausedActive
	branchPausedPaused
)

func selectUpdateBranch(priorStatus, planStatus string) updateBranch {
	switch {
	case priorStatus == statusActive && planStatus == statusActive:
		return branchActiveActive
	case priorStatus == statusActive && planStatus == statusPaused:
		return branchActivePaused
	case priorStatus == statusPaused && planStatus == statusActive:
		return branchPausedActive
	default:
		return branchPausedPaused
	}
}

func planUpdateOperations(prior, plan Model) []apiOperation {
	branch := selectUpdateBranch(prior.Status.ValueString(), plan.Status.ValueString())
	settingsChanged := settingsRawChanged(prior, plan)

	switch branch {
	case branchActiveActive:
		if !tuningParamsChanged(prior, plan) && !settingsChanged {
			return nil
		}
		ops := []apiOperation{opPause}
		if settingsChanged {
			ops = append(ops, opUpdateSettings)
		}
		return append(ops, opResume)
	case branchActivePaused:
		return []apiOperation{opPause}
	case branchPausedActive:
		ops := make([]apiOperation, 0, 2)
		if settingsChanged {
			ops = append(ops, opUpdateSettings)
		}
		return append(ops, opResume)
	case branchPausedPaused:
		return nil
	default:
		return nil
	}
}

func planDeleteOperations(prior Model) []apiOperation {
	ops := make([]apiOperation, 0, 5)
	if prior.Status.ValueString() == statusActive {
		ops = append(ops, opPause)
	}
	ops = append(ops, opClose, opUnfollow)
	if prior.DeleteIndexOnDestroy.ValueBool() {
		ops = append(ops, opDeleteIndex)
	} else {
		ops = append(ops, opOpenIndex)
	}
	return ops
}

func settingsRawChanged(prior, plan Model) bool {
	if !typeutils.IsKnown(plan.SettingsRaw) {
		return false
	}
	if !typeutils.IsKnown(prior.SettingsRaw) {
		return true
	}
	return !prior.SettingsRaw.Equal(plan.SettingsRaw)
}

func tuningParamsChanged(prior, plan Model) bool {
	return !prior.MaxOutstandingReadRequests.Equal(plan.MaxOutstandingReadRequests) ||
		!prior.MaxOutstandingWriteRequests.Equal(plan.MaxOutstandingWriteRequests) ||
		!prior.MaxReadRequestOperationCount.Equal(plan.MaxReadRequestOperationCount) ||
		!prior.MaxReadRequestSize.Equal(plan.MaxReadRequestSize) ||
		!prior.MaxRetryDelay.Equal(plan.MaxRetryDelay) ||
		!prior.MaxWriteBufferCount.Equal(plan.MaxWriteBufferCount) ||
		!prior.MaxWriteBufferSize.Equal(plan.MaxWriteBufferSize) ||
		!prior.MaxWriteRequestOperationCount.Equal(plan.MaxWriteRequestOperationCount) ||
		!prior.MaxWriteRequestSize.Equal(plan.MaxWriteRequestSize) ||
		!prior.ReadPollTimeout.Equal(plan.ReadPollTimeout)
}

func parseSettingsRawForCreate(settingsRaw string) (*estypes.IndexSettings, diag.Diagnostics) {
	raw, diags := typeutils.UnmarshalJSONDiag[map[string]any](settingsRaw, "Failed to parse settings_raw")
	if diags.HasError() {
		return nil, diags
	}

	normalized, changed := ccr.NormalizeFlatSettingsKeys(raw)
	if !changed {
		settings, settingsDiags := typeutils.UnmarshalJSONDiag[estypes.IndexSettings](settingsRaw, "Failed to parse settings_raw into index settings")
		if settingsDiags.HasError() {
			return nil, settingsDiags
		}
		return &settings, nil
	}

	settingsBytes, err := json.Marshal(normalized)
	if err != nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic("Failed to marshal normalized settings", err.Error()),
		}
	}

	settings, settingsDiags := typeutils.UnmarshalJSONDiag[estypes.IndexSettings](string(settingsBytes), "Failed to parse settings_raw into index settings")
	if settingsDiags.HasError() {
		return nil, settingsDiags
	}
	return &settings, nil
}

func parseSettingsRawForUpdate(settingsRaw string) (map[string]any, diag.Diagnostics) {
	return typeutils.UnmarshalJSONDiag[map[string]any](settingsRaw, "Failed to parse settings_raw")
}

func buildFollowRequest(model Model) (*follow.Request, diag.Diagnostics) {
	req := &follow.Request{
		LeaderIndex:    model.LeaderIndex.ValueString(),
		RemoteCluster:  model.RemoteCluster.ValueString(),
		DataStreamName: typeutils.OptionalString(model.DataStreamName),
	}

	var diags diag.Diagnostics

	if typeutils.IsKnown(model.SettingsRaw) {
		settings, settingsDiags := parseSettingsRawForCreate(model.SettingsRaw.ValueString())
		diags.Append(settingsDiags...)
		if diags.HasError() {
			return nil, diags
		}
		req.Settings = settings
	}

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
	diags.Append(ccr.ApplyToFollowRequest(tuning, req)...)
	if diags.HasError() {
		return nil, diags
	}

	return req, diags
}

func buildResumeFollowRequest(model Model) *resumefollow.Request {
	req := &resumefollow.Request{}
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
	ccr.ApplyToResumeFollowRequest(tuning, req)
	return req
}

func mapParametersToModel(params *estypes.FollowerIndexParameters, model Model) Model {
	p := ccr.TuningParamsFromParameters(params)
	model.MaxOutstandingReadRequests = p.MaxOutstandingReadRequests
	model.MaxOutstandingWriteRequests = p.MaxOutstandingWriteRequests
	model.MaxReadRequestOperationCount = p.MaxReadRequestOperationCount
	model.MaxReadRequestSize = p.MaxReadRequestSize
	model.MaxRetryDelay = p.MaxRetryDelay
	model.MaxWriteBufferCount = p.MaxWriteBufferCount
	model.MaxWriteBufferSize = p.MaxWriteBufferSize
	model.MaxWriteRequestOperationCount = p.MaxWriteRequestOperationCount
	model.MaxWriteRequestSize = p.MaxWriteRequestSize
	model.ReadPollTimeout = p.ReadPollTimeout
	return model
}

func mapFollowerIndexToModel(follower *estypes.FollowerIndex, prior Model) Model {
	model := prior
	model.RemoteCluster = types.StringValue(follower.RemoteCluster)
	model.LeaderIndex = types.StringValue(follower.LeaderIndex)
	model.Status = types.StringValue(follower.Status.String())

	if follower.Parameters != nil {
		model = mapParametersToModel(follower.Parameters, model)
	}

	// delete_index_on_destroy is a local-only attribute that is never returned
	// by the API. On import the baseline carries no value, so default it to
	// false to satisfy the documented post-import state.
	if !typeutils.IsKnown(model.DeleteIndexOnDestroy) {
		model.DeleteIndexOnDestroy = types.BoolValue(false)
	}

	return model
}
