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

package queryrulesets

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NewQueryRulesetDataSource returns a new query ruleset data source for registration with the provider.
func NewQueryRulesetDataSource() datasource.DataSource {
	return entitycore.NewElasticsearchDataSource(
		entitycore.ComponentElasticsearch,
		"query_ruleset",
		entitycore.ElasticsearchDataSourceOptions[QueryRulesetData]{
			Schema: dataSourceSchemaFactory,
			Read:   readQueryRulesetDataSource,
		},
	)
}

func readQueryRulesetDataSource(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, data QueryRulesetData) (QueryRulesetData, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	id, idDiags := client.ID(ctx, resourceID)
	diags.Append(idDiags...)
	if diags.HasError() {
		return data, false, diags
	}
	data.ID = types.StringValue(id.String())

	resp, getDiags := elasticsearch.GetQueryRuleset(ctx, client, resourceID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return data, false, diags
	}

	if resp == nil {
		return data, false, diags
	}

	data.populateFromAPI(ctx, resp.Rules, &diags)
	return data, !diags.HasError(), diags
}
