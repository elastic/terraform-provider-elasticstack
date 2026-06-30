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

package tag

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func getDataSourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Lists Kibana tags, optionally filtered by an Elasticsearch `simple_query_string` expression against `name` and `description`. Requires Kibana 9.5.0 or later.",
		Attributes: map[string]schema.Attribute{
			attrQuery: schema.StringAttribute{
				MarkdownDescription: "Elasticsearch `simple_query_string` filter applied to tag `name` and `description` fields. When omitted, all tags in the space are returned.",
				Optional:            true,
			},
			attrSpaceID: schema.StringAttribute{
				MarkdownDescription: "Kibana space identifier. When omitted, the default space is used.",
				Optional:            true,
				Computed:            true,
			},
			attrTags: schema.ListNestedAttribute{
				MarkdownDescription: "Tags matching the query in the configured space.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Tag UUID.",
							Computed:            true,
						},
						attrName: schema.StringAttribute{
							MarkdownDescription: "Display name of the tag.",
							Computed:            true,
						},
						attrColor: schema.StringAttribute{
							MarkdownDescription: "Hex color of the tag.",
							Computed:            true,
						},
						attrDescription: schema.StringAttribute{
							MarkdownDescription: "Description of the tag.",
							Computed:            true,
						},
						attrManaged: schema.BoolAttribute{
							MarkdownDescription: "Whether the tag is managed by Kibana.",
							Computed:            true,
						},
						attrCreatedAt: schema.StringAttribute{
							MarkdownDescription: "ISO 8601 timestamp when the tag was created.",
							Computed:            true,
						},
						attrUpdatedAt: schema.StringAttribute{
							MarkdownDescription: "ISO 8601 timestamp when the tag was last updated.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}
