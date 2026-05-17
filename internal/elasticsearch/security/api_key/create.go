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
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func (r Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var planModel tfModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, clientDiags := r.Client().GetElasticsearchClient(ctx, planModel.ElasticsearchConnection)
	resp.Diagnostics.Append(clientDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(entitycore.EnforceVersionRequirements(ctx, client, &planModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if planModel.Type.ValueString() == "cross_cluster" {
		resp.Diagnostics.Append(r.createCrossClusterAPIKey(ctx, client, &planModel)...)
	} else {
		resp.Diagnostics.Append(r.createAPIKey(ctx, client, &planModel)...)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, planModel)...)
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
		resp.Diagnostics.AddError("API Key Not Found After Create", fmt.Sprintf("API key %q was not found immediately after creation.", compID.ResourceID))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, finalModel)...)
}

func (r *Resource) createCrossClusterAPIKey(ctx context.Context, client *clients.ElasticsearchScopedClient, planModel *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics

	createRequest, modelDiags := planModel.toCrossClusterAPICreateRequest(ctx)
	diags.Append(modelDiags...)
	if diags.HasError() {
		return diags
	}

	putResponse, createDiags := elasticsearch.CreateCrossClusterAPIKey(ctx, client, createRequest)
	diags.Append(createDiags...)
	if diags.HasError() {
		return diags
	}
	if putResponse == nil {
		diags.Append(diag.Diagnostics{
			diag.NewErrorDiagnostic("API Key Creation Failed", "Cross-cluster API key creation returned nil response"),
		}...)
		return diags
	}

	id, sdkDiags := client.ID(ctx, putResponse.Id)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return diags
	}

	planModel.ID = basetypes.NewStringValue(id.String())
	planModel.populateFromCrossClusterCreate(putResponse)
	return diags
}

func (r *Resource) createAPIKey(ctx context.Context, client *clients.ElasticsearchScopedClient, planModel *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics

	createRequest, modelDiags := planModel.toAPICreateRequest()
	diags.Append(modelDiags...)
	if diags.HasError() {
		return diags
	}

	putResponse, createDiags := elasticsearch.CreateAPIKey(ctx, client, createRequest)
	diags.Append(createDiags...)
	if diags.HasError() {
		return diags
	}
	if putResponse == nil {
		diags.Append(diag.Diagnostics{
			diag.NewErrorDiagnostic("API Key Creation Failed", "API key creation returned nil response"),
		}...)
		return diags
	}

	id, sdkDiags := client.ID(ctx, putResponse.Id)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return diags
	}

	planModel.ID = basetypes.NewStringValue(id.String())
	planModel.populateFromCreate(putResponse)
	return diags
}
