package apikey

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateModel tfModel
	resp.Diagnostics.Append(req.State.Get(ctx, &stateModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := clients.MaybeNewAPIClientFromFrameworkResource(ctx, stateModel.ElasticsearchConnection, r.client)
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

func (r *Resource) read(ctx context.Context, client *clients.APIClient, model tfModel) (*tfModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	compID, diags := model.GetID()
	if diags.HasError() {
		return nil, diags
	}

	apiKey, diags := elasticsearch.GetAPIKey(client, compID.ResourceID)
	if diags.HasError() {
		return nil, diags
	}
	if apiKey == nil {
		return nil, nil
	}

	version, sdkDiags := client.ServerVersion(ctx)
	diags = diagutil.FrameworkDiagsFromSDK(sdkDiags)
	if diags.HasError() {
		return nil, diags
	}

	diags.Append(model.populateFromAPI(*apiKey, version)...)
	return &model, diags
}
