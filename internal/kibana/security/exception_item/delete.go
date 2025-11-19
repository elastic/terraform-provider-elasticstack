package exception_item

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *exceptionItemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state exceptionItemModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract composite ID
	compId, diags := clients.CompositeIdFromStrFw(state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Kibana client
	kibanaClient, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get Kibana client",
			fmt.Sprintf("Unable to get Kibana client: %s", err),
		)
		return
	}

	// Build delete parameters
	itemID := compId.ResourceId
	nsType := "single"
	if utils.IsKnown(state.NamespaceType) {
		nsType = state.NamespaceType.ValueString()
	}
	nsTypeAPI := kbapi.SecurityExceptionsAPIExceptionNamespaceType(nsType)

	params := kbapi.DeleteExceptionListItemParams{
		ItemId:        &itemID,
		NamespaceType: &nsTypeAPI,
	}

	// Make API call
	diags = kibana_oapi.DeleteExceptionListItem(ctx, kibanaClient, &params)
	resp.Diagnostics.Append(diags...)
}
