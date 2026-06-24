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

package osquerypack

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getDataSourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Reads an Osquery query pack from Kibana by `pack_id` (`saved_object_id`). " +
			"Use this data source to reference prebuilt (read-only) packs that cannot be managed by the " +
			"`elasticstack_kibana_osquery_pack` resource. Requires Kibana 8.5.0 or later.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Composite identifier in the form `<space_id>/<pack_id>`.",
				Computed:            true,
			},
			"pack_id": schema.StringAttribute{
				MarkdownDescription: "Kibana saved object identifier for the pack (`saved_object_id`).",
				Required:            true,
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: "Kibana space identifier. When omitted, the default space is used.",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Human-readable name of the Osquery pack.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the Osquery pack.",
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the pack is enabled.",
				Computed:            true,
			},
			"policy_ids": schema.ListAttribute{
				MarkdownDescription: "Fleet agent policy IDs this pack is deployed to.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"shards": schema.MapAttribute{
				MarkdownDescription: "Percent (1-100) of hosts per policy ID that receive the pack.",
				Computed:            true,
				ElementType:         types.Float64Type,
			},
			"queries": queriesDataSourceSchema(),
			"read_only": schema.BoolAttribute{
				MarkdownDescription: "Whether the pack is prebuilt and read-only. Prebuilt packs can be read by this data source but not managed by the resource.",
				Computed:            true,
			},
		},
	}
}

func queriesDataSourceSchema() schema.MapNestedAttribute {
	return schema.MapNestedAttribute{
		MarkdownDescription: "Osquery queries in the pack. Map keys are query names (canonical identifiers in Kibana).",
		Computed:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: queryDataSourceNestedAttributes(),
		},
	}
}

func queryDataSourceNestedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"query": schema.StringAttribute{
			MarkdownDescription: "Osquery SQL query text.",
			Computed:            true,
		},
		"platform": schema.SetAttribute{
			MarkdownDescription: "Target platforms for the query. Allowed values: `linux`, `darwin`, `windows`.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"version": schema.StringAttribute{
			MarkdownDescription: "Query version string.",
			Computed:            true,
		},
		"snapshot": schema.BoolAttribute{
			MarkdownDescription: "Whether the query is a snapshot.",
			Computed:            true,
		},
		"removed": schema.BoolAttribute{
			MarkdownDescription: "Whether the query is marked removed.",
			Computed:            true,
		},
		"saved_query_id": schema.StringAttribute{
			MarkdownDescription: "References an `elasticstack_kibana_osquery_saved_query` resource.",
			Computed:            true,
		},
		"ecs_mapping": ecsMappingDataSourceSchema(),
	}
}

func ecsMappingDataSourceSchema() schema.MapNestedAttribute {
	return schema.MapNestedAttribute{
		MarkdownDescription: "Maps query result columns to ECS field paths.",
		Computed:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				attrEcsMappingField: schema.StringAttribute{
					MarkdownDescription: "Query result column name to map from.",
					Computed:            true,
				},
				attrEcsMappingValue: schema.StringAttribute{
					MarkdownDescription: "Static scalar ECS mapping value.",
					Computed:            true,
				},
				attrEcsMappingValues: schema.SetAttribute{
					MarkdownDescription: "Static array ECS mapping values.",
					Computed:            true,
					ElementType:         types.StringType,
				},
			},
		},
	}
}
