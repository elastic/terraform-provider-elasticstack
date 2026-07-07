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
	"context"
	"encoding/json"
	"testing"

	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/followerindexstatus"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSchema_statusValidation(t *testing.T) {
	t.Parallel()

	s := getSchema(context.Background())
	statusAttr, ok := s.Attributes["status"].(schema.StringAttribute)
	require.True(t, ok)
	require.NotEmpty(t, statusAttr.Validators)

	assertStringValidatorAccepts(t, statusAttr.Validators, statusActive)
	assertStringValidatorAccepts(t, statusAttr.Validators, statusPaused)
	assertStringValidatorRejects(t, statusAttr.Validators, "invalid")
}

func TestSelectUpdateBranch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		prior, plan string
		want        updateBranch
	}{
		{statusActive, statusActive, branchActiveActive},
		{statusActive, statusPaused, branchActivePaused},
		{statusPaused, statusActive, branchPausedActive},
		{statusPaused, statusPaused, branchPausedPaused},
	}

	for _, tc := range tests {
		t.Run(tc.prior+"->"+tc.plan, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, selectUpdateBranch(tc.prior, tc.plan))
		})
	}
}

func TestPlanUpdateOperations_allBranches(t *testing.T) {
	t.Parallel()

	basePrior := Model{
		Status:                     types.StringValue(statusActive),
		MaxOutstandingReadRequests: types.Int64Value(12),
		SettingsRaw:                jsontypes.NewNormalizedValue(`{"index.refresh_interval":"30s"}`),
	}
	basePlan := basePrior

	t.Run("active to active tuning change", func(t *testing.T) {
		t.Parallel()
		prior := basePrior
		plan := basePrior
		plan.MaxOutstandingReadRequests = types.Int64Value(24)
		assert.Equal(t, []apiOperation{opPause, opResume}, planUpdateOperations(prior, plan))
	})

	t.Run("active to active settings change", func(t *testing.T) {
		t.Parallel()
		prior := basePrior
		plan := basePrior
		plan.SettingsRaw = jsontypes.NewNormalizedValue(`{"index.refresh_interval":"60s"}`)
		assert.Equal(t, []apiOperation{opPause, opUpdateSettings, opResume}, planUpdateOperations(prior, plan))
	})

	t.Run("active to active no material change", func(t *testing.T) {
		t.Parallel()
		prior := basePrior
		plan := basePlan
		plan.DeleteIndexOnDestroy = types.BoolValue(true)
		assert.Empty(t, planUpdateOperations(prior, plan))
	})

	t.Run("active to paused", func(t *testing.T) {
		t.Parallel()
		prior := basePrior
		plan := basePrior
		plan.Status = types.StringValue(statusPaused)
		assert.Equal(t, []apiOperation{opPause}, planUpdateOperations(prior, plan))
	})

	t.Run("paused to active", func(t *testing.T) {
		t.Parallel()
		prior := basePrior
		prior.Status = types.StringValue(statusPaused)
		plan := prior
		plan.Status = types.StringValue(statusActive)
		assert.Equal(t, []apiOperation{opResume}, planUpdateOperations(prior, plan))
	})

	t.Run("paused to active with settings change", func(t *testing.T) {
		t.Parallel()
		prior := basePrior
		prior.Status = types.StringValue(statusPaused)
		plan := prior
		plan.Status = types.StringValue(statusActive)
		plan.SettingsRaw = jsontypes.NewNormalizedValue(`{"index.refresh_interval":"60s"}`)
		assert.Equal(t, []apiOperation{opUpdateSettings, opResume}, planUpdateOperations(prior, plan))
	})

	t.Run("paused to paused tuning deferred", func(t *testing.T) {
		t.Parallel()
		prior := basePrior
		prior.Status = types.StringValue(statusPaused)
		plan := prior
		plan.MaxOutstandingReadRequests = types.Int64Value(99)
		assert.Empty(t, planUpdateOperations(prior, plan))
	})

	t.Run("active to active tuning only does not update settings", func(t *testing.T) {
		t.Parallel()
		prior := basePrior
		plan := basePrior
		plan.MaxOutstandingReadRequests = types.Int64Value(24)
		ops := planUpdateOperations(prior, plan)
		assert.Equal(t, []apiOperation{opPause, opResume}, ops)
		assert.NotContains(t, ops, opUpdateSettings)
	})

	t.Run("active to active settings removal does not pause", func(t *testing.T) {
		t.Parallel()
		prior := basePrior
		plan := basePrior
		plan.SettingsRaw = jsontypes.NewNormalizedNull()
		assert.Empty(t, planUpdateOperations(prior, plan))
	})
}

