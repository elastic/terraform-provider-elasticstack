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

	// Remove the alias from old indices that are not in the new plan
	var indicesToRemove []string
	planIndicesMap := make(map[string]bool)
	for _, idx := range planIndices {
		planIndicesMap[idx] = true
	}
	
	for _, idx := range currentIndices {
		if !planIndicesMap[idx] {
			indicesToRemove = append(indicesToRemove, idx)
		}
	}

	if len(indicesToRemove) > 0 {
		resp.Diagnostics.Append(elasticsearch.DeleteAlias(ctx, r.client, aliasName, indicesToRemove)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Update/create the alias with new configuration
	resp.Diagnostics.Append(elasticsearch.PutAlias(ctx, r.client, aliasName, planIndices, &planAliasModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read back the alias to ensure state consistency
	finalModel, diags := readAlias(ctx, r.client, aliasName)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, finalModel)...)
}