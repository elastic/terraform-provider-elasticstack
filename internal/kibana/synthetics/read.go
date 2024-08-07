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

	var state *tfModelV0
	diags := request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	namespace := state.SpaceID.ValueString()
	monitorId := state.ID.ValueString()
	result, err := kibanaClient.KibanaSynthetics.Monitor.Get(kbapi.MonitorID(monitorId), namespace)
	if err != nil {
		var apiError *kbapi.APIError
		if errors.As(err, &apiError) && apiError.Code == 404 {
			response.State.RemoveResource(ctx)
			return
		}

		response.Diagnostics.AddError(fmt.Sprintf("Failed to get monitor `%s`, namespace %s", monitorId, namespace), err.Error())
		return
	}

	state, err = toModelV0(result)
	if err != nil {
		response.Diagnostics.AddError("Failed to convert Kibana monitor API to TF state", err.Error())
		return
	}

	// Set refreshed state
	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
}
