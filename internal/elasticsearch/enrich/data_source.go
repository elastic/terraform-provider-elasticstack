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

package enrich

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewEnrichPolicyDataSource() datasource.DataSource {
	return entitycore.NewElasticsearchDataSource[PolicyData](
		entitycore.ComponentElasticsearch,
		"enrich_policy",
		GetDataSourceSchema,
		readDataSource,
	)
}

func GetDataSourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: enrichPolicyDataSourceMarkdownDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the policy.",
				Required:            true,
			},
			"policy_type": schema.StringAttribute{
				MarkdownDescription: "The type of enrich policy, can be one of geo_match, match, range.",
				Computed:            true,
			},
			"indices": schema.SetAttribute{
				MarkdownDescription: "Array of one or more source indices used to create the enrich index.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"match_field": schema.StringAttribute{
				MarkdownDescription: "Field from the source indices used to match incoming documents.",
				Computed:            true,
			},
			"enrich_fields": schema.SetAttribute{
				MarkdownDescription: "Fields to add to matching incoming documents. These fields must be present in the source indices.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"query": schema.StringAttribute{
				MarkdownDescription: "Query used to filter documents in the enrich index. The policy only uses documents matching this query to enrich incoming documents. Defaults to a match_all query.",
				CustomType:          jsontypes.NormalizedType{},
				Computed:            true,
			},
		},
	}
}

func readDataSource(ctx context.Context, esClient *clients.ElasticsearchScopedClient, config PolicyData) (PolicyData, diag.Diagnostics) {
	var diags diag.Diagnostics
	policyName := config.Name.ValueString()

	id, sdkDiags := esClient.ID(ctx, policyName)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return config, diags
	}
	config.ID = types.StringValue(id.String())

	policy, sdkDiags := elasticsearch.GetEnrichPolicy(ctx, esClient, policyName)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return config, diags
	}

	if policy == nil {
		diags.AddError("Policy not found", fmt.Sprintf("Enrich policy '%s' not found", policyName))
		return config, diags
	}

	config.populateFromPolicy(ctx, policy, &diags)
	return config, diags
}
