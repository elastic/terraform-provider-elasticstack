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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// CreateRESTAPIKeyOperation runs the create-side flow shared by the resource
// Create path and the ephemeral resource Open path for `rest` API keys: it
// validates restriction support against the server version, builds the typed
// create request from the model, calls Elasticsearch, and populates the model
// fields returned by the create response. Callers are responsible for assigning
// any composite ID and persisting the resulting state.
func CreateRESTAPIKeyOperation(ctx context.Context, client *clients.ElasticsearchScopedClient, model *TfModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(ValidateRestrictionSupport(ctx, client, *model)...)
	if diags.HasError() {
		return diags
	}

	createRequest, modelDiags := model.toAPICreateRequest()
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
		diags.AddError("API Key Creation Failed", "API key creation returned nil response")
		return diags
	}

	model.populateFromCreate(putResponse)
	return diags
}

// CreateCrossClusterAPIKeyOperation runs the create-side flow shared by the
// resource Create path and the ephemeral resource Open path for
// `cross_cluster` API keys: it enforces the cross-cluster version requirement,
// builds the typed create request from the model, calls Elasticsearch, and
// populates the model fields returned by the create response. Callers are
// responsible for assigning any composite ID and persisting the resulting
// state.
func CreateCrossClusterAPIKeyOperation(ctx context.Context, client *clients.ElasticsearchScopedClient, model *TfModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(entitycore.EnforceVersionRequirements(ctx, client, model)...)
	if diags.HasError() {
		return diags
	}

	createRequest, modelDiags := model.toCrossClusterAPICreateRequest(ctx)
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
		diags.AddError("API Key Creation Failed", "Cross-cluster API key creation returned nil response")
		return diags
	}

	model.populateFromCrossClusterCreate(putResponse)
	return diags
}
