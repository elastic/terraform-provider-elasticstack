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
	"fmt"

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
		dataSourceSchemaFactory,
		readQueryRulesetDataSource,
	)
}

func readQueryRulesetDataSource(ctx context.Context, client *clients.ElasticsearchScopedClient, data QueryRulesetData) (QueryRulesetData, diag.Diagnostics) {
	var diags diag.Diagnostics

	rulesetID := data.RulesetID.ValueString()

	id, idDiags := client.ID(ctx, rulesetID)
	diags.Append(idDiags...)
	if diags.HasError() {
		return data, diags
	}
	data.ID = types.StringValue(id.String())

	resp, getDiags := elasticsearch.GetQueryRuleset(ctx, client, rulesetID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return data, diags
	}

	if resp == nil {
		diags.AddError("Query ruleset not found", fmt.Sprintf("Query ruleset '%s' not found", rulesetID))
		return data, diags
	}

	data.populateFromAPI(ctx, resp, &diags)
	return data, diags
}
