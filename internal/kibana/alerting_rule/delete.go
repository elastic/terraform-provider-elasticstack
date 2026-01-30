package alerting_rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state tfModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, state.KibanaConnection, r.client)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	compositeID, idDiags := state.GetID()
	response.Diagnostics.Append(idDiags...)
	if response.Diagnostics.HasError() {
		return
	}

	spaceId := state.SpaceID.ValueString()

	sdkDiags := kibana.DeleteAlertingRule(ctx, client, compositeID.ResourceId, spaceId)
	if sdkDiags.HasError() {
		for _, d := range sdkDiags {
			response.Diagnostics.AddError(d.Summary, d.Detail)
		}
	}
}
