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

package apikey

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planModel tfModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, clientDiags := r.Client().GetElasticsearchClient(ctx, planModel.GetElasticsearchConnection())
	resp.Diagnostics.Append(clientDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if planModel.Type.ValueString() == "cross_cluster" {
		updateDiags := updateCrossClusterAPIKey(ctx, client, planModel)
		resp.Diagnostics.Append(updateDiags...)
	} else {
		updateDiags := updateAPIKey(ctx, client, planModel)
		resp.Diagnostics.Append(updateDiags...)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	compID, idDiags := clients.CompositeIDFromStrFw(planModel.GetID().ValueString())
	resp.Diagnostics.Append(idDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	finalModel, found, readDiags := readAPIKey(ctx, client, compID.ResourceID, planModel)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("API Key Not Found After Update", fmt.Sprintf("API key %q was not found immediately after update.", compID.ResourceID))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, finalModel)...)
}

func updateCrossClusterAPIKey(ctx context.Context, client *clients.ElasticsearchScopedClient, planModel tfModel) diag.Diagnostics {
	var diags diag.Diagnostics

	updateRequest, modelDiags := planModel.toUpdateCrossClusterAPIRequest(ctx)
	diags.Append(modelDiags...)
	if diags.HasError() {
		return diags
	}

	diags.Append(elasticsearch.UpdateCrossClusterAPIKey(ctx, client, planModel.KeyID.ValueString(), updateRequest)...)
	return diags
}

func updateAPIKey(ctx context.Context, client *clients.ElasticsearchScopedClient, planModel tfModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(validateRestrictionSupport(ctx, client, planModel)...)
	if diags.HasError() {
		return diags
	}

	updateRequest, modelDiags := planModel.toUpdateAPIRequest()
	diags.Append(modelDiags...)
	if diags.HasError() {
		return diags
	}

	diags.Append(elasticsearch.UpdateAPIKey(ctx, client, planModel.KeyID.ValueString(), updateRequest)...)
	return diags
}
