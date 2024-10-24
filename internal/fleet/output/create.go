package output

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *outputResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var planModel outputModel

	diags := req.Plan.Get(ctx, &planModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	body, diags := planModel.toAPICreateModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, diags := fleet.CreateOutput(ctx, client, body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = planModel.populateFromAPI(ctx, output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, planModel)
	resp.Diagnostics.Append(diags...)
}
