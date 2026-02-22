package apikey

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

	client, diags := clients.MaybeNewAPIClientFromFrameworkResource(ctx, planModel.ElasticsearchConnection, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if planModel.Type.ValueString() == "cross_cluster" {
		updateDiags := r.updateCrossClusterAPIKey(ctx, client, planModel)
		resp.Diagnostics.Append(updateDiags...)
	} else {
		updateDiags := r.updateAPIKey(ctx, client, planModel)
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

func (r *Resource) updateCrossClusterAPIKey(ctx context.Context, client *clients.APIClient, planModel tfModel) diag.Diagnostics {
	// Handle cross-cluster API key update
	crossClusterModel, modelDiags := planModel.toCrossClusterAPIModel(ctx)
	if modelDiags.HasError() {
		return modelDiags
	}

	updateDiags := elasticsearch.UpdateCrossClusterAPIKey(client, crossClusterModel)
	return updateDiags
}

func (r *Resource) updateAPIKey(ctx context.Context, client *clients.APIClient, planModel tfModel) diag.Diagnostics {
	// Handle regular API key update
	apiModel, modelDiags := r.buildAPIModel(ctx, planModel, client)
	if modelDiags.HasError() {
		return modelDiags
	}

	updateDiags := elasticsearch.UpdateAPIKey(client, apiModel)
	return updateDiags
}
