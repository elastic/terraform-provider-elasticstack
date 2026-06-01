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
			"engines_json": schema.StringAttribute{
				Description: "Normalized JSON of the engines array.",
				Computed:    true,
			},
			"status_json": schema.StringAttribute{
				Description: "Normalized JSON of the full status response.",
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
			},
		},
	}
}
