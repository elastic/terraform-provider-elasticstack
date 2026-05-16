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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewDataSource() datasource.DataSource {
	return entitycore.NewElasticsearchDataSource[Model](
		entitycore.ComponentElasticsearch,
		"index_template",
		getDataSourceSchema,
		readDataSource,
	)
}

func readDataSource(ctx context.Context, esClient *clients.ElasticsearchScopedClient, config Model) (Model, diag.Diagnostics) {
	var diags diag.Diagnostics

	name := config.Name.ValueString()

	// For the data source there is no prior state: pass config as the prior so that
	// ElasticsearchConnection and any alias reference values come from configuration.
	out, found, diags := readIndexTemplate(ctx, esClient, name, config)
	if diags.HasError() {
		return config, diags
	}
	if !found {
		tflog.Info(ctx, fmt.Sprintf(`Index template "%s" not found; leaving data source attributes unset (legacy SDK behavior)`, name))
		return config, diags
	}

	id, sdkDiags := esClient.ID(ctx, name)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return config, diags
	}

	out.ElasticsearchConnection = config.ElasticsearchConnection
	out.Name = types.StringValue(name)
	out.ID = types.StringValue(id.String())

	return out, diags
}
