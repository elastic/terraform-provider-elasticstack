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
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// readIndexTemplate loads the named index template from Elasticsearch and maps it into [Model].
// It accepts the prior state model so alias reconciliation and canonicalization can run inside the
// callback. The second return value is true when the template exists; false on 404.
// ID and ElasticsearchConnection are copied from prior to the returned model.
func readIndexTemplate(ctx context.Context, client *clients.ElasticsearchScopedClient, name string, prior Model) (Model, bool, diag.Diagnostics) {
	tpl, diags := elasticsearch.GetIndexTemplate(ctx, client, name)
	if diags.HasError() {
		return Model{}, false, diags
	}
	if tpl == nil {
		return Model{}, false, diags
	}

	var out Model
	diags.Append(out.fromAPIModel(ctx, tpl.Name, &tpl.IndexTemplate)...)
	if diags.HasError() {
		return Model{}, false, diags
	}

	diags.Append(applyTemplateAliasReconciliationFromReference(ctx, &out, &prior)...)
	if diags.HasError() {
		return Model{}, false, diags
	}
	diags.Append(canonicalizeTemplateAliasSetInModel(ctx, &out)...)
	if diags.HasError() {
		return Model{}, false, diags
	}

	out.ID = prior.ID
	out.ElasticsearchConnection = prior.ElasticsearchConnection

	return out, true, diags
}
