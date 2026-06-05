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
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
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
	ID                            types.String `tfsdk:"id"`
	Name                          types.String `tfsdk:"name"`
	RemoteCluster                 types.String `tfsdk:"remote_cluster"`
	LeaderIndex                   types.String `tfsdk:"leader_index"`
	DataStreamName                types.String `tfsdk:"data_stream_name"`
	SettingsRaw                   types.String `tfsdk:"settings_raw"`
	MaxOutstandingReadRequests    types.Int64  `tfsdk:"max_outstanding_read_requests"`
	MaxOutstandingWriteRequests   types.Int64  `tfsdk:"max_outstanding_write_requests"`
	MaxReadRequestOperationCount  types.Int64  `tfsdk:"max_read_request_operation_count"`
	MaxReadRequestSize            types.String `tfsdk:"max_read_request_size"`
	MaxRetryDelay                 types.String `tfsdk:"max_retry_delay"`
	MaxWriteBufferCount           types.Int64  `tfsdk:"max_write_buffer_count"`
	MaxWriteBufferSize            types.String `tfsdk:"max_write_buffer_size"`
	MaxWriteRequestOperationCount types.Int64  `tfsdk:"max_write_request_operation_count"`
	MaxWriteRequestSize           types.String `tfsdk:"max_write_request_size"`
	ReadPollTimeout               types.String `tfsdk:"read_poll_timeout"`
	DeleteIndexOnDestroy          types.Bool   `tfsdk:"delete_index_on_destroy"`
	Status                        types.String `tfsdk:"status"`
}

func (m Model) GetID() types.String { return m.ID }

func (m Model) GetResourceID() types.String { return m.Name }

var _ entitycore.ElasticsearchResourceModel = Model{}
