package alias

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *aliasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var planModel tfModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &planModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	aliasModel, indices, diags := planModel.toAPIModel()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	aliasName := planModel.Name.ValueString()

	// Create the alias
	resp.Diagnostics.Append(elasticsearch.PutAlias(ctx, r.client, aliasName, indices, &aliasModel)...)
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

func readAlias(ctx context.Context, client *clients.ApiClient, aliasName string) (*tfModel, diag.Diagnostics) {
	indices, diags := elasticsearch.GetAlias(ctx, client, aliasName)
	if diags.HasError() {
		return nil, diags
	}

	if indices == nil || len(indices) == 0 {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Alias not found after creation",
				"The alias was not found after creation, which indicates an error in the Elasticsearch API response.",
			),
		}
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
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Alias data not found after creation",
				"The alias data was not found after creation, which indicates an error in the Elasticsearch API response.",
			),
		}
	}

	finalModel := &tfModel{}
	diags = finalModel.populateFromAPI(ctx, aliasName, *aliasData, indexNames)
	if diags.HasError() {
		return nil, diags
	}

	return finalModel, nil
}