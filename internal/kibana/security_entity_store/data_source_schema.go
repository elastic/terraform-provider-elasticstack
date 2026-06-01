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

package security_entity_store

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func getDataSourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "Reads Elastic Security Entity Store status for a Kibana space.",
		Attributes: map[string]schema.Attribute{
			"space_id": schema.StringAttribute{
				Description: "An identifier for the Kibana space. If omitted, the default space is used.",
				Optional:    true,
				Computed:    true,
			},
			"include_components": schema.BoolAttribute{
				Description: "If true, returns a detailed status of each engine including all its components.",
				Optional:    true,
			},
			"installed": schema.BoolAttribute{
				Description: "True when the Entity Store is installed.",
				Computed:    true,
			},
			"overall_status": schema.StringAttribute{
				Description: "The overall operational status of the Entity Store.",
				Computed:    true,
			},
			"engines": schema.ListNestedAttribute{
				Description: "Per-engine status details.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Description: "The entity type managed by this engine.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Current status of the engine.",
							Computed:    true,
						},
						"index_pattern": schema.StringAttribute{
							Description: "Index pattern used by the engine.",
							Computed:    true,
						},
						"field_history_length": schema.Int64Attribute{
							Description: "Number of historical values kept per field.",
							Computed:    true,
						},
						"delay": schema.StringAttribute{
							Description: "Delay used for log extraction.",
							Computed:    true,
						},
						"frequency": schema.StringAttribute{
							Description: "Frequency used for log extraction.",
							Computed:    true,
						},
						"lookback_period": schema.StringAttribute{
							Description: "Lookback period used for log extraction.",
							Computed:    true,
						},
						"filter": schema.StringAttribute{
							Description: "Filter query applied to the engine.",
							Computed:    true,
						},
						"timeout": schema.StringAttribute{
							Description: "Timeout setting for the engine.",
							Computed:    true,
						},
						"timestamp_field": schema.StringAttribute{
							Description: "Timestamp field used by the engine.",
							Computed:    true,
						},
						"error_action": schema.StringAttribute{
							Description: "Action associated with the last engine error, if any.",
							Computed:    true,
						},
						"error_message": schema.StringAttribute{
							Description: "Message describing the last engine error, if any.",
							Computed:    true,
						},
						"components": schema.ListNestedAttribute{
							Description: "Component-level status for the engine.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Description: "Component identifier.",
										Computed:    true,
									},
									"installed": schema.BoolAttribute{
										Description: "Whether the component is installed.",
										Computed:    true,
									},
									"resource": schema.StringAttribute{
										Description: "Type of Elasticsearch or Kibana resource backing this component.",
										Computed:    true,
									},
									"health": schema.StringAttribute{
										Description: "Health status of the component.",
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
			"status_json": schema.StringAttribute{
				Description: "Normalized JSON of the full status response.",
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
			},
		},
	}
}
