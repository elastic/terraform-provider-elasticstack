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
	"sort"
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
		return fmt.Sprintf("apiOperation(%d)", op)
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

// normalizeFlatSettingsKeys converts flat dotted keys (e.g. index.refresh_interval)
// into nested maps for unmarshalling into types.IndexSettings.
// The bool return indicates whether any normalization was performed.
func normalizeFlatSettingsKeys(m map[string]any) (map[string]any, bool) {
	hasDotted := false
	for k := range m {
		if strings.Contains(k, ".") {
			hasDotted = true
			break
		}
	}
	if !hasDotted {
		return m, false
	}

	flat := make(map[string]any)
	root := make(map[string]any)
	for k, v := range m {
		if strings.Contains(k, ".") {
			flat[k] = v
		} else {
			root[k] = v
		}
	}

	unflattened := unflattenDottedMap(flat)
	return mergeSettingsMaps(root, unflattened), true
}

func unflattenDottedMap(flat map[string]any) map[string]any {
	root := make(map[string]any)
	keys := make([]string, 0, len(flat))
	for k := range flat {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := flat[k]
		parts := strings.Split(k, ".")
		cur := root
		for i := range parts {
			p := parts[i]
			if i == len(parts)-1 {
				cur[p] = v
				break
			}
			existing, ok := cur[p]
			if !ok {
				nm := make(map[string]any)
				cur[p] = nm
				cur = nm
				continue
			}
			nm, ok := existing.(map[string]any)
			if !ok {
				nm = make(map[string]any)
				cur[p] = nm
			}
			cur = nm
		}
	}
	return root
}

func mergeSettingsMaps(base, overlay map[string]any) map[string]any {
	if len(base) == 0 {
		return overlay
	}
	if len(overlay) == 0 {
		return base
	}
	out := make(map[string]any, len(base)+len(overlay))
	maps.Copy(out, base)
	for k, v := range overlay {
		existing, ok := out[k]
		if !ok {
			out[k] = v
			continue
		}
		baseMap, baseOK := existing.(map[string]any)
		overlayMap, overlayOK := v.(map[string]any)
		if baseOK && overlayOK {
			out[k] = mergeSettingsMaps(baseMap, overlayMap)
			continue
		}
		out[k] = v
	}
	return out
}

func parseSettingsRawForCreate(settingsRaw string) (*estypes.IndexSettings, diag.Diagnostics) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(settingsRaw), &raw); err != nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic("Failed to parse settings_raw", err.Error()),
		}
	}

	normalized, changed := normalizeFlatSettingsKeys(raw)
	if !changed {
		var settings estypes.IndexSettings
		if err := json.Unmarshal([]byte(settingsRaw), &settings); err != nil {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic("Failed to parse settings_raw into index settings", err.Error()),
			}
		}
		return &settings, nil
	}

	settingsBytes, err := json.Marshal(normalized)
	if err != nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic("Failed to marshal normalized settings", err.Error()),
		}
	}

	var settings estypes.IndexSettings
	if err := json.Unmarshal(settingsBytes, &settings); err != nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic("Failed to parse settings_raw into index settings", err.Error()),
		}
	}

	return &settings, nil
}

func parseSettingsRawForUpdate(settingsRaw string) (map[string]any, diag.Diagnostics) {
	var settings map[string]any
	if err := json.Unmarshal([]byte(settingsRaw), &settings); err != nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic("Failed to parse settings_raw", err.Error()),
		}
	}
	return settings, nil
}

func buildFollowRequest(model Model) (*follow.Request, diag.Diagnostics) {
	req := &follow.Request{
		LeaderIndex:    model.LeaderIndex.ValueString(),
		RemoteCluster:  model.RemoteCluster.ValueString(),
		DataStreamName: typeutils.OptStringPtr(model.DataStreamName),
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

	if v := ccr.OptInt64Ptr(model.MaxOutstandingReadRequests); v != nil {
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
	if v := ccr.DurationFromString(model.MaxRetryDelay); v != nil {
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
	if v := ccr.DurationFromString(model.ReadPollTimeout); v != nil {
		req.ReadPollTimeout = v
	}

	return req, diags
}

func buildResumeFollowRequest(model Model) *resumefollow.Request {
	req := &resumefollow.Request{}

	if v := ccr.OptInt64Ptr(model.MaxOutstandingReadRequests); v != nil {
		req.MaxOutstandingReadRequests = v
	}
	if v := ccr.OptInt64Ptr(model.MaxOutstandingWriteRequests); v != nil {
		req.MaxOutstandingWriteRequests = v
	}
	if v := ccr.OptInt64Ptr(model.MaxReadRequestOperationCount); v != nil {
		req.MaxReadRequestOperationCount = v
	}
	if v := typeutils.OptStringPtr(model.MaxReadRequestSize); v != nil {
		req.MaxReadRequestSize = v
	}
	if v := ccr.DurationFromString(model.MaxRetryDelay); v != nil {
		req.MaxRetryDelay = v
	}
	if v := ccr.OptInt64Ptr(model.MaxWriteBufferCount); v != nil {
		req.MaxWriteBufferCount = v
	}
	if v := typeutils.OptStringPtr(model.MaxWriteBufferSize); v != nil {
		req.MaxWriteBufferSize = v
	}
	if v := ccr.OptInt64Ptr(model.MaxWriteRequestOperationCount); v != nil {
		req.MaxWriteRequestOperationCount = v
	}
	if v := typeutils.OptStringPtr(model.MaxWriteRequestSize); v != nil {
		req.MaxWriteRequestSize = v
	}
	if v := ccr.DurationFromString(model.ReadPollTimeout); v != nil {
		req.ReadPollTimeout = v
	}

	return req
}

func intPointerToInt64(v *int) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*v))
}

func mapParametersToModel(params *estypes.FollowerIndexParameters, model Model) Model {
	model.MaxOutstandingReadRequests = typeutils.Int64PointerValue(params.MaxOutstandingReadRequests)
	model.MaxOutstandingWriteRequests = intPointerToInt64(params.MaxOutstandingWriteRequests)
	model.MaxReadRequestOperationCount = intPointerToInt64(params.MaxReadRequestOperationCount)
	model.MaxReadRequestSize = ccr.ByteSizeToString(params.MaxReadRequestSize)
	model.MaxRetryDelay = ccr.DurationToString(params.MaxRetryDelay)
	model.MaxWriteBufferCount = intPointerToInt64(params.MaxWriteBufferCount)
	model.MaxWriteBufferSize = ccr.ByteSizeToString(params.MaxWriteBufferSize)
	model.MaxWriteRequestOperationCount = intPointerToInt64(params.MaxWriteRequestOperationCount)
	model.MaxWriteRequestSize = ccr.ByteSizeToString(params.MaxWriteRequestSize)
	model.ReadPollTimeout = ccr.DurationToString(params.ReadPollTimeout)
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
