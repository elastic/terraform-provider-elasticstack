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
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewDataSource() datasource.DataSource {
	return entitycore.NewElasticsearchDataSource[Model](
		entitycore.ComponentElasticsearch,
		"index_template",
		entitycore.ElasticsearchDataSourceOptions[Model]{
			Schema: getDataSourceSchema,
			Read:   readDataSource,
		},
	)
}

func readDataSource(ctx context.Context, esClient *clients.ElasticsearchScopedClient, resourceID string, config Model) (Model, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	out, found, diags := readIndexTemplate(ctx, esClient, resourceID, config)
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

	return out, true, diags
}
