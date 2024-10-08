package index

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var planModel tfModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, planModel.ElasticsearchConnection, r.client)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	name := planModel.Name.ValueString()
	id, sdkDiags := client.ID(ctx, name)
	if sdkDiags.HasError() {
		resp.Diagnostics.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
		return
	}

	planModel.ID = types.StringValue(id.String())
	apiModel, diags := planModel.toAPIModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverFlavor, sdkDiags := client.ServerFlavor(ctx)
	if sdkDiags.HasError() {
		resp.Diagnostics.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
		return
	}

	params := planModel.toPutIndexParams(serverFlavor)

	resp.Diagnostics.Append(elasticsearch.PutIndex(ctx, client, &apiModel, &params)...)
	if resp.Diagnostics.HasError() {
		return
	}

	finalModel, diags := readIndex(ctx, planModel, client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, finalModel)...)
}
