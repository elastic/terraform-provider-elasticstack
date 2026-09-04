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
	"github.com/elastic/go-elasticsearch/v8/typedapi/ccr/follow"
	"github.com/elastic/go-elasticsearch/v8/typedapi/ccr/resumefollow"
	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ccr"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	statusActive = "active"
	statusPaused = "paused"
)

// Model is the Terraform state model for elasticstack_elasticsearch_ccr_follower_index.
type Model struct {
	entitycore.ElasticsearchConnectionField
	entitycore.ResourceTimeoutsField
	ID                            types.String         `tfsdk:"id"`
	Name                          types.String         `tfsdk:"name"`
	RemoteCluster                 types.String         `tfsdk:"remote_cluster"`
	LeaderIndex                   types.String         `tfsdk:"leader_index"`
	DataStreamName                types.String         `tfsdk:"data_stream_name"`
	SettingsRaw                   jsontypes.Normalized `tfsdk:"settings_raw"`
	MaxOutstandingReadRequests    types.Int64          `tfsdk:"max_outstanding_read_requests"`
	MaxOutstandingWriteRequests   types.Int64          `tfsdk:"max_outstanding_write_requests"`
	MaxReadRequestOperationCount  types.Int64          `tfsdk:"max_read_request_operation_count"`
	MaxReadRequestSize            types.String         `tfsdk:"max_read_request_size"`
	MaxRetryDelay                 customtypes.Duration `tfsdk:"max_retry_delay"`
	MaxWriteBufferCount           types.Int64          `tfsdk:"max_write_buffer_count"`
	MaxWriteBufferSize            types.String         `tfsdk:"max_write_buffer_size"`
	MaxWriteRequestOperationCount types.Int64          `tfsdk:"max_write_request_operation_count"`
	MaxWriteRequestSize           types.String         `tfsdk:"max_write_request_size"`
	ReadPollTimeout               customtypes.Duration `tfsdk:"read_poll_timeout"`
	DeleteIndexOnDestroy          types.Bool           `tfsdk:"delete_index_on_destroy"`
	Status                        types.String         `tfsdk:"status"`
}

func (m Model) GetID() types.String { return m.ID }

func (m Model) GetResourceID() types.String { return m.Name }

var _ entitycore.ElasticsearchResourceModel = Model{}

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
