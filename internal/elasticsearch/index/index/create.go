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

package index

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var planModel tfModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := r.Client().GetElasticsearchClient(ctx, planModel.ElasticsearchConnection)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	useExisting := false
	if !planModel.UseExisting.IsNull() && !planModel.UseExisting.IsUnknown() {
		useExisting = planModel.UseExisting.ValueBool()
	}

	if useExisting {
		configuredName := planModel.Name.ValueString()
		if elasticsearch.DateMathIndexNameRe.MatchString(configuredName) {
			resp.Diagnostics.AddWarning(
				"use_existing ignored for date math index names",
				fmt.Sprintf("use_existing has no effect when name is a date math expression (%q); proceeding with normal index creation.", configuredName),
			)
		} else {
			existingPtr, getDiags := elasticsearch.GetIndex(ctx, client, configuredName)
			resp.Diagnostics.Append(getDiags...)
			if resp.Diagnostics.HasError() {
				return
			}
			if existingPtr != nil {
				existingModel, convertDiags := indexStateToModel(*existingPtr)
				resp.Diagnostics.Append(convertDiags...)
				if resp.Diagnostics.HasError() {
					return
				}
				r.adoptExistingIndexOnCreate(ctx, resp, client, &planModel, configuredName, existingModel)
				return
			}
		}
	}

	apiModel, diags := planModel.toAPIModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverFlavor, sdkDiags := client.ServerFlavor(ctx)
	if sdkDiags.HasError() {
		resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		return
	}

	params := planModel.toPutIndexParams(serverFlavor)

	concreteName, diags := elasticsearch.PutIndex(ctx, client, &apiModel, &params)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Compute id from the cluster UUID and the concrete index name returned by
	// Elasticsearch.  For static names concreteName equals the configured name.
	// For date math names concreteName is the resolved index (e.g. logs-2024.01.15).
	id, sdkDiags := client.ID(ctx, concreteName)
	if sdkDiags.HasError() {
		resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		return
	}

	planModel.ID = types.StringValue(id.String())
	planModel.ConcreteName = types.StringValue(concreteName)

	finalModel, diags := readIndex(ctx, planModel, client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, finalModel)...)
}

func (r *Resource) adoptExistingIndexOnCreate(
	ctx context.Context,
	resp *resource.CreateResponse,
	client *clients.ElasticsearchScopedClient,
	plan *tfModel,
	concreteName string,
	existing models.Index,
) {
	mismatches, cmpDiags := compareStaticSettings(ctx, plan, existing)
	resp.Diagnostics.Append(cmpDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(mismatches) > 0 {
		resp.Diagnostics.AddError(
			"existing index has incompatible static settings",
			formatStaticSettingMismatchesDetail(concreteName, mismatches),
		)
		return
	}

	synthetic := tfModel{
		ElasticsearchConnection: plan.ElasticsearchConnection,
	}
	existingState, convertDiags := modelToIndexState(existing)
	resp.Diagnostics.Append(convertDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(synthetic.populateFromAPI(ctx, concreteName, existingState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	planAPIModel, diags := plan.toAPIModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	syntheticAPIModel, diags := synthetic.toAPIModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if typeutils.IsKnown(plan.Alias) && !plan.Alias.Equal(synthetic.Alias) {
		resp.Diagnostics.Append(r.updateAliases(ctx, client, concreteName, planAPIModel.Aliases, syntheticAPIModel.Aliases)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(r.updateSettings(ctx, client, concreteName, planAPIModel.Settings, syntheticAPIModel.Settings)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if typeutils.IsKnown(plan.Mappings) {
		resp.Diagnostics.Append(r.updateMappings(ctx, client, concreteName, plan.Mappings, synthetic.Mappings)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	id, sdkDiags := client.ID(ctx, concreteName)
	if sdkDiags.HasError() {
		resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		return
	}

	plan.ConcreteName = types.StringValue(concreteName)
	plan.ID = types.StringValue(id.String())

	finalModel, diags := readIndex(ctx, *plan, client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if finalModel == nil {
		resp.Diagnostics.AddError(
			"Index disappeared during adoption",
			fmt.Sprintf("index %q was present for updates but not found when reading final state", concreteName),
		)
		return
	}

	resp.Diagnostics.AddWarning(
		"Adopted existing Elasticsearch index",
		fmt.Sprintf("adopted existing index %q rather than creating a new one", concreteName),
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, finalModel)...)
}
