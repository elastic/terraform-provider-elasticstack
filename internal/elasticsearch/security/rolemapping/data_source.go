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
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// roleMappingDataSourceModel mirrors Data without entitycore.ResourceTimeoutsField:
// data sources do not expose a timeouts attribute, so reusing the resource model
// (which embeds it) would fail decoding against the timeouts-free data source schema.
type roleMappingDataSourceModel struct {
	entitycore.ElasticsearchConnectionField
	ID            types.String         `tfsdk:"id"`
	Name          types.String         `tfsdk:"name"`
	Enabled       types.Bool           `tfsdk:"enabled"`
	Rules         NormalizedRulesValue `tfsdk:"rules"`
	Roles         types.Set            `tfsdk:"roles"`
	RoleTemplates jsontypes.Normalized `tfsdk:"role_templates"`
	Metadata      jsontypes.Normalized `tfsdk:"metadata"`
}

func (m roleMappingDataSourceModel) toData() Data {
	return Data{
		ElasticsearchConnectionField: m.ElasticsearchConnectionField,
		ID:                           m.ID,
		Name:                         m.Name,
		Enabled:                      m.Enabled,
		Rules:                        m.Rules,
		Roles:                        m.Roles,
		RoleTemplates:                m.RoleTemplates,
		Metadata:                     m.Metadata,
	}
}

func roleMappingDataSourceModelFromData(d Data) roleMappingDataSourceModel {
	return roleMappingDataSourceModel{
		ElasticsearchConnectionField: d.ElasticsearchConnectionField,
		ID:                           d.ID,
		Name:                         d.Name,
		Enabled:                      d.Enabled,
		Rules:                        d.Rules,
		Roles:                        d.Roles,
		RoleTemplates:                d.RoleTemplates,
		Metadata:                     d.Metadata,
	}
}

func NewRoleMappingDataSource() datasource.DataSource {
	return entitycore.NewElasticsearchDataSource[roleMappingDataSourceModel](
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
				CustomType:          NormalizedRulesType{},
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

func readDataSource(ctx context.Context, esClient *clients.ElasticsearchScopedClient, config roleMappingDataSourceModel) (roleMappingDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	roleMappingName := config.Name.ValueString()

	stateData := config.toData()

	id, idDiags := esClient.ID(ctx, roleMappingName)
	diags.Append(idDiags...)
	if diags.HasError() {
		return config, diags
	}
	stateData.ID = types.StringValue(id.String())

	readData, readDiags := readRoleMapping(ctx, stateData, roleMappingName, esClient)
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

	return roleMappingDataSourceModelFromData(*readData), diags
}
