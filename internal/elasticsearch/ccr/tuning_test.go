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

package ccr

import (
	"testing"

	"github.com/elastic/go-elasticsearch/v9/typedapi/ccr/follow"
	"github.com/elastic/go-elasticsearch/v9/typedapi/ccr/putautofollowpattern"
	"github.com/elastic/go-elasticsearch/v9/typedapi/ccr/resumefollow"
	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fullTuningParams() TuningParams {
	return TuningParams{
		MaxOutstandingReadRequests:    types.Int64Value(10),
		MaxOutstandingWriteRequests:   types.Int64Value(8),
		MaxReadRequestOperationCount:  types.Int64Value(512),
		MaxReadRequestSize:            types.StringValue("100mb"),
		MaxRetryDelay:                 customtypes.NewDurationValue("30s"),
		MaxWriteBufferCount:           types.Int64Value(100),
		MaxWriteBufferSize:            types.StringValue("200mb"),
		MaxWriteRequestOperationCount: types.Int64Value(256),
		MaxWriteRequestSize:           types.StringValue("64mb"),
		ReadPollTimeout:               customtypes.NewDurationValue("5m"),
	}
}

func TestApplyToPutAutoFollowRequest_allFields(t *testing.T) {
	t.Parallel()

	p := fullTuningParams()
	req := &putautofollowpattern.Request{}
	diags := ApplyToPutAutoFollowRequest(p, req)
	require.False(t, diags.HasError(), diags)

	require.NotNil(t, req.MaxOutstandingReadRequests)
	assert.Equal(t, 10, *req.MaxOutstandingReadRequests)
	require.NotNil(t, req.MaxOutstandingWriteRequests)
	assert.Equal(t, 8, *req.MaxOutstandingWriteRequests)
	require.NotNil(t, req.MaxReadRequestOperationCount)
	assert.Equal(t, 512, *req.MaxReadRequestOperationCount)
	assert.Equal(t, estypes.ByteSize("100mb"), req.MaxReadRequestSize)
	assert.Equal(t, estypes.Duration("30s"), req.MaxRetryDelay)
	require.NotNil(t, req.MaxWriteBufferCount)
	assert.Equal(t, 100, *req.MaxWriteBufferCount)
	assert.Equal(t, estypes.ByteSize("200mb"), req.MaxWriteBufferSize)
	require.NotNil(t, req.MaxWriteRequestOperationCount)
	assert.Equal(t, 256, *req.MaxWriteRequestOperationCount)
	assert.Equal(t, estypes.ByteSize("64mb"), req.MaxWriteRequestSize)
	assert.Equal(t, estypes.Duration("5m"), req.ReadPollTimeout)
}

func TestApplyToPutAutoFollowRequest_skipsNonPositiveReadRequests(t *testing.T) {
	t.Parallel()

	p := TuningParams{MaxOutstandingReadRequests: types.Int64Value(0)}
	req := &putautofollowpattern.Request{}
	diags := ApplyToPutAutoFollowRequest(p, req)
	require.False(t, diags.HasError())
	assert.Nil(t, req.MaxOutstandingReadRequests)
}

func TestApplyToPutAutoFollowRequest_nullFieldsSkipped(t *testing.T) {
	t.Parallel()

	p := TuningParams{
		MaxOutstandingReadRequests: types.Int64Null(),
		MaxReadRequestSize:         types.StringNull(),
		MaxRetryDelay:              customtypes.NewDurationNull(),
	}
	req := &putautofollowpattern.Request{}
	diags := ApplyToPutAutoFollowRequest(p, req)
	require.False(t, diags.HasError())
	assert.Nil(t, req.MaxOutstandingReadRequests)
	assert.Nil(t, req.MaxReadRequestSize)
	assert.Nil(t, req.MaxRetryDelay)
}

func TestApplyToFollowRequest_allFields(t *testing.T) {
	t.Parallel()

	p := fullTuningParams()
	req := &follow.Request{}
	diags := ApplyToFollowRequest(p, req)
	require.False(t, diags.HasError(), diags)

	require.NotNil(t, req.MaxOutstandingReadRequests)
	assert.Equal(t, int64(10), *req.MaxOutstandingReadRequests)
	require.NotNil(t, req.MaxOutstandingWriteRequests)
	assert.Equal(t, 8, *req.MaxOutstandingWriteRequests)
	require.NotNil(t, req.MaxReadRequestOperationCount)
	assert.Equal(t, 512, *req.MaxReadRequestOperationCount)
	assert.Equal(t, estypes.ByteSize("100mb"), req.MaxReadRequestSize)
	assert.Equal(t, estypes.Duration("30s"), req.MaxRetryDelay)
	require.NotNil(t, req.MaxWriteBufferCount)
	assert.Equal(t, 100, *req.MaxWriteBufferCount)
	assert.Equal(t, estypes.ByteSize("200mb"), req.MaxWriteBufferSize)
	require.NotNil(t, req.MaxWriteRequestOperationCount)
	assert.Equal(t, 256, *req.MaxWriteRequestOperationCount)
	assert.Equal(t, estypes.ByteSize("64mb"), req.MaxWriteRequestSize)
	assert.Equal(t, estypes.Duration("5m"), req.ReadPollTimeout)
}

func TestApplyToResumeFollowRequest_allFields(t *testing.T) {
	t.Parallel()

	p := fullTuningParams()
	req := &resumefollow.Request{}
	ApplyToResumeFollowRequest(p, req)

	require.NotNil(t, req.MaxOutstandingReadRequests)
	assert.Equal(t, int64(10), *req.MaxOutstandingReadRequests)
	require.NotNil(t, req.MaxOutstandingWriteRequests)
	assert.Equal(t, int64(8), *req.MaxOutstandingWriteRequests)
	require.NotNil(t, req.MaxReadRequestOperationCount)
	assert.Equal(t, int64(512), *req.MaxReadRequestOperationCount)
	require.NotNil(t, req.MaxReadRequestSize)
	assert.Equal(t, "100mb", *req.MaxReadRequestSize)
	assert.Equal(t, estypes.Duration("30s"), req.MaxRetryDelay)
	require.NotNil(t, req.MaxWriteBufferCount)
	assert.Equal(t, int64(100), *req.MaxWriteBufferCount)
	require.NotNil(t, req.MaxWriteBufferSize)
	assert.Equal(t, "200mb", *req.MaxWriteBufferSize)
	require.NotNil(t, req.MaxWriteRequestOperationCount)
	assert.Equal(t, int64(256), *req.MaxWriteRequestOperationCount)
	require.NotNil(t, req.MaxWriteRequestSize)
	assert.Equal(t, "64mb", *req.MaxWriteRequestSize)
	assert.Equal(t, estypes.Duration("5m"), req.ReadPollTimeout)
}

func TestTuningParamsFromParameters_allFields(t *testing.T) {
	t.Parallel()

	n10 := int64(10)
	n8 := 8
	n512 := 512
	n100 := 100
	n256 := 256
	params := &estypes.FollowerIndexParameters{
		MaxOutstandingReadRequests:    &n10,
		MaxOutstandingWriteRequests:   &n8,
		MaxReadRequestOperationCount:  &n512,
		MaxReadRequestSize:            estypes.ByteSize("100mb"),
		MaxRetryDelay:                 estypes.Duration("30s"),
		MaxWriteBufferCount:           &n100,
		MaxWriteBufferSize:            estypes.ByteSize("200mb"),
		MaxWriteRequestOperationCount: &n256,
		MaxWriteRequestSize:           estypes.ByteSize("64mb"),
		ReadPollTimeout:               estypes.Duration("5m"),
	}

	p := TuningParamsFromParameters(params)
	assert.Equal(t, types.Int64Value(10), p.MaxOutstandingReadRequests)
	assert.Equal(t, types.Int64Value(8), p.MaxOutstandingWriteRequests)
	assert.Equal(t, types.Int64Value(512), p.MaxReadRequestOperationCount)
	assert.Equal(t, types.StringValue("100mb"), p.MaxReadRequestSize)
	assert.Equal(t, customtypes.NewDurationValue("30s"), p.MaxRetryDelay)
	assert.Equal(t, types.Int64Value(100), p.MaxWriteBufferCount)
	assert.Equal(t, types.StringValue("200mb"), p.MaxWriteBufferSize)
	assert.Equal(t, types.Int64Value(256), p.MaxWriteRequestOperationCount)
	assert.Equal(t, types.StringValue("64mb"), p.MaxWriteRequestSize)
	assert.Equal(t, customtypes.NewDurationValue("5m"), p.ReadPollTimeout)
}

func TestTuningParamsFromParameters_nilPointers(t *testing.T) {
	t.Parallel()

	params := &estypes.FollowerIndexParameters{}
	p := TuningParamsFromParameters(params)

	assert.Equal(t, types.Int64Null(), p.MaxOutstandingReadRequests)
	assert.Equal(t, types.Int64Null(), p.MaxOutstandingWriteRequests)
	assert.Equal(t, types.Int64Null(), p.MaxReadRequestOperationCount)
	assert.Equal(t, types.StringNull(), p.MaxReadRequestSize)
	assert.Equal(t, customtypes.NewDurationNull(), p.MaxRetryDelay)
	assert.Equal(t, types.Int64Null(), p.MaxWriteBufferCount)
	assert.Equal(t, types.StringNull(), p.MaxWriteBufferSize)
	assert.Equal(t, types.Int64Null(), p.MaxWriteRequestOperationCount)
	assert.Equal(t, types.StringNull(), p.MaxWriteRequestSize)
	assert.Equal(t, customtypes.NewDurationNull(), p.ReadPollTimeout)
}