func TestPlanDeleteOperations(t *testing.T) {
	t.Parallel()

	t.Run("active prior state", func(t *testing.T) {
		t.Parallel()
		prior := Model{Status: types.StringValue(statusActive), DeleteIndexOnDestroy: types.BoolValue(false)}
		assert.Equal(t,
			[]apiOperation{opPause, opClose, opUnfollow, opOpenIndex},
			planDeleteOperations(prior),
		)
	})

	t.Run("paused prior state skips pause", func(t *testing.T) {
		t.Parallel()
		prior := Model{Status: types.StringValue(statusPaused), DeleteIndexOnDestroy: types.BoolValue(false)}
		assert.Equal(t,
			[]apiOperation{opClose, opUnfollow, opOpenIndex},
			planDeleteOperations(prior),
		)
	})

	t.Run("delete index on destroy true", func(t *testing.T) {
		t.Parallel()
		prior := Model{Status: types.StringValue(statusActive), DeleteIndexOnDestroy: types.BoolValue(true)}
		assert.Equal(t,
			[]apiOperation{opPause, opClose, opUnfollow, opDeleteIndex},
			planDeleteOperations(prior),
		)
	})

	t.Run("delete index on destroy false ends with open", func(t *testing.T) {
		t.Parallel()
		prior := Model{Status: types.StringValue(statusPaused), DeleteIndexOnDestroy: types.BoolValue(false)}
		ops := planDeleteOperations(prior)
		require.NotEmpty(t, ops)
		assert.Equal(t, opOpenIndex, ops[len(ops)-1])
	})
}

func TestNormalizeFlatSettingsKeys(t *testing.T) {
	t.Parallel()

	t.Run("flat dotted keys", func(t *testing.T) {
		t.Parallel()
		in := map[string]any{"index.refresh_interval": "30s"}
		out, _ := normalizeFlatSettingsKeys(in)
		assert.Equal(t, map[string]any{
			"index": map[string]any{"refresh_interval": "30s"},
		}, out)
	})

	t.Run("nested keys unchanged", func(t *testing.T) {
		t.Parallel()
		in := map[string]any{"index": map[string]any{"refresh_interval": "30s"}}
		out, _ := normalizeFlatSettingsKeys(in)
		assert.Equal(t, in, out)
	})
}

func TestParseSettingsRawForCreate_formats(t *testing.T) {
	t.Parallel()

	t.Run("flat format unmarshals into IndexSettings", func(t *testing.T) {
		t.Parallel()
		settings, diags := parseSettingsRawForCreate(`{"index.refresh_interval":"30s"}`)
		require.False(t, diags.HasError(), diags)
		require.NotNil(t, settings)

		bytes, err := json.Marshal(settings)
		require.NoError(t, err)
		assert.Contains(t, string(bytes), "refresh_interval")
	})

	t.Run("nested format unmarshals into IndexSettings", func(t *testing.T) {
		t.Parallel()
		settings, diags := parseSettingsRawForCreate(`{"index":{"refresh_interval":"30s"}}`)
		require.False(t, diags.HasError(), diags)
		require.NotNil(t, settings)

		bytes, err := json.Marshal(settings)
		require.NoError(t, err)
		assert.Contains(t, string(bytes), "refresh_interval")
	})

	t.Run("invalid json rejected", func(t *testing.T) {
		t.Parallel()
		_, diags := parseSettingsRawForCreate("not-valid-json")
		require.True(t, diags.HasError())
	})
}

func TestBuildFollowRequest_byteSizeAndDuration(t *testing.T) {
	t.Parallel()

	model := Model{
		Name:                        types.StringValue("follower"),
		RemoteCluster:               types.StringValue("dc2"),
		LeaderIndex:                 types.StringValue("leader"),
		MaxReadRequestSize:          types.StringValue("100mb"),
		MaxRetryDelay:               customtypes.NewDurationValue("10s"),
		ReadPollTimeout:             customtypes.NewDurationValue("10m"),
		MaxOutstandingReadRequests:  types.Int64Value(12),
		MaxOutstandingWriteRequests: types.Int64Value(8),
	}

	req, diags := buildFollowRequest(model)
	require.False(t, diags.HasError(), diags)
	require.NotNil(t, req)

	assert.Equal(t, estypes.ByteSize("100mb"), req.MaxReadRequestSize)
	assert.Equal(t, estypes.Duration("10s"), req.MaxRetryDelay)
	assert.Equal(t, estypes.Duration("10m"), req.ReadPollTimeout)
	require.NotNil(t, req.MaxOutstandingReadRequests)
	assert.Equal(t, int64(12), *req.MaxOutstandingReadRequests)
	require.NotNil(t, req.MaxOutstandingWriteRequests)
	assert.Equal(t, 8, *req.MaxOutstandingWriteRequests)
}

