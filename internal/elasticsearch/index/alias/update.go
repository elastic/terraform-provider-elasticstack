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
	
	// Get the current indices from state for removal
	var currentIndices []string
	resp.Diagnostics.Append(stateModel.Indices.ElementsAs(ctx, &currentIndices, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the planned indices
	planAliasModel, planIndices, diags := planModel.toAPIModel()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// First, remove the alias from all current indices to ensure clean state
	if len(currentIndices) > 0 {
		resp.Diagnostics.Append(elasticsearch.DeleteAlias(ctx, r.client, aliasName, currentIndices)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Then add the alias to the new indices with the updated configuration
	resp.Diagnostics.Append(elasticsearch.PutAlias(ctx, r.client, aliasName, planIndices, &planAliasModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read back the alias to ensure state consistency, using planned model as input to preserve planned values
	finalModel, diags := readAliasWithPlan(ctx, r.client, aliasName, &planModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, finalModel)...)
}