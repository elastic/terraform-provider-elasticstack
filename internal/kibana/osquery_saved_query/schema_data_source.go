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

package osquerysavedquery

import (
	"context"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getDataSourceSchema(_ context.Context) dsschema.Schema {
	return dsschema.Schema{
		MarkdownDescription: "Reads an Osquery saved query from Kibana, including prebuilt queries shipped with the osquery_manager integration. " +
			"Requires Kibana 8.5.0 or later. A common use is looking up saved query IDs referenced by Security detection rule response actions.",
		Attributes: map[string]dsschema.Attribute{
			attrID: dsschema.StringAttribute{
				MarkdownDescription: "Composite identifier in the form `<space_id>/<saved_query_id>`.",
				Computed:            true,
			},
			attrSavedObjectID: dsschema.StringAttribute{
				MarkdownDescription: "Kibana saved object identifier used by Kibana's Osquery saved query detail API.",
				Computed:            true,
			},
			attrSavedQueryID: dsschema.StringAttribute{
				MarkdownDescription: "Stable identifier for the saved query to look up.",
				Required:            true,
			},
			attrSpaceID: dsschema.StringAttribute{
				// Datasource schema (unlike resource schema) has no Default field; resolveDataSourceSpaceID
				// applies clients.DefaultSpaceID at read time when space_id is omitted or empty.
				MarkdownDescription: "Kibana space identifier. When omitted, the default space is used.",
				Optional:            true,
				Computed:            true,
			},
			attrQuery: dsschema.StringAttribute{
				MarkdownDescription: "Osquery SQL query text.",
				Computed:            true,
			},
			attrDescription: dsschema.StringAttribute{
				MarkdownDescription: "Human-readable description of the saved query.",
				Computed:            true,
			},
			attrPlatform: dsschema.SetAttribute{
				MarkdownDescription: "Target platforms for the query.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			attrInterval: dsschema.Int64Attribute{
				MarkdownDescription: "Query execution interval in seconds.",
				Computed:            true,
			},
			attrVersion: dsschema.StringAttribute{
				MarkdownDescription: "Saved query version string.",
				Computed:            true,
			},
			attrSnapshot: dsschema.BoolAttribute{
				MarkdownDescription: "Whether the saved query is a snapshot.",
				Computed:            true,
			},
			attrRemoved: dsschema.BoolAttribute{
				MarkdownDescription: "Whether the saved query is marked removed.",
				Computed:            true,
			},
			attrEcsMapping: ecsMappingDataSourceSchema(),
			attrPrebuilt: dsschema.BoolAttribute{
				MarkdownDescription: "Whether the saved query is prebuilt by the osquery_manager integration package.",
				Computed:            true,
			},
		},
	}
}

func ecsMappingDataSourceSchema() dsschema.MapNestedAttribute {
	return dsschema.MapNestedAttribute{
		MarkdownDescription: "Maps query result columns to ECS field paths.",
		Computed:            true,
		NestedObject: dsschema.NestedAttributeObject{
			Attributes: map[string]dsschema.Attribute{
				attrEcsMappingField: dsschema.StringAttribute{
					MarkdownDescription: "Query result column name to map from.",
					Computed:            true,
				},
				attrEcsMappingValue: dsschema.StringAttribute{
					MarkdownDescription: "Static scalar ECS mapping value.",
					Computed:            true,
				},
				attrEcsMappingValues: dsschema.SetAttribute{
					MarkdownDescription: "Static array ECS mapping values.",
					Computed:            true,
					ElementType:         types.StringType,
				},
			},
		},
	}
}
