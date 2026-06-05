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

// queryRulesetDataSourceModel mirrors QueryRulesetData without
// entitycore.ResourceTimeoutsField: data sources do not expose a timeouts
// attribute, so reusing the resource model (which embeds it) would fail
// decoding against the timeouts-free data source schema.
type queryRulesetDataSourceModel struct {
	entitycore.ElasticsearchConnectionField
	ID        types.String `tfsdk:"id"`
	RulesetID types.String `tfsdk:"ruleset_id"`
	Rules     types.List   `tfsdk:"rules"`
}

// GetVersionRequirements satisfies [entitycore.WithVersionRequirements] so the
// data source enforces the same minimum Elasticsearch version as the resource.
func (queryRulesetDataSourceModel) GetVersionRequirements(ctx context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return QueryRulesetData{}.GetVersionRequirements(ctx)
}

func (m queryRulesetDataSourceModel) toData() QueryRulesetData {
	return QueryRulesetData{
		ElasticsearchConnectionField: m.ElasticsearchConnectionField,
		ID:                           m.ID,
		RulesetID:                    m.RulesetID,
		Rules:                        m.Rules,
	}
}

func queryRulesetDataSourceModelFromData(d QueryRulesetData) queryRulesetDataSourceModel {
	return queryRulesetDataSourceModel{
		ElasticsearchConnectionField: d.ElasticsearchConnectionField,
		ID:                           d.ID,
		RulesetID:                    d.RulesetID,
		Rules:                        d.Rules,
	}
}

// NewQueryRulesetDataSource returns a new query ruleset data source for registration with the provider.
func NewQueryRulesetDataSource() datasource.DataSource {
	return entitycore.NewElasticsearchDataSource[queryRulesetDataSourceModel](
		entitycore.ComponentElasticsearch,
		"query_ruleset",
		dataSourceSchemaFactory,
		readQueryRulesetDataSource,
	)
}

func readQueryRulesetDataSource(ctx context.Context, client *clients.ElasticsearchScopedClient, config queryRulesetDataSourceModel) (queryRulesetDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	data := config.toData()
	rulesetID := data.RulesetID.ValueString()

	id, idDiags := client.ID(ctx, rulesetID)
	diags.Append(idDiags...)
	if diags.HasError() {
		return config, diags
	}
	data.ID = types.StringValue(id.String())

	resp, getDiags := elasticsearch.GetQueryRuleset(ctx, client, rulesetID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return config, diags
	}

	if resp == nil {
		diags.AddError("Query ruleset not found", fmt.Sprintf("Query ruleset '%s' not found", rulesetID))
		return config, diags
	}

	data.populateFromAPI(ctx, resp.Rules, &diags)
	return queryRulesetDataSourceModelFromData(data), diags
}
