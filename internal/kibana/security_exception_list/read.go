package security_exception_list

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *ExceptionListResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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

	state.SpaceID = types.StringValue(compId.ClusterId)

	// Read by resource ID from composite ID
	id := kbapi.SecurityExceptionsAPIExceptionListId(compId.ResourceId)
	params := &kbapi.ReadExceptionListParams{
		Id: &id,
	}
	// Include namespace_type if specified (required for agnostic lists)
	if utils.IsKnown(state.NamespaceType) {
		nsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(state.NamespaceType.ValueString())
		params.NamespaceType = &nsType
	}

	readResp, diags := kibana_oapi.GetExceptionList(ctx, client, compId.ClusterId, params)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If namespace_type was not known (e.g., during import) and the list was not found,
	// try reading with namespace_type=agnostic
	if readResp == nil && !utils.IsKnown(state.NamespaceType) {
		agnosticNsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType("agnostic")
		params.NamespaceType = &agnosticNsType
		readResp, diags = kibana_oapi.GetExceptionList(ctx, client, compId.ClusterId, params)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if readResp == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state with response using model method
	diags = state.fromAPI(ctx, readResp)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
