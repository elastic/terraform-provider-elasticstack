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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NewSynonymSetDataSource returns a new synonym set data source for registration with the provider.
func NewSynonymSetDataSource() datasource.DataSource {
	return entitycore.NewElasticsearchDataSource(
		entitycore.ComponentElasticsearch,
		"synonym_set",
		entitycore.ElasticsearchDataSourceOptions[SynonymSetData]{
			Schema: dataSourceSchemaFactory,
			Read:   readSynonymSetDataSource,
		},
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

func readSynonymSetDataSource(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, data SynonymSetData) (SynonymSetData, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	id, idDiags := client.ID(ctx, resourceID)
	diags.Append(idDiags...)
	if diags.HasError() {
		return data, false, diags
	}
	data.ID = types.StringValue(id.String())

	rules, getDiags := elasticsearch.GetSynonymSet(ctx, client, resourceID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return data, false, diags
	}

	if rules == nil {
		return data, false, diags
	}

	data.populateFromAPI(ctx, rules, &diags)
	return data, !diags.HasError(), diags
}
