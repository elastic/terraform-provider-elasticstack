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

package workflow

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Schema defines the schema for the data source.
func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Export an Agent Builder workflow by ID. See https://www.elastic.co/guide/en/kibana/current/agent-builder-api.html",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The workflow ID to export.",
				Required:    true,
			},
			"space_id": schema.StringAttribute{
				Description: "An identifier for the space. If space_id is not provided, the default space is used.",
				Optional:    true,
				Computed:    true,
			},
			"workflow_id": schema.StringAttribute{
				Description: "The ID of the exported workflow.",
				Computed:    true,
			},
			"yaml": schema.StringAttribute{
				Description: "The exported workflow definition in YAML format.",
				Computed:    true,
			},
		},
	}
}
