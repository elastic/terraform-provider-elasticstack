package alias

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func (r *aliasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var planModel tfModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &planModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	aliasName := planModel.Name.ValueString()

	// Set the ID using client.ID
	id, sdkDiags := r.client.ID(ctx, aliasName)
	if sdkDiags.HasError() {
		resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		return
	}
	planModel.ID = basetypes.NewStringValue(id.String())

	// Get alias configurations from the plan
	configs, diags := planModel.toAliasConfigs(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to alias actions
	var actions []elasticsearch.AliasAction
	for _, config := range configs {
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

	// Create the alias atomically
	resp.Diagnostics.Append(elasticsearch.UpdateAliasesAtomic(ctx, r.client, actions)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read back the alias to ensure state consistency, updating the current model
	diags = readAliasIntoModel(ctx, r.client, aliasName, &planModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, planModel)...)
}