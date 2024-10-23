package data_stream_lifecycle

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
}

func (r *Resource) read(ctx context.Context, client *clients.ApiClient, model tfModel) (*tfModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	compId, diags := model.GetID()
	if diags.HasError() {
		return nil, diags
	}

	ds, diags := elasticsearch.GetDataStreamLifecycle(ctx, client, compId.ResourceId, model.ExpandWildcards.ValueString())
	if diags.HasError() {
		return nil, diags
	}
	if ds == nil || len(*ds) == 0 {
		return nil, nil
	}

	diags.Append(model.populateFromAPI(ctx, *ds)...)
	return &model, diags
}
