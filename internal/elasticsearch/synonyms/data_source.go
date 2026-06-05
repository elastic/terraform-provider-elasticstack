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

package synonyms

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// synonymSetDataSourceModel mirrors SynonymSetData without entitycore.ResourceTimeoutsField:
// data sources do not expose a timeouts attribute, so reusing the resource model
// (which embeds it) would fail decoding against the timeouts-free data source schema.
type synonymSetDataSourceModel struct {
	entitycore.ElasticsearchConnectionField
	ID           types.String `tfsdk:"id"`
	SynonymSetID types.String `tfsdk:"synonym_set_id"`
	SynonymsSet  types.List   `tfsdk:"synonyms_set"`
}

func (m synonymSetDataSourceModel) toData() SynonymSetData {
	return SynonymSetData{
		ElasticsearchConnectionField: m.ElasticsearchConnectionField,
		ID:                           m.ID,
		SynonymSetID:                 m.SynonymSetID,
		SynonymsSet:                  m.SynonymsSet,
	}
}

func synonymSetDataSourceModelFromData(d SynonymSetData) synonymSetDataSourceModel {
	return synonymSetDataSourceModel{
		ElasticsearchConnectionField: d.ElasticsearchConnectionField,
		ID:                           d.ID,
		SynonymSetID:                 d.SynonymSetID,
		SynonymsSet:                  d.SynonymsSet,
	}
}

// NewSynonymSetDataSource returns a new synonym set data source for registration with the provider.
func NewSynonymSetDataSource() datasource.DataSource {
	return entitycore.NewElasticsearchDataSource[synonymSetDataSourceModel](
		entitycore.ComponentElasticsearch,
		"synonym_set",
		dataSourceSchemaFactory,
		readSynonymSetDataSource,
	)
}

// dataSourceSchemaFactory returns the schema for the synonym set data source.
// The elasticsearch_connection block is injected automatically by the envelope.
func dataSourceSchemaFactory(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: synonymSetDataSourceMarkdownDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource.",
				Computed:            true,
			},
			"synonym_set_id": schema.StringAttribute{
				MarkdownDescription: "The name of the synonym set to look up.",
				Required:            true,
			},
			"synonyms_set": schema.ListNestedAttribute{
				MarkdownDescription: "The list of synonym rules for this synonym set.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The identifier for this synonym rule.",
							Computed:            true,
						},
						synonymsAttrName: schema.StringAttribute{
							MarkdownDescription: "The synonym rule in Solr format (e.g. `\"i-pod, i pod => ipod\"` or `\"universe, cosmos\"`).",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func readSynonymSetDataSource(ctx context.Context, client *clients.ElasticsearchScopedClient, config synonymSetDataSourceModel) (synonymSetDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	data := config.toData()
	synonymSetID := data.SynonymSetID.ValueString()

	id, idDiags := client.ID(ctx, synonymSetID)
	diags.Append(idDiags...)
	if diags.HasError() {
		return config, diags
	}
	data.ID = types.StringValue(id.String())

	rules, getDiags := elasticsearch.GetSynonymSet(ctx, client, synonymSetID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return config, diags
	}

	if rules == nil {
		diags.AddError("Synonym set not found", fmt.Sprintf("Synonym set '%s' not found", synonymSetID))
		return config, diags
	}

	data.populateFromAPI(ctx, rules, &diags)
	return synonymSetDataSourceModelFromData(data), diags
}
