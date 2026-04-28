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
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// readIndexTemplate loads the named index template from Elasticsearch and maps it into [Model].
// The second return value is true when the template exists; false on 404 or when diagnostics contain errors after a successful GET.
func readIndexTemplate(ctx context.Context, client *clients.ElasticsearchScopedClient, name string) (Model, bool, diag.Diagnostics) {
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
	return out, true, diags
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var prior Model
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}

	compID, diags := clients.CompositeIDFromStrFw(prior.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	templateName := compID.ResourceID

	client, diags := r.Client().GetElasticsearchClient(ctx, prior.ElasticsearchConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, found, diags := readIndexTemplate(ctx, client, templateName)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		tflog.Warn(ctx, fmt.Sprintf(`Index template "%s" not found, removing from state`, templateName))
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(applyTemplateAliasReconciliationFromReference(ctx, &out, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(canonicalizeTemplateAliasSetInModel(ctx, &out)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out.ElasticsearchConnection = prior.ElasticsearchConnection
	out.ID = prior.ID

	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}
