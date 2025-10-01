package alias

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *aliasResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var stateModel tfModel

	resp.Diagnostics.Append(req.State.Get(ctx, &stateModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	aliasName := stateModel.Name.ValueString()

	// Get current configuration from state
	currentConfigs, diags := stateModel.toAliasConfigs(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build remove actions for all indices
	var actions []elasticsearch.AliasAction
	for _, config := range currentConfigs {
		actions = append(actions, elasticsearch.AliasAction{
			Type:  "remove",
			Index: config.Name,
			Alias: aliasName,
		})
	}

	// Remove the alias from all indices
	if len(actions) > 0 {
		resp.Diagnostics.Append(elasticsearch.UpdateAliasesAtomic(ctx, r.client, actions)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
}
