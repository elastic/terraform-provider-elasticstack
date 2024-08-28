package indices

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Read refreshes the Terraform state with the latest data.
func (d *dataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var stateModel tfModel

	// Resolve search attribute
	var search string
	diags := req.Config.GetAttribute(ctx, path.Root("search"), &search)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call client API
	indexApiModels, sdkDiags := elasticsearch.GetIndices(ctx, &d.client, search)
	resp.Diagnostics.Append(utils.ConvertSDKDiagnosticsToFramework(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map response body to model
	indices := []indexTfModel{}
	for indexName, indexApiModel := range indexApiModels {
		indexStateModel := indexTfModel{}

		diags := indexStateModel.populateFromAPI(ctx, indexName, indexApiModel)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		indices = append(indices, indexStateModel)
	}

	indicesList, diags := types.ListValueFrom(ctx, indicesElementType(), indices)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateModel.ID = types.StringValue(search)
	stateModel.Indices = indicesList

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, stateModel)...)
}
