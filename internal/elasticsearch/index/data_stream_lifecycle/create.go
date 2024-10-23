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
	resp.Diagnostics.Append(r.create(ctx, req.Plan, &resp.State)...)
}

func (r Resource) create(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State) diag.Diagnostics {
	var planModel tfModel
	diags := plan.Get(ctx, &planModel)
	if diags.HasError() {
		return diags
	}

	client, d := clients.MaybeNewApiClientFromFrameworkResource(ctx, planModel.ElasticsearchConnection, r.client)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	name := planModel.Name.ValueString()
	id, sdkDiags := client.ID(ctx, name)
	if sdkDiags.HasError() {
		diags.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
		return diags
	}

	planModel.ID = types.StringValue(id.String())

	apiModel, d := planModel.toAPIModel(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	diags.Append(elasticsearch.PutDataStreamLifecycle(ctx, client, name, planModel.ExpandWildcards.ValueString(), apiModel)...)
	if diags.HasError() {
		return diags
	}

	finalModel, d := r.read(ctx, client, planModel)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	diags.Append(state.Set(ctx, finalModel)...)
	return diags
}
