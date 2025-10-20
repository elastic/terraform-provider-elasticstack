package monitor

import (
	"context"
	"fmt"

	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	kibanaClient := synthetics.GetKibanaClient(r, response.Diagnostics)
	if kibanaClient == nil {
		return
	}

	plan := new(tfModelV0)
	diags := request.Plan.Get(ctx, plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(plan.enforceVersionConstraints(ctx, r.client)...)
	if response.Diagnostics.HasError() {
		return
	}

	input, diags := plan.toKibanaAPIRequest(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	monitorId, dg := synthetics.GetCompositeId(plan.ID.ValueString())
	response.Diagnostics.Append(dg...)
	if response.Diagnostics.HasError() {
		return
	}

	spaceId := plan.SpaceID.ValueString()
	result, err := kibanaClient.KibanaSynthetics.Monitor.Update(ctx, kbapi.MonitorID(monitorId.ResourceId), input.config, input.fields, spaceId)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to update Kibana monitor `%s`, space %s", input.config.Name, spaceId), err.Error())
		return
	}

	plan, diags = plan.toModelV0(ctx, result, spaceId)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	diags = response.State.Set(ctx, plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
}
