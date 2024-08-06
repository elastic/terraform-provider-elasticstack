package private_location

import (
	"context"
	"errors"
	"fmt"
	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {

	kibanaClient := synthetics.GetKibanaClient(r, response.Diagnostics)
	if kibanaClient == nil {
		return
	}

	var state tfModelV0
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	label := state.Label.ValueString()
	namespace := state.SpaceID.ValueString()
	result, err := kibanaClient.KibanaSynthetics.PrivateLocation.Get(label, namespace)
	if err != nil {
		var apiError *kbapi.APIError
		if errors.As(err, &apiError) && apiError.Code == 404 {
			response.State.RemoveResource(ctx)
			return
		}

		response.Diagnostics.AddError(fmt.Sprintf("Failed to get private location `%s`, namespace %s", label, namespace), err.Error())
		return
	}

	state = toModelV0(*result)

	// Set refreshed state
	diags = response.State.Set(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
}
