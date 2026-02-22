package securityexceptionlist

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
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
	compID, compIDDiags := clients.CompositeIDFromStrFw(state.ID.ValueString())
	resp.Diagnostics.Append(compIDDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.SpaceID = types.StringValue(compID.ClusterID)

	// Read by resource ID from composite ID
	id := compID.ResourceID
	params := &kbapi.ReadExceptionListParams{
		Id: &id,
	}
	// Include namespace_type if specified (required for agnostic lists)
	if typeutils.IsKnown(state.NamespaceType) {
		nsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(state.NamespaceType.ValueString())
		params.NamespaceType = &nsType
	}

	readResp, diags := kibanaoapi.GetExceptionList(ctx, client, compID.ClusterID, params)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If namespace_type was not known (e.g., during import) and the list was not found,
	// try reading with namespace_type=agnostic
	if readResp == nil && !typeutils.IsKnown(state.NamespaceType) {
		agnosticNsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType("agnostic")
		params.NamespaceType = &agnosticNsType
		readResp, diags = kibanaoapi.GetExceptionList(ctx, client, compID.ClusterID, params)
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
