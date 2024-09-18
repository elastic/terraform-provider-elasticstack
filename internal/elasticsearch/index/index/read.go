package index

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateModel tfModel
	resp.Diagnostics.Append(req.State.Get(ctx, &stateModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, stateModel.ElasticsearchConnection, r.client)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	model, diags := readIndex(ctx, stateModel, client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if model == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func readIndex(ctx context.Context, stateModel tfModel, client *clients.ApiClient) (*tfModel, diag.Diagnostics) {
	id, diags := stateModel.GetID()
	if diags.HasError() {
		return nil, diags
	}

	indexName := id.ResourceId
	apiModel, diags := elasticsearch.GetIndex(ctx, client, indexName)
	if diags.HasError() {
		return nil, diags
	}

	if apiModel == nil {
		return nil, nil
	}

	diags = stateModel.populateFromAPI(ctx, indexName, *apiModel)
	if diags.HasError() {
		return nil, diags
	}

	return &stateModel, nil
}
