package alias

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *aliasResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateModel tfModel

	resp.Diagnostics.Append(req.State.Get(ctx, &stateModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	aliasName := stateModel.Name.ValueString()

	// Get the alias
	indices, diags := elasticsearch.GetAlias(ctx, r.client, aliasName)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If no indices returned, the alias doesn't exist
	if indices == nil || len(indices) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	// Extract indices and alias data from the response
	var indexNames []string
	var aliasData *models.IndexAlias
	
	for indexName, index := range indices {
		if alias, exists := index.Aliases[aliasName]; exists {
			indexNames = append(indexNames, indexName)
			if aliasData == nil {
				// Use the first alias definition we find (they should all be the same)
				aliasData = &alias
			}
		}
	}

	if aliasData == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update the state model
	resp.Diagnostics.Append(stateModel.populateFromAPI(ctx, aliasName, *aliasData, indexNames)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, stateModel)...)
}