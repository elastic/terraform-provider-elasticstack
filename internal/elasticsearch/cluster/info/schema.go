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

package info

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func getDataSourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Gets information about the Elastic cluster.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource.",
				Computed:            true,
			},
			"cluster_name": schema.StringAttribute{
				MarkdownDescription: "Name of the cluster, based on the `cluster.name` setting.",
				Computed:            true,
			},
			"cluster_uuid": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for the cluster.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the node.",
				Computed:            true,
			},
			"tagline": schema.StringAttribute{
				MarkdownDescription: "Elasticsearch tag line.",
				Computed:            true,
			},
			"version": schema.ListNestedAttribute{
				MarkdownDescription: "Contains version information for the Elasticsearch cluster.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"build_date": schema.StringAttribute{
							MarkdownDescription: "Build date.",
							Computed:            true,
						},
						"build_flavor": schema.StringAttribute{
							MarkdownDescription: "Build Flavor.",
							Computed:            true,
						},
						"build_hash": schema.StringAttribute{
							MarkdownDescription: "Short hash of the last git commit in this release.",
							Computed:            true,
						},
						"build_snapshot": schema.BoolAttribute{
							MarkdownDescription: "Build Snapshot.",
							Computed:            true,
						},
						"build_type": schema.StringAttribute{
							MarkdownDescription: "Build Type.",
							Computed:            true,
						},
						"lucene_version": schema.StringAttribute{
							MarkdownDescription: "Lucene Version.",
							Computed:            true,
						},
						"minimum_index_compatibility_version": schema.StringAttribute{
							MarkdownDescription: "Minimum index compatibility version.",
							Computed:            true,
						},
						"minimum_wire_compatibility_version": schema.StringAttribute{
							MarkdownDescription: "Minimum wire compatibility version.",
							Computed:            true,
						},
						"number": schema.StringAttribute{
							MarkdownDescription: "Elasticsearch version number.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}
