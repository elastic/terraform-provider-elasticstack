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

package slm

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const defaultExpandWildcards = "open,hidden"
const defaultSnapshotName = "<snap-{now/d}>"

func GetSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: schemaMarkdownDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "ID for the snapshot lifecycle policy you want to create or update.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"schedule": schema.StringAttribute{
				MarkdownDescription: "Periodic or absolute schedule at which the policy creates snapshots.",
				Required:            true,
			},
			"repository": schema.StringAttribute{
				MarkdownDescription: "Repository used to store snapshots created by this policy.",
				Required:            true,
			},
			"snapshot_name": schema.StringAttribute{
				MarkdownDescription: "Name automatically assigned to each snapshot created by the policy.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(defaultSnapshotName),
			},
			"expand_wildcards": schema.StringAttribute{
				MarkdownDescription: "Determines how wildcard patterns in the `indices` parameter match data streams and indices. Supports comma-separated values, such as `closed,hidden`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(defaultExpandWildcards),
				Validators: []validator.String{
					expandWildcardsValidator{},
				},
			},
			"ignore_unavailable": schema.BoolAttribute{
				MarkdownDescription: "If `false`, the snapshot fails if any data stream or index in indices is missing or closed. If `true`, the snapshot ignores missing or closed data streams and indices.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"include_global_state": schema.BoolAttribute{
				MarkdownDescription: "If `true`, include the cluster state in the snapshot.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"indices": schema.ListAttribute{
				MarkdownDescription: "List of data streams and indices to include in the snapshot.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"feature_states": schema.SetAttribute{
				MarkdownDescription: "Feature states to include in the snapshot.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"metadata": schema.StringAttribute{
				MarkdownDescription: "Attaches arbitrary metadata to the snapshot.",
				Optional:            true,
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"partial": schema.BoolAttribute{
				MarkdownDescription: "If `false`, the entire snapshot will fail if one or more indices included in the snapshot do not have all primary shards available.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"expire_after": schema.StringAttribute{
				MarkdownDescription: "Time period after which a snapshot is considered expired and eligible for deletion.",
				Optional:            true,
				Computed:            true,
			},
			"max_count": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of snapshots to retain, even if the snapshots have not yet expired.",
				Optional:            true,
				Computed:            true,
			},
			"min_count": schema.Int64Attribute{
				MarkdownDescription: "Minimum number of snapshots to retain, even if the snapshots have expired.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}
