package api_key

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
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
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	finalModel, diags := r.read(ctx, client, stateModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if finalModel == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, *finalModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.saveClusterVersion(ctx, client, resp.Private)...)
}

func (r *Resource) read(ctx context.Context, client *clients.ApiClient, model tfModel) (*tfModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	compId, diags := model.GetID()
	if diags.HasError() {
		return nil, diags
	}

	apiKey, diags := elasticsearch.GetApiKey(client, compId.ResourceId)
	if diags.HasError() {
		return nil, diags
	}
	if apiKey == nil {
		return nil, nil
	}

	version, sdkDiags := client.ServerVersion(ctx)
	diags = utils.FrameworkDiagsFromSDK(sdkDiags)
	if diags.HasError() {
		return nil, diags
	}

	diags.Append(model.populateFromAPI(*apiKey, version)...)
	return &model, diags
}
