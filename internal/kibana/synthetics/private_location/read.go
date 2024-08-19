package private_location

import (
	"context"
	"errors"
	"fmt"
	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"strings"
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

	resourceId := state.ID.ValueString()
	namespace := state.SpaceID.ValueString()
	if strings.Contains(resourceId, "/") {
		compositeId, dg := synthetics.GetCompositeId(state.ID.ValueString())
		response.Diagnostics.Append(dg...)
		if response.Diagnostics.HasError() {
			return
		}

		resourceId = compositeId.ResourceId
		namespace = compositeId.ClusterId
	} else {
		resourceId = state.Label.ValueString()
	}

	result, err := kibanaClient.KibanaSynthetics.PrivateLocation.Get(ctx, resourceId, namespace)
	if err != nil {
		var apiError *kbapi.APIError
		if errors.As(err, &apiError) && apiError.Code == 404 {
			response.State.RemoveResource(ctx)
			return
		}

		response.Diagnostics.AddError(fmt.Sprintf("Failed to get private location `%s`, namespace %s", resourceId, namespace), err.Error())
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
