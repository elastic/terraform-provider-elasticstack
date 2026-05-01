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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getDataSourceSchema() dsschema.Schema {
	return dsschema.Schema{
		Description: "Returns information about a Fleet output. See the [Fleet output API documentation](https://www.elastic.co/docs/api/doc/kibana/v9/group/endpoint-fleet-outputs) for more details.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Description: "Generated ID for the outputs.",
				Computed:    true,
			},
			"space_id": dsschema.StringAttribute{
				Description: "The Kibana space ID where this output is available.",
				Optional:    true,
			},
			"outputs": dsschema.ListNestedAttribute{
				Description: "The list of outputs",
				Computed:    true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"id": dsschema.StringAttribute{
							Description: "Unique identifier of the output.",
							Computed:    true,
						},
						"name": dsschema.StringAttribute{
							Description: "The name of the output.",
							Computed:    true,
						},
						"type": dsschema.StringAttribute{
							Description: "The output type.",
							Computed:    true,
						},
						"hosts": dsschema.ListAttribute{
							Description: "A list of hosts.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"ca_sha256": dsschema.StringAttribute{
							Description: "Fingerprint of the Elasticsearch CA certificate.",
							Computed:    true,
						},
						"ca_trusted_fingerprint": dsschema.StringAttribute{
							Description: "Fingerprint of trusted CA.",
							Computed:    true,
						},
						"default_integrations": dsschema.BoolAttribute{
							Description: "This output is the default for agent integrations.",
							Computed:    true,
						},
						"default_monitoring": dsschema.BoolAttribute{
							Description: "This output is the default for agent monitoring.",
							Computed:    true,
						},
						"config_yaml": dsschema.StringAttribute{
							Description: "Advanced YAML configuration.",
							Computed:    true,
							Sensitive:   true,
						},
					},
				},
			},
		},
	}
}

func getOutputItemElemType() attr.Type {
	return getDataSourceSchema().Attributes["outputs"].GetType().(attr.TypeWithElementType).ElementType()
}
