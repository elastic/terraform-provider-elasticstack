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
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func (r Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var planModel tfModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if planModel.Type.ValueString() == "cross_cluster" {
		createDiags := r.createCrossClusterAPIKey(ctx, &planModel)
		resp.Diagnostics.Append(createDiags...)
	} else {
		createDiags := r.createAPIKey(ctx, &planModel)
		resp.Diagnostics.Append(createDiags...)
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

	esClient, clientDiags := r.Client().GetElasticsearchClient(ctx, planModel.GetElasticsearchConnection())
	resp.Diagnostics.Append(clientDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	finalModel, found, readDiags := readAPIKey(ctx, esClient, compID.ResourceID, planModel)
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

func (r Resource) validateRestrictionSupport(ctx context.Context, model tfModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(model.RoleDescriptors) {
		return diags
	}

	var roleDescriptors map[string]models.APIKeyRoleDescriptor
	unmarshalDiags := model.RoleDescriptors.Unmarshal(&roleDescriptors)
	if unmarshalDiags.HasError() {
		diags.Append(unmarshalDiags...)
		return diags
	}

	hasRestriction := false
	keysWithRestrictions := []string{}
	for key, descriptor := range roleDescriptors {
		if descriptor.Restriction != nil {
			hasRestriction = true
			keysWithRestrictions = append(keysWithRestrictions, key)
		}
	}

	if hasRestriction {
		isSupported, supportDiags := r.doesCurrentVersionSupportRestrictionOnAPIKey(ctx, model)
		diags.Append(supportDiags...)
		if diags.HasError() {
			return diags
		}

		if !isSupported {
			diags.AddAttributeError(
				path.Root("roles_descriptors"),
				"Specifying `restriction` on an API key role description is not supported in this version of Elasticsearch",
				fmt.Sprintf("Specifying `restriction` on an API key role description is not supported in this version of Elasticsearch. Role descriptor(s) %s", strings.Join(keysWithRestrictions, ", ")),
			)
			return diags
		}
	}

	return diags
}

func (r Resource) doesCurrentVersionSupportRestrictionOnAPIKey(ctx context.Context, model tfModel) (bool, diag.Diagnostics) {
	client, diags := r.Client().GetElasticsearchClient(ctx, model.ElasticsearchConnection)
	if diags.HasError() {
		return false, diags
	}

	currentVersion, sdkDiags := client.ServerVersion(ctx)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return false, diags
	}

	return currentVersion.GreaterThanOrEqual(MinVersionWithRestriction), diags
}

func (r Resource) doesCurrentVersionSupportCrossClusterAPIKey(ctx context.Context, model tfModel) (bool, diag.Diagnostics) {
	client, diags := r.Client().GetElasticsearchClient(ctx, model.ElasticsearchConnection)
	if diags.HasError() {
		return false, diags
	}

	currentVersion, sdkDiags := client.ServerVersion(ctx)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return false, diags
	}

	return currentVersion.GreaterThanOrEqual(MinVersionWithCrossCluster), diags
}

func (r *Resource) createCrossClusterAPIKey(ctx context.Context, planModel *tfModel) diag.Diagnostics {
	client, diags := r.Client().GetElasticsearchClient(ctx, planModel.ElasticsearchConnection)
	if diags.HasError() {
		return diags
	}

	// Check if the current version supports cross-cluster API keys
	isSupported, supportDiags := r.doesCurrentVersionSupportCrossClusterAPIKey(ctx, *planModel)
	diags.Append(supportDiags...)
	if diags.HasError() {
		return diags
	}
	if !isSupported {
		diags.Append(diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Cross-cluster API keys not supported",
				fmt.Sprintf("Cross-cluster API keys are only supported in Elasticsearch version %s and above.", MinVersionWithCrossCluster.String()),
			),
		}...)
		return diags
	}

	// Handle cross-cluster API key creation
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

func (r *Resource) createAPIKey(ctx context.Context, planModel *tfModel) diag.Diagnostics {
	client, diags := r.Client().GetElasticsearchClient(ctx, planModel.ElasticsearchConnection)
	if diags.HasError() {
		return diags
	}

	// Validate restriction support
	diags.Append(r.validateRestrictionSupport(ctx, *planModel)...)
	if diags.HasError() {
		return diags
	}

	// Handle regular API key creation
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
