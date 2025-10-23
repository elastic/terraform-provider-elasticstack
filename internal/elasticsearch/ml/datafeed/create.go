package datafeed

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *datafeedResource) create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.resourceReady(&resp.Diagnostics) {
		return
	}

	var plan Datafeed
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	datafeedId := plan.DatafeedID.ValueString()
	if datafeedId == "" {
		resp.Diagnostics.AddError("Invalid Configuration", "datafeed_id cannot be empty")
		return
	}

	// Convert to API create model
	createRequest, diags := plan.toAPICreateModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createDiags := elasticsearch.PutDatafeed(ctx, r.client, datafeedId, *createRequest)
	resp.Diagnostics.Append(createDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the created datafeed to get the full state.
	compID, sdkDiags := r.client.ID(ctx, datafeedId)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(compID.String())
	found, readDiags := r.read(ctx, &plan)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Failed to read created datafeed", "Datafeed not found after creation")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
