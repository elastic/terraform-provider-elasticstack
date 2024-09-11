package index

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.resourceReady(&resp.Diagnostics) {
		return
	}

	var model tfModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if model.DeletionProtection.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			path.Root("deletion_protection"),
			"cannot destroy index without setting deletion_protection=false and running `terraform apply`",
			"cannot destroy index without setting deletion_protection=false and running `terraform apply`",
		)
		return
	}

	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, model.ElasticsearchConnection, r.client)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	id, diags := model.GetID()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(elasticsearch.DeleteIndex(ctx, client, id.ResourceId)...)
}
