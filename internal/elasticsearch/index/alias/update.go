package alias

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *aliasResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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

	aliasName := planModel.Name.ValueString()

	// Get current configuration from state
	currentConfigs, diags := stateModel.toAliasConfigs(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get planned configuration
	plannedConfigs, diags := planModel.toAliasConfigs(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build atomic actions
	var actions []elasticsearch.AliasAction

	// Create maps for easy lookup
	currentIndexMap := make(map[string]AliasIndexConfig)
	for _, config := range currentConfigs {
		currentIndexMap[config.Name] = config
	}

	plannedIndexMap := make(map[string]AliasIndexConfig)
	for _, config := range plannedConfigs {
		plannedIndexMap[config.Name] = config
	}

	// Remove indices that are no longer in the plan
	for indexName := range currentIndexMap {
		if _, exists := plannedIndexMap[indexName]; !exists {
			actions = append(actions, elasticsearch.AliasAction{
				Type:  "remove",
				Index: indexName,
				Alias: aliasName,
			})
		}
	}

	// Add or update indices in the plan
	for _, config := range plannedConfigs {
		action := elasticsearch.AliasAction{
			Type:          "add",
			Index:         config.Name,
			Alias:         aliasName,
			IsWriteIndex:  config.IsWriteIndex,
			Filter:        config.Filter,
			IndexRouting:  config.IndexRouting,
			IsHidden:      config.IsHidden,
			Routing:       config.Routing,
			SearchRouting: config.SearchRouting,
		}
		actions = append(actions, action)
	}

	// Apply the atomic changes
	if len(actions) > 0 {
		resp.Diagnostics.Append(elasticsearch.UpdateAliasesAtomic(ctx, r.client, actions)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Read back the alias to ensure state consistency, updating the current model
	diags = readAliasIntoModel(ctx, r.client, aliasName, &planModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, planModel)...)
}
