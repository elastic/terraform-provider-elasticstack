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

package template

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// templateDataSourceModel mirrors Model without entitycore.ResourceTimeoutsField:
// data sources do not expose a timeouts attribute, so reusing the resource model
// (which embeds it) would fail decoding against the timeouts-free data source schema.
type templateDataSourceModel struct {
	entitycore.ElasticsearchConnectionField
	ID                              types.String         `tfsdk:"id"`
	Name                            types.String         `tfsdk:"name"`
	ComposedOf                      types.List           `tfsdk:"composed_of"`
	IgnoreMissingComponentTemplates types.List           `tfsdk:"ignore_missing_component_templates"`
	IndexPatterns                   types.Set            `tfsdk:"index_patterns"`
	Metadata                        jsontypes.Normalized `tfsdk:"metadata"`
	Priority                        types.Int64          `tfsdk:"priority"`
	Version                         types.Int64          `tfsdk:"version"`
	AllowAutoCreate                 types.Bool           `tfsdk:"allow_auto_create"`
	DataStream                      types.Object         `tfsdk:"data_stream"`
	Template                        types.Object         `tfsdk:"template"`
}

func (m templateDataSourceModel) GetID() types.String         { return m.ID }
func (m templateDataSourceModel) GetResourceID() types.String { return m.Name }

func (m templateDataSourceModel) toModel() Model {
	return Model{
		ElasticsearchConnectionField:    m.ElasticsearchConnectionField,
		ID:                              m.ID,
		Name:                            m.Name,
		ComposedOf:                      m.ComposedOf,
		IgnoreMissingComponentTemplates: m.IgnoreMissingComponentTemplates,
		IndexPatterns:                   m.IndexPatterns,
		Metadata:                        m.Metadata,
		Priority:                        m.Priority,
		Version:                         m.Version,
		AllowAutoCreate:                 m.AllowAutoCreate,
		DataStream:                      m.DataStream,
		Template:                        m.Template,
	}
}

func templateDataSourceModelFromModel(m Model) templateDataSourceModel {
	return templateDataSourceModel{
		ElasticsearchConnectionField:    m.ElasticsearchConnectionField,
		ID:                              m.ID,
		Name:                            m.Name,
		ComposedOf:                      m.ComposedOf,
		IgnoreMissingComponentTemplates: m.IgnoreMissingComponentTemplates,
		IndexPatterns:                   m.IndexPatterns,
		Metadata:                        m.Metadata,
		Priority:                        m.Priority,
		Version:                         m.Version,
		AllowAutoCreate:                 m.AllowAutoCreate,
		DataStream:                      m.DataStream,
		Template:                        m.Template,
	}
}

func NewDataSource() datasource.DataSource {
	return entitycore.NewElasticsearchDataSource[templateDataSourceModel](
		entitycore.ComponentElasticsearch,
		"index_template",
		entitycore.ElasticsearchDataSourceOptions[templateDataSourceModel]{
			Schema: getDataSourceSchema,
			Read:   readDataSource,
		},
	)
}

func readDataSource(ctx context.Context, esClient *clients.ElasticsearchScopedClient, resourceID string, config templateDataSourceModel) (templateDataSourceModel, bool, diag.Diagnostics) {
	// For the data source there is no prior state: pass config as the prior so that
	// ElasticsearchConnection and any alias reference values come from configuration.
	out, found, diags := readIndexTemplate(ctx, esClient, resourceID, config.toModel())
	if diags.HasError() {
		return config, false, diags
	}
	if !found {
		return config, false, diags
	}

	id, idDiags := esClient.ID(ctx, resourceID)
	diags.Append(idDiags...)
	if diags.HasError() {
		return config, false, diags
	}

	out.ElasticsearchConnection = config.ElasticsearchConnection
	out.Name = types.StringValue(resourceID)
	out.ID = types.StringValue(id.String())

	return templateDataSourceModelFromModel(out), true, diags
}
