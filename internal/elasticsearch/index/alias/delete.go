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

	// Get the current indices from state
	var indices []string
	resp.Diagnostics.Append(stateModel.Indices.ElementsAs(ctx, &indices, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the alias from all indices
	resp.Diagnostics.Append(elasticsearch.DeleteAlias(ctx, r.client, aliasName, indices)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
