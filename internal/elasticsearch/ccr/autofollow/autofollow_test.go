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
	"testing"

	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSchema_leaderIndexPatternsValidation(t *testing.T) {
	t.Parallel()

	s := getSchema(context.Background())
	patternsAttr, ok := s.Attributes["leader_index_patterns"].(schema.ListAttribute)
	require.True(t, ok)
	require.NotEmpty(t, patternsAttr.Validators)

	assertListValidatorRejectsEmpty(t, patternsAttr.Validators)
	assertListValidatorAcceptsNonempty(t, patternsAttr.Validators)
}

func TestSelectUpdateActiveBranch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		priorActive, planActive bool
		want                    updateActiveBranch
	}{
		{true, true, branchActiveUnchanged},
		{true, false, branchActiveToInactive},
		{false, true, branchInactiveToActive},
		{false, false, branchActiveUnchanged},
	}

	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, selectUpdateActiveBranch(tc.priorActive, tc.planActive))
		})
	}
}

func TestPlanCreateOperations(t *testing.T) {
	t.Parallel()

	t.Run("active true puts only", func(t *testing.T) {
		t.Parallel()
		plan := Model{Active: types.BoolValue(true)}
		assert.Equal(t, []apiOperation{opPut}, planCreateOperations(plan))
	})

	t.Run("active false puts then pauses", func(t *testing.T) {
		t.Parallel()
		plan := Model{Active: types.BoolValue(false)}
		assert.Equal(t, []apiOperation{opPut, opPause}, planCreateOperations(plan))
	})
}

func TestPlanUpdateOperations_allBranches(t *testing.T) {
	t.Parallel()

	basePrior := Model{Active: types.BoolValue(true)}
	basePlan := basePrior

	t.Run("active unchanged puts only", func(t *testing.T) {
		t.Parallel()
		prior := basePrior
		plan := basePlan
		plan.RemoteCluster = types.StringValue("dc3")
		assert.Equal(t, []apiOperation{opPut}, planUpdateOperations(prior, plan))
	})

	t.Run("active true to false puts then pauses", func(t *testing.T) {
		t.Parallel()
		prior := basePrior
		plan := basePrior
		plan.Active = types.BoolValue(false)
		assert.Equal(t, []apiOperation{opPut, opPause}, planUpdateOperations(prior, plan))
	})

	t.Run("active false to true puts then resumes", func(t *testing.T) {
		t.Parallel()
		prior := basePrior
		prior.Active = types.BoolValue(false)
		plan := prior
		plan.Active = types.BoolValue(true)
		assert.Equal(t, []apiOperation{opPut, opResume}, planUpdateOperations(prior, plan))
	})

	t.Run("active false unchanged puts only", func(t *testing.T) {
		t.Parallel()
		prior := basePrior
		prior.Active = types.BoolValue(false)
		plan := prior
		plan.LeaderIndexPatterns = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("metrics-*")})
		assert.Equal(t, []apiOperation{opPut}, planUpdateOperations(prior, plan))
	})
}

func TestBuildPutAutoFollowPatternRequest_byteSizeAndDuration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := Model{
		Name:                        types.StringValue("etl-logs"),
		RemoteCluster:               types.StringValue("dc2"),
		LeaderIndexPatterns:         types.ListValueMust(types.StringType, []attr.Value{types.StringValue("logs-*")}),
		MaxReadRequestSize:          types.StringValue("100mb"),
		MaxRetryDelay:               customtypes.NewDurationValue("10s"),
		ReadPollTimeout:             customtypes.NewDurationValue("10m"),
		MaxOutstandingReadRequests:  types.Int64Value(12),
		MaxOutstandingWriteRequests: types.Int64Value(8),
	}

	req, diags := buildPutAutoFollowPatternRequest(ctx, model)
	require.False(t, diags.HasError(), diags)
	require.NotNil(t, req)

	assert.Equal(t, estypes.ByteSize("100mb"), req.MaxReadRequestSize)
	assert.Equal(t, estypes.Duration("10s"), req.MaxRetryDelay)
	assert.Equal(t, estypes.Duration("10m"), req.ReadPollTimeout)
	require.NotNil(t, req.MaxOutstandingReadRequests)
	assert.Equal(t, 12, *req.MaxOutstandingReadRequests)
	require.NotNil(t, req.MaxOutstandingWriteRequests)
	assert.Equal(t, 8, *req.MaxOutstandingWriteRequests)
}

func TestBuildPutAutoFollowPatternRequest_settingsRaw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := Model{
		RemoteCluster:       types.StringValue("dc2"),
		LeaderIndexPatterns: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("logs-*")}),
		SettingsRaw:         jsontypes.NewNormalizedValue(`{"index.refresh_interval":"30s"}`),
	}

	req, diags := buildPutAutoFollowPatternRequest(ctx, model)
	require.False(t, diags.HasError(), diags)
	require.NotNil(t, req.Settings)
	assert.Contains(t, req.Settings, "index.refresh_interval")

	raw := req.Settings["index.refresh_interval"]
	var interval string
	require.NoError(t, json.Unmarshal(raw, &interval))
	assert.Equal(t, "30s", interval)
}

func TestParseSettingsRaw_invalidJSON(t *testing.T) {
	t.Parallel()

	_, diags := parseSettingsRaw("not-valid-json")
	require.True(t, diags.HasError())
}
func TestMapAutoFollowPatternToModel_mapsAPIAndPreservesUnreadableTuning(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	prior := Model{
		MaxOutstandingReadRequests:    types.Int64Value(99),
		MaxOutstandingWriteRequests:   types.Int64Value(7),
		MaxReadRequestOperationCount:  types.Int64Value(512),
		MaxReadRequestSize:            types.StringValue("50mb"),
		MaxRetryDelay:                 customtypes.NewDurationValue("30s"),
		MaxWriteBufferCount:           types.Int64Value(100),
		MaxWriteBufferSize:            types.StringValue("200mb"),
		MaxWriteRequestOperationCount: types.Int64Value(256),
		MaxWriteRequestSize:           types.StringValue("150mb"),
		ReadPollTimeout:               customtypes.NewDurationValue("5m"),
		SettingsRaw:                   jsontypes.NewNormalizedValue(`{"index.refresh_interval":"30s"}`),
	}

	followPattern := "logs-{{leader_index}}-replica"
	summary := &estypes.AutoFollowPatternSummary{
		Active:                       true,
		RemoteCluster:                "dc2",
		LeaderIndexPatterns:          []string{"logs-*"},
		LeaderIndexExclusionPatterns: []string{"logs-debug-*"},
		FollowIndexPattern:           &followPattern,
		MaxOutstandingReadRequests:   10,
	}

	model, diags := mapAutoFollowPatternToModel(ctx, summary, prior)
	require.False(t, diags.HasError(), diags)

	assert.Equal(t, types.BoolValue(true), model.Active)
	assert.Equal(t, types.StringValue("dc2"), model.RemoteCluster)
	assert.Equal(t, types.StringValue(followPattern), model.FollowIndexPattern)
	// max_outstanding_read_requests is mapped from the API summary (the only
	// tuning parameter the GET API returns), not preserved from prior state.
	assert.Equal(t, types.Int64Value(10), model.MaxOutstandingReadRequests)
	assert.Equal(t, types.Int64Value(7), model.MaxOutstandingWriteRequests)
	assert.Equal(t, types.Int64Value(512), model.MaxReadRequestOperationCount)
	assert.Equal(t, types.StringValue("50mb"), model.MaxReadRequestSize)
	assert.Equal(t, customtypes.NewDurationValue("30s"), model.MaxRetryDelay)
	assert.Equal(t, types.Int64Value(100), model.MaxWriteBufferCount)
	assert.Equal(t, types.StringValue("200mb"), model.MaxWriteBufferSize)
	assert.Equal(t, types.Int64Value(256), model.MaxWriteRequestOperationCount)
	assert.Equal(t, types.StringValue("150mb"), model.MaxWriteRequestSize)
	assert.Equal(t, customtypes.NewDurationValue("5m"), model.ReadPollTimeout)
	assert.Equal(t, jsontypes.NewNormalizedValue(`{"index.refresh_interval":"30s"}`), model.SettingsRaw)

	var listDiags diag.Diagnostics
	patterns := typeutils.ListTypeToSliceString(ctx, model.LeaderIndexPatterns, path.Root("leader_index_patterns"), &listDiags)
	require.False(t, listDiags.HasError(), listDiags)
	assert.Equal(t, []string{"logs-*"}, patterns)

	exclusions := typeutils.ListTypeToSliceString(ctx, model.LeaderIndexExclusionPatterns, path.Root("leader_index_exclusion_patterns"), &listDiags)
	require.False(t, listDiags.HasError(), listDiags)
	assert.Equal(t, []string{"logs-debug-*"}, exclusions)
}

func assertListValidatorRejectsEmpty(t *testing.T, validators []validator.List) {
	t.Helper()
	ctx := context.Background()
	empty := types.ListValueMust(types.StringType, []attr.Value{})
	for _, v := range validators {
		var resp validator.ListResponse
		v.ValidateList(ctx, validator.ListRequest{
			Path:           path.Root("leader_index_patterns"),
			PathExpression: path.MatchRoot("leader_index_patterns"),
			ConfigValue:    empty,
		}, &resp)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	t.Fatal("expected empty leader_index_patterns to be rejected")
}

func assertListValidatorAcceptsNonempty(t *testing.T, validators []validator.List) {
	t.Helper()
	ctx := context.Background()
	list := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("logs-*")})
	for _, v := range validators {
		var resp validator.ListResponse
		v.ValidateList(ctx, validator.ListRequest{
			Path:           path.Root("leader_index_patterns"),
			PathExpression: path.MatchRoot("leader_index_patterns"),
			ConfigValue:    list,
		}, &resp)
		require.False(t, resp.Diagnostics.HasError(), "expected non-empty leader_index_patterns to be accepted: %v", resp.Diagnostics)
	}
}

func TestListValidatorSizeAtLeastOne(t *testing.T) {
	t.Parallel()

	v := listvalidator.SizeAtLeast(1)
	assertListValidatorRejectsEmpty(t, []validator.List{v})
	assertListValidatorAcceptsNonempty(t, []validator.List{v})
}
