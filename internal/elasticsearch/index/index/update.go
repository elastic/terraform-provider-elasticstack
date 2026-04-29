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
	"reflect"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planModel tfModel
	var stateModel tfModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &stateModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := r.Client().GetElasticsearchClient(ctx, planModel.ElasticsearchConnection)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	// Use the concrete index identity from the current state so update operations
	// target the concrete managed index, not the configured (possibly date math) name.
	stateID, diags := stateModel.GetID()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	concreteName := stateID.ResourceID

	// Carry the id and concrete_name forward from state into the plan model so the
	// post-update read uses the same concrete identity.
	planModel.ID = stateModel.ID
	planModel.ConcreteName = stateModel.ConcreteName

	planAPIModel, diags := planModel.toAPIModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateAPIModel, diags := stateModel.toAPIModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !planModel.Alias.Equal(stateModel.Alias) {
		resp.Diagnostics.Append(r.updateAliases(ctx, client, concreteName, planAPIModel.Aliases, stateAPIModel.Aliases)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(r.updateSettings(ctx, client, concreteName, planAPIModel.Settings, stateAPIModel.Settings)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.updateMappings(ctx, client, concreteName, planModel.Mappings, stateModel.Mappings)...)
	if resp.Diagnostics.HasError() {
		return
	}

	finalModel, diags := readIndex(ctx, planModel, client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, finalModel)...)
}

func (r *Resource) updateAliases(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	indexName string,
	planAliases map[string]models.IndexAlias,
	stateAliases map[string]models.IndexAlias,
) diag.Diagnostics {
	var diags diag.Diagnostics
	aliasesToDelete := []string{}
	for aliasName := range stateAliases {
		if _, ok := planAliases[aliasName]; !ok {
			aliasesToDelete = append(aliasesToDelete, aliasName)
		}
	}

	if len(aliasesToDelete) > 0 {
		opDiags := elasticsearch.DeleteIndexAlias(ctx, client, indexName, aliasesToDelete)
		diags.Append(opDiags...)
		if diags.HasError() {
			return diags
		}
	}

	for _, alias := range planAliases {
		opDiags := elasticsearch.UpdateIndexAlias(ctx, client, indexName, &alias)
		diags.Append(opDiags...)
		if diags.HasError() {
			return diags
		}
	}

	return diags
}

func (r *Resource) updateSettings(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	indexName string,
	planSettings map[string]any,
	stateSettings map[string]any,
) diag.Diagnostics {
	var diags diag.Diagnostics

	planDynamicSettings := map[string]any{}
	stateDynamicSettings := map[string]any{}

	for _, key := range dynamicSettingsKeys {
		if planSetting, ok := planSettings[key]; ok {
			planDynamicSettings[key] = planSetting
		}

		if stateSetting, ok := stateSettings[key]; ok {
			stateDynamicSettings[key] = stateSetting
		}
	}

	if !reflect.DeepEqual(planDynamicSettings, stateDynamicSettings) {
		// Settings which are being removed must be explicitly set to null in the new settings
		for setting := range stateDynamicSettings {
			if _, ok := planDynamicSettings[setting]; !ok {
				planDynamicSettings[setting] = nil
			}
		}

		opDiags := elasticsearch.UpdateIndexSettings(ctx, client, indexName, planDynamicSettings)
		diags.Append(opDiags...)
		if diags.HasError() {
			return diags
		}
	}

	return diags
}

func (r *Resource) updateMappings(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	indexName string,
	planMappings mappingsValue,
	stateMappings mappingsValue,
) diag.Diagnostics {
	var diags diag.Diagnostics

	areEqual, diags := planMappings.StringSemanticEquals(ctx, stateMappings)
	if diags.HasError() {
		return diags
	}

	if areEqual {
		return nil
	}

	opDiags := elasticsearch.UpdateIndexMappings(ctx, client, indexName, planMappings.ValueString())
	diags.Append(opDiags...)
	if diags.HasError() {
		return diags
	}

	return diags
}
