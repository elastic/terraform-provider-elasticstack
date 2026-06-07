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

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
			},
			"leader_index_patterns": schema.ListAttribute{
				MarkdownDescription: descLeaderIndexPatterns,
				Required:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"leader_index_exclusion_patterns": schema.ListAttribute{
				MarkdownDescription: descLeaderIndexExclusionPatterns,
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"follow_index_pattern": schema.StringAttribute{
				MarkdownDescription: descFollowIndexPattern,
				Optional:            true,
			},
			"settings_raw": schema.StringAttribute{
				MarkdownDescription: descSettingsRaw,
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"max_outstanding_read_requests": schema.Int64Attribute{
				MarkdownDescription: descMaxOutstandingReadRequests,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"max_outstanding_write_requests": schema.Int64Attribute{
				MarkdownDescription: descMaxOutstandingWriteRequests,
				Optional:            true,
			},
			"max_read_request_operation_count": schema.Int64Attribute{
				MarkdownDescription: descMaxReadRequestOperationCount,
				Optional:            true,
			},
			"max_read_request_size": schema.StringAttribute{
				MarkdownDescription: descMaxReadRequestSize,
				Optional:            true,
			},
			"max_retry_delay": schema.StringAttribute{
				MarkdownDescription: descMaxRetryDelay,
				Optional:            true,
				CustomType:          customtypes.DurationType{},
			},
			"max_write_buffer_count": schema.Int64Attribute{
				MarkdownDescription: descMaxWriteBufferCount,
				Optional:            true,
			},
			"max_write_buffer_size": schema.StringAttribute{
				MarkdownDescription: descMaxWriteBufferSize,
				Optional:            true,
			},
			"max_write_request_operation_count": schema.Int64Attribute{
				MarkdownDescription: descMaxWriteRequestOperationCount,
				Optional:            true,
			},
			"max_write_request_size": schema.StringAttribute{
				MarkdownDescription: descMaxWriteRequestSize,
				Optional:            true,
			},
			"read_poll_timeout": schema.StringAttribute{
				MarkdownDescription: descReadPollTimeout,
				Optional:            true,
				CustomType:          customtypes.DurationType{},
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: descActive,
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
		},
	}
}
