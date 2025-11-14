package exception_item

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *ExceptionItemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ExceptionItemModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Kibana client", err.Error())
		return
	}

	// Delete by ID
	id := kbapi.SecurityExceptionsAPIExceptionListItemId(state.ID.ValueString())
	params := &kbapi.DeleteExceptionListItemParams{
		Id: &id,
	}

	diags = kibana_oapi.DeleteExceptionListItem(ctx, client, params)
	resp.Diagnostics.Append(diags...)
}
