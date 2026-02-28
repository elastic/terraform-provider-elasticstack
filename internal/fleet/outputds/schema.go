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

package outputds

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d *outputDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns information about a Fleet output. See the [Fleet output API documentation](https://www.elastic.co/docs/api/doc/kibana/v9/group/endpoint-fleet-outputs) for more details.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the output.",
				Required:    true,
			},
			"space_id": schema.StringAttribute{
				Description: "The Kibana space ID where this output is available.",
				Optional:    true,
			},
			"id": schema.StringAttribute{
				Description: "Unique identifier of the output.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The output type.",
				Computed:    true,
			},
			"hosts": schema.ListAttribute{
				Description: "A list of hosts.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"ca_sha256": schema.StringAttribute{
				Description: "Fingerprint of the Elasticsearch CA certificate.",
				Computed:    true,
			},
			"ca_trusted_fingerprint": schema.StringAttribute{
				Description: "Fingerprint of trusted CA.",
				Computed:    true,
			},
			"default_integrations": schema.BoolAttribute{
				Description: "This output is the default for agent integrations.",
				Computed:    true,
			},
			"default_monitoring": schema.BoolAttribute{
				Description: "This output is the default for agent monitoring.",
				Computed:    true,
			},
			"config_yaml": schema.StringAttribute{
				Description: "Advanced YAML configuration.",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}
