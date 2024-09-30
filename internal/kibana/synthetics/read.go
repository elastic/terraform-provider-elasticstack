package synthetics

import (
	"context"
	"errors"
	"fmt"
	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {

	kibanaClient := GetKibanaClient(r, response.Diagnostics)
	if kibanaClient == nil {
		return
	}

	state := new(tfModelV0)
	diags := request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	compositeId, dg := GetCompositeId(state.ID.ValueString())
	response.Diagnostics.Append(dg...)
	if response.Diagnostics.HasError() {
		return
	}

	namespace := compositeId.ClusterId
	monitorId := kbapi.MonitorID(compositeId.ResourceId)
	result, err := kibanaClient.KibanaSynthetics.Monitor.Get(ctx, monitorId, namespace)
	if err != nil {
		var apiError *kbapi.APIError
		if errors.As(err, &apiError) && apiError.Code == 404 {
			response.State.RemoveResource(ctx)
			return
		}

		response.Diagnostics.AddError(fmt.Sprintf("Failed to get monitor `%s`, namespace %s", monitorId, namespace), err.Error())
		return
	}

	state, diags = state.toModelV0(ctx, result)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Set refreshed state
	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
}
