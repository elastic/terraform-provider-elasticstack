package data_stream_lifecycle

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.create(ctx, req.Plan, &resp.State, &resp.Diagnostics)
}

func (r Resource) create(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State, respDiags *diag.Diagnostics) {
	var planModel tfModel
	respDiags.Append(plan.Get(ctx, &planModel)...)
	if respDiags.HasError() {
		return
	}

	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, planModel.ElasticsearchConnection, r.client)
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return
	}

	name := planModel.Name.ValueString()
	id, sdkDiags := client.ID(ctx, name)
	if sdkDiags.HasError() {
		respDiags.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
		return
	}

	planModel.ID = types.StringValue(id.String())

	apiModel, diags := planModel.toAPIModel(ctx)
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return
	}

	respDiags.Append(elasticsearch.PutDataStreamLifecycle(ctx, client, name, planModel.ExpandWildcards.ValueString(), apiModel)...)
	if respDiags.HasError() {
		return
	}

	finalModel, diags := r.read(ctx, client, planModel)
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return
	}

	respDiags.Append(state.Set(ctx, finalModel)...)

}
