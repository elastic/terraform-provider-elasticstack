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
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
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

func validateRestrictionSupport(ctx context.Context, client *clients.ElasticsearchScopedClient, model tfModel) diag.Diagnostics {
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
		currentVersion, verDiags := client.ServerVersion(ctx)
		diags.Append(verDiags...)
		if diags.HasError() {
			return diags
		}

		if currentVersion.LessThan(MinVersionWithRestriction) {
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

func (r Resource) doesCurrentVersionSupportCrossClusterAPIKey(ctx context.Context, model tfModel) (bool, diag.Diagnostics) {
	client, diags := r.Client().GetElasticsearchClient(ctx, model.ElasticsearchConnection)
	if diags.HasError() {
		return false, diags
	}

	currentVersion, verDiags := client.ServerVersion(ctx)
	diags.Append(verDiags...)
	if diags.HasError() {
		return false, diags
	}

	return currentVersion.GreaterThanOrEqual(MinVersionWithCrossCluster), diags
}

func (r *Resource) createCrossClusterAPIKey(ctx context.Context, client *clients.ElasticsearchScopedClient, planModel *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics

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

	id, idDiags := client.ID(ctx, putResponse.Id)
	diags.Append(idDiags...)
	if diags.HasError() {
		return diags
	}

	planModel.ID = basetypes.NewStringValue(id.String())
	planModel.populateFromCrossClusterCreate(putResponse)
	return diags
}

func (r *Resource) createAPIKey(ctx context.Context, client *clients.ElasticsearchScopedClient, planModel *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(validateRestrictionSupport(ctx, client, *planModel)...)
	if diags.HasError() {
		return diags
	}

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

	id, idDiags := client.ID(ctx, putResponse.Id)
	diags.Append(idDiags...)
	if diags.HasError() {
		return diags
	}

	planModel.ID = basetypes.NewStringValue(id.String())
	planModel.populateFromCreate(putResponse)
	return diags
}
