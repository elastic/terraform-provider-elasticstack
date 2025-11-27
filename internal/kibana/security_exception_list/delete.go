package security_exception_list

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *ExceptionListResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ExceptionListModel

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

	// Parse composite ID to get space_id and resource_id
	compId, compIdDiags := clients.CompositeIdFromStrFw(state.ID.ValueString())
	resp.Diagnostics.Append(compIdDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete by resource ID from composite ID
	id := kbapi.SecurityExceptionsAPIExceptionListId(compId.ResourceId)
	params := &kbapi.DeleteExceptionListParams{
		Id: &id,
	}

	diags = kibana_oapi.DeleteExceptionList(ctx, client, compId.ClusterId, params)
	resp.Diagnostics.Append(diags...)
}
