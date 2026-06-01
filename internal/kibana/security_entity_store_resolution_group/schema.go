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

package security_entity_store_resolution_group

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

const (
	attrID                  = "id"
	attrSpaceID             = "space_id"
	attrEntityID            = "entity_id"
	attrResolutionGroupJSON = "resolution_group_json"
)

func getDataSourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Retrieve the resolution group for a given entity in the Kibana Entity Store. " +
			"Returns the target entity, all linked alias entities, and the group size. " +
			"Requires Elastic Stack 9.1.0 or later.",
		Attributes: map[string]schema.Attribute{
			attrID: schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The composite ID: `<space_id>/<entity_id>`.",
			},
			attrSpaceID: schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "An identifier for the space. If not provided, the default space is used.",
			},
			attrEntityID: schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The entity identifier to look up the resolution group for.",
			},
			attrResolutionGroupJSON: schema.StringAttribute{
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
				MarkdownDescription: "The normalised JSON representation of the resolution group returned by the Kibana API.",
			},
		},
	}
}