func TestBuildResumeFollowRequest_typeMapping(t *testing.T) {
	t.Parallel()

	model := Model{
		MaxOutstandingReadRequests:    types.Int64Value(12),
		MaxOutstandingWriteRequests:   types.Int64Value(8),
		MaxReadRequestOperationCount:  types.Int64Value(4),
		MaxReadRequestSize:            types.StringValue("100mb"),
		MaxRetryDelay:                 customtypes.NewDurationValue("10s"),
		MaxWriteBufferCount:           types.Int64Value(16),
		MaxWriteBufferSize:            types.StringValue("200mb"),
		MaxWriteRequestOperationCount: types.Int64Value(32),
		MaxWriteRequestSize:           types.StringValue("64mb"),
		ReadPollTimeout:               customtypes.NewDurationValue("10m"),
	}

	req := buildResumeFollowRequest(model)
	require.NotNil(t, req)

	// resumefollow represents all counts as *int64 and byte sizes as *string.
	require.NotNil(t, req.MaxOutstandingReadRequests)
	assert.Equal(t, int64(12), *req.MaxOutstandingReadRequests)
	require.NotNil(t, req.MaxOutstandingWriteRequests)
	assert.Equal(t, int64(8), *req.MaxOutstandingWriteRequests)
	require.NotNil(t, req.MaxReadRequestSize)
	assert.Equal(t, "100mb", *req.MaxReadRequestSize)
	require.NotNil(t, req.MaxWriteRequestSize)
	assert.Equal(t, "64mb", *req.MaxWriteRequestSize)
	assert.Equal(t, estypes.Duration("10s"), req.MaxRetryDelay)
	assert.Equal(t, estypes.Duration("10m"), req.ReadPollTimeout)
}
func TestMapFollowerIndexToModel_preservesTuningWhenPaused(t *testing.T) {
	t.Parallel()

	prior := Model{
		MaxOutstandingReadRequests: types.Int64Value(12),
		MaxReadRequestSize:         types.StringValue("100mb"),
		SettingsRaw:                jsontypes.NewNormalizedValue(`{"index.refresh_interval":"30s"}`),
		DataStreamName:             types.StringValue("logs"),
	}

	follower := &estypes.FollowerIndex{
		FollowerIndex: "follower",
		LeaderIndex:   "leader",
		RemoteCluster: "dc2",
		Status:        mustFollowerStatus(statusPaused),
		Parameters:    nil,
	}

	model := mapFollowerIndexToModel(follower, prior)
	assert.Equal(t, types.Int64Value(12), model.MaxOutstandingReadRequests)
	assert.Equal(t, types.StringValue("100mb"), model.MaxReadRequestSize)
	assert.Equal(t, jsontypes.NewNormalizedValue(`{"index.refresh_interval":"30s"}`), model.SettingsRaw)
	assert.Equal(t, types.StringValue("logs"), model.DataStreamName)
	assert.Equal(t, types.StringValue(statusPaused), model.Status)
}

func mustFollowerStatus(s string) followerindexstatus.FollowerIndexStatus {
	var status followerindexstatus.FollowerIndexStatus
	if err := status.UnmarshalText([]byte(s)); err != nil {
		panic(err)
	}
	return status
}

func assertStringValidatorAccepts(t *testing.T, validators []validator.String, value string) {
	t.Helper()
	req := validator.StringRequest{ConfigValue: types.StringValue(value)}
	for _, v := range validators {
		var resp validator.StringResponse
		v.ValidateString(context.Background(), req, &resp)
		require.False(t, resp.Diagnostics.HasError(), "expected %q to be accepted: %v", value, resp.Diagnostics)
	}
}

func assertStringValidatorRejects(t *testing.T, validators []validator.String, value string) {
	t.Helper()
	req := validator.StringRequest{ConfigValue: types.StringValue(value)}
	var combined diag.Diagnostics
	for _, v := range validators {
		var resp validator.StringResponse
		v.ValidateString(context.Background(), req, &resp)
		combined.Append(resp.Diagnostics...)
	}
	require.True(t, combined.HasError(), "expected %q to be rejected", value)
}
