package slo

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	clientkibana "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state tfModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		response.Diagnostics.AddError("Provider not configured", "Expected configured API client")
		return
	}

	exists, diags := r.readSloFromAPI(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if !exists {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (r *Resource) readSloFromAPI(ctx context.Context, state *tfModel) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	compID, idDiags := clients.CompositeIdFromStrFw(state.ID.ValueString())
	diags.Append(idDiags...)
	if diags.HasError() {
		return false, diags
	}

	apiModel, sdkDiags := clientkibana.GetSlo(ctx, r.client, compID.ResourceId, compID.ClusterId)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return false, diags
	}
	if apiModel == nil {
		return false, diags
	}

	state.ID = types.StringValue((&clients.CompositeId{ClusterId: apiModel.SpaceID, ResourceId: apiModel.SloID}).String())
	diags.Append(state.populateFromAPI(apiModel)...)
	if diags.HasError() {
		return true, diags
	}

	return true, diags
}

func (r *Resource) readAndPopulate(ctx context.Context, plan *tfModel, diags *diag.Diagnostics) {
	exists, readDiags := r.readSloFromAPI(ctx, plan)
	diags.Append(readDiags...)
	if diags.HasError() {
		return
	}
	if !exists {
		diags.AddError("SLO not found", "SLO was created/updated but could not be found afterwards")
		return
	}
}
