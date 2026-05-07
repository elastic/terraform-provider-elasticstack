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

package rolemapping

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewRoleMappingDataSource() datasource.DataSource {
	return entitycore.NewElasticsearchDataSource[Data](
		entitycore.ComponentElasticsearch,
		"security_role_mapping",
		getDataSourceSchema,
		readDataSource,
	)
}

func getDataSourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Retrieves role mappings. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-get-role-mapping.html",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The distinct name that identifies the role mapping, used solely as an identifier.",
				Required:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Mappings that have `enabled` set to `false` are ignored when role mapping is performed.",
				Computed:            true,
			},
			"rules": schema.StringAttribute{
				MarkdownDescription: "The rules that determine which users should be matched by the mapping. A rule is a logical condition that is expressed by using a JSON DSL.",
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"roles": schema.SetAttribute{
				MarkdownDescription: "A list of role names that are granted to the users that match the role mapping rules.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"role_templates": schema.StringAttribute{
				MarkdownDescription: "A list of mustache templates that will be evaluated to determine the roles names that should granted to the users that match the role mapping rules.",
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"metadata": schema.StringAttribute{
				MarkdownDescription: "Additional metadata that helps define which roles are assigned to each user. Keys beginning with `_` are reserved for system usage.",
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
		},
	}
}

func readDataSource(ctx context.Context, esClient *clients.ElasticsearchScopedClient, config Data) (Data, diag.Diagnostics) {
	var diags diag.Diagnostics
	roleMappingName := config.Name.ValueString()

	id, sdkDiags := esClient.ID(ctx, roleMappingName)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return config, diags
	}
	config.ID = types.StringValue(id.String())

	readData, readDiags := readRoleMapping(ctx, config, roleMappingName, esClient)
	diags.Append(readDiags...)
	if diags.HasError() {
		return config, diags
	}

	if readData == nil {
		diags.AddError(
			"Role mapping not found",
			fmt.Sprintf("Role mapping '%s' not found", roleMappingName),
		)
		return config, diags
	}

	return *readData, diags
}
