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

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ccr"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func getSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: resourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: descID,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: descName,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"remote_cluster": schema.StringAttribute{
				MarkdownDescription: descRemoteCluster,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"leader_index": schema.StringAttribute{
				MarkdownDescription: descLeaderIndex,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data_stream_name": schema.StringAttribute{
				MarkdownDescription: descDataStreamName,
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"settings_raw": schema.StringAttribute{
				MarkdownDescription: descSettingsRaw,
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"max_outstanding_read_requests": schema.Int64Attribute{
				MarkdownDescription: ccr.DescMaxOutstandingReadRequests,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"max_outstanding_write_requests": schema.Int64Attribute{
				MarkdownDescription: ccr.DescMaxOutstandingWriteRequests,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"max_read_request_operation_count": schema.Int64Attribute{
				MarkdownDescription: ccr.DescMaxReadRequestOperationCount,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"max_read_request_size": schema.StringAttribute{
				MarkdownDescription: ccr.DescMaxReadRequestSize,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"max_retry_delay": schema.StringAttribute{
				CustomType:          customtypes.DurationType{},
				MarkdownDescription: ccr.DescMaxRetryDelay,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"max_write_buffer_count": schema.Int64Attribute{
				MarkdownDescription: ccr.DescMaxWriteBufferCount,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"max_write_buffer_size": schema.StringAttribute{
				MarkdownDescription: ccr.DescMaxWriteBufferSize,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"max_write_request_operation_count": schema.Int64Attribute{
				MarkdownDescription: ccr.DescMaxWriteRequestOperationCount,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"max_write_request_size": schema.StringAttribute{
				MarkdownDescription: ccr.DescMaxWriteRequestSize,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"read_poll_timeout": schema.StringAttribute{
				CustomType:          customtypes.DurationType{},
				MarkdownDescription: ccr.DescReadPollTimeout,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"delete_index_on_destroy": schema.BoolAttribute{
				MarkdownDescription: descDeleteIndexOnDestroy,
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"status": schema.StringAttribute{
				MarkdownDescription: descStatus,
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(statusActive),
				Validators: []validator.String{
					stringvalidator.OneOf(statusActive, statusPaused),
				},
			},
		},
	}
}
