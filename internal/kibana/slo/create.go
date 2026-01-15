package slo

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	clientkibana "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan tfModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		response.Diagnostics.AddError("Provider not configured", "Expected configured API client")
		return
	}

	apiModel, diags := plan.toAPIModel()
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	serverVersion, sdkDiags := r.client.ServerVersion(ctx)
	response.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if response.Diagnostics.HasError() {
		return
	}

	if apiModel.Settings != nil && apiModel.Settings.PreventInitialBackfill != nil {
		if serverVersion.LessThan(SLOSupportsPreventInitialBackfillMinVersion) {
			response.Diagnostics.AddError(
				"Unsupported Elastic Stack version",
				"The 'prevent_initial_backfill' setting requires Elastic Stack version "+SLOSupportsPreventInitialBackfillMinVersion.String()+" or higher.",
			)
			return
		}
	}

	if plan.hasDataViewID() && serverVersion.LessThan(SLOSupportsDataViewIDMinVersion) {
		response.Diagnostics.AddError(
			"Unsupported Elastic Stack version",
			"data_view_id is not supported on Elastic Stack versions < "+SLOSupportsDataViewIDMinVersion.String(),
		)
		return
	}

	supportsGroupBy := serverVersion.GreaterThanOrEqual(SLOSupportsGroupByMinVersion)
	if !supportsGroupBy {
		if len(apiModel.GroupBy) > 0 {
			response.Diagnostics.AddError(
				"Unsupported Elastic Stack version",
				"group_by is not supported in this version of the Elastic Stack. group_by requires "+SLOSupportsGroupByMinVersion.String()+" or higher.",
			)
			return
		}
	}

	supportsMultipleGroupBy := supportsGroupBy && serverVersion.GreaterThanOrEqual(SLOSupportsMultipleGroupByMinVersion)
	if len(apiModel.GroupBy) > 1 && !supportsMultipleGroupBy {
		response.Diagnostics.AddError(
			"Unsupported Elastic Stack version",
			"multiple group_by fields are not supported in this version of the Elastic Stack. Multiple group_by fields requires "+SLOSupportsMultipleGroupByMinVersion.String(),
		)
		return
	}

	res, sdkDiags := clientkibana.CreateSlo(ctx, r.client, apiModel, supportsMultipleGroupBy)
	response.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if response.Diagnostics.HasError() {
		return
	}

	compositeID := (&clients.CompositeId{ClusterId: apiModel.SpaceID, ResourceId: res.SloID}).String()
	plan.ID = types.StringValue(compositeID)

	// Read back to populate computed fields.
	r.readAndPopulate(ctx, &plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}
