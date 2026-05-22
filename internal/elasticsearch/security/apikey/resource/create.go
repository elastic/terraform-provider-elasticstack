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

package resource

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/apikey"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func (r Resource) Create(ctx context.Context, req fwresource.CreateRequest, resp *fwresource.CreateResponse) {
	var planModel apikey.TfModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, clientDiags := r.Client().GetElasticsearchClient(ctx, planModel.ElasticsearchConnection)
	resp.Diagnostics.Append(clientDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if planModel.Type.ValueString() == apikey.CrossClusterAPIKeyType {
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

	compID, idDiags := clients.CompositeIDFromStr(planModel.GetID().ValueString())
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

func (r *Resource) createCrossClusterAPIKey(ctx context.Context, client *clients.ElasticsearchScopedClient, planModel *apikey.TfModel) diag.Diagnostics {
	diags := apikey.CreateCrossClusterAPIKeyOperation(ctx, client, planModel)
	if diags.HasError() {
		return diags
	}
	return assignCompositeID(ctx, client, planModel)
}

func (r *Resource) createAPIKey(ctx context.Context, client *clients.ElasticsearchScopedClient, planModel *apikey.TfModel) diag.Diagnostics {
	diags := apikey.CreateRESTAPIKeyOperation(ctx, client, planModel)
	if diags.HasError() {
		return diags
	}
	return assignCompositeID(ctx, client, planModel)
}

func assignCompositeID(ctx context.Context, client *clients.ElasticsearchScopedClient, planModel *apikey.TfModel) diag.Diagnostics {
	id, diags := client.ID(ctx, planModel.KeyID.ValueString())
	if diags.HasError() {
		return diags
	}
	planModel.ID = basetypes.NewStringValue(id.String())
	return diags
}
