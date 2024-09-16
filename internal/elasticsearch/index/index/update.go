package index

import (
	"context"
	"maps"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.resourceReady(&resp.Diagnostics) {
		return
	}

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

	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, planModel.ElasticsearchConnection, r.client)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	name := planModel.Name.ValueString()
	id, sdkDiags := client.ID(ctx, name)
	if sdkDiags.HasError() {
		resp.Diagnostics.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
		return
	}

	planModel.ID = types.StringValue(id.String())
	planApiModel, diags := planModel.toAPIModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateApiModel, diags := stateModel.toAPIModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !planModel.Alias.Equal(stateModel.Alias) {
		resp.Diagnostics.Append(r.updateAliases(ctx, client, name, planApiModel.Aliases, stateApiModel.Aliases)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(r.updateSettings(ctx, client, name, planApiModel.Settings, stateApiModel.Settings)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.updateMappings(ctx, client, name, planModel.Mappings, stateModel.Mappings)...)

	finalModel, diags := readIndex(ctx, planModel, client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, finalModel)...)
}

func (r *Resource) updateAliases(ctx context.Context, client *clients.ApiClient, indexName string, planAliases map[string]models.IndexAlias, stateAliases map[string]models.IndexAlias) diag.Diagnostics {
	aliasesToDelete := []string{}
	for aliasName := range stateAliases {
		if _, ok := planAliases[aliasName]; !ok {
			aliasesToDelete = append(aliasesToDelete, aliasName)
		}
	}

	if len(aliasesToDelete) > 0 {
		diags := elasticsearch.DeleteIndexAlias(ctx, client, indexName, aliasesToDelete)
		if diags.HasError() {
			return diags
		}
	}

	for _, alias := range planAliases {
		diags := elasticsearch.UpdateIndexAlias(ctx, client, indexName, &alias)
		if diags.HasError() {
			return diags
		}
	}

	return nil
}

func (r *Resource) updateSettings(ctx context.Context, client *clients.ApiClient, indexName string, planSettings map[string]interface{}, stateSettings map[string]interface{}) diag.Diagnostics {
	planDynamicSettings := map[string]interface{}{}
	stateDynamicSettings := map[string]interface{}{}

	for _, key := range dynamicSettingsKeys {
		if planSetting, ok := planSettings[key]; ok {
			planDynamicSettings[key] = planSetting
		}

		if stateSetting, ok := stateSettings[key]; ok {
			stateDynamicSettings[key] = stateSetting
		}
	}

	if !maps.Equal(planDynamicSettings, stateDynamicSettings) {
		// Settings which are being removed must be explicitly set to null in the new settings
		for setting := range stateDynamicSettings {
			if _, ok := planDynamicSettings[setting]; !ok {
				planDynamicSettings[setting] = nil
			}
		}

		diags := elasticsearch.UpdateIndexSettings(ctx, client, indexName, planDynamicSettings)
		if diags.HasError() {
			return diags
		}
	}

	return nil
}

func (r *Resource) updateMappings(ctx context.Context, client *clients.ApiClient, indexName string, planMappings jsontypes.Normalized, stateMappings jsontypes.Normalized) diag.Diagnostics {
	areEqual, diags := planMappings.StringSemanticEquals(ctx, stateMappings)
	if diags.HasError() {
		return diags
	}

	if areEqual {
		return nil
	}

	diags = elasticsearch.UpdateIndexMappings(ctx, client, indexName, planMappings.ValueString())
	if diags.HasError() {
		return diags
	}

	return nil
}
