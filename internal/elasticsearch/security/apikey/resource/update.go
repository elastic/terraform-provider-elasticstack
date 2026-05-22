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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/apikey"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func writeAPIKey(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.WriteRequest[apikey.TfModel]) (entitycore.WriteResult[apikey.TfModel], diag.Diagnostics) {
	planModel := req.Plan
	var diags diag.Diagnostics
	if planModel.Type.ValueString() == apikey.CrossClusterAPIKeyType {
		diags.Append(updateCrossClusterAPIKey(ctx, client, planModel)...)
	} else {
		diags.Append(updateAPIKey(ctx, client, planModel)...)
	}
	return entitycore.WriteResult[apikey.TfModel]{Model: planModel}, diags
}

func updateCrossClusterAPIKey(ctx context.Context, client *clients.ElasticsearchScopedClient, planModel apikey.TfModel) diag.Diagnostics {
	var diags diag.Diagnostics

	updateRequest, modelDiags := planModel.ToUpdateCrossClusterAPIRequest(ctx)
	diags.Append(modelDiags...)
	if diags.HasError() {
		return diags
	}

	diags.Append(elasticsearch.UpdateCrossClusterAPIKey(ctx, client, planModel.KeyID.ValueString(), updateRequest)...)
	return diags
}

func updateAPIKey(ctx context.Context, client *clients.ElasticsearchScopedClient, planModel apikey.TfModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(apikey.ValidateRestrictionSupport(ctx, client, planModel)...)
	if diags.HasError() {
		return diags
	}

	updateRequest, modelDiags := planModel.ToUpdateAPIRequest()
	diags.Append(modelDiags...)
	if diags.HasError() {
		return diags
	}

	diags.Append(elasticsearch.UpdateAPIKey(ctx, client, planModel.KeyID.ValueString(), updateRequest)...)
	return diags
}
