package api_key

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planModel tfModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, planModel.ElasticsearchConnection, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if planModel.Type.ValueString() == "cross_cluster" {
		updateDiags := r.updateCrossClusterApiKey(ctx, client, planModel)
		resp.Diagnostics.Append(updateDiags...)
	} else {
		updateDiags := r.updateApiKey(ctx, client, planModel)
		resp.Diagnostics.Append(updateDiags...)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	finalModel, diags := r.read(ctx, client, planModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, *finalModel)...)
}

func (r *Resource) updateCrossClusterApiKey(ctx context.Context, client *clients.ApiClient, planModel tfModel) diag.Diagnostics {
	// Handle cross-cluster API key update
	crossClusterModel, modelDiags := planModel.toCrossClusterAPIModel(ctx)
	if modelDiags.HasError() {
		return modelDiags
	}

	updateDiags := elasticsearch.UpdateCrossClusterApiKey(client, crossClusterModel)
	return diag.Diagnostics(updateDiags)
}

func (r *Resource) updateApiKey(ctx context.Context, client *clients.ApiClient, planModel tfModel) diag.Diagnostics {
	// Handle regular API key update
	apiModel, modelDiags := r.buildApiModel(ctx, planModel, client)
	if modelDiags.HasError() {
		return modelDiags
	}

	updateDiags := elasticsearch.UpdateApiKey(client, apiModel)
	return diag.Diagnostics(updateDiags)
}
