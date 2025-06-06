package parameter

import (
	"context"
	"errors"
	"fmt"

	"github.com/disaster37/go-kibana-rest/v8"
	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func (r *Resource) readState(ctx context.Context, kibanaClient *kibana.Client, model tfModelV0, state *tfsdk.State, diagnostics *diag.Diagnostics) {
	resourceId := model.ID.ValueString()

	compositeId, dg := tryReadCompositeId(resourceId)
	diagnostics.Append(dg...)
	if diagnostics.HasError() {
		return
	}

	if compositeId != nil {
		resourceId = compositeId.ResourceId
	}

	result, err := kibanaClient.KibanaSynthetics.Parameter.Get(ctx, resourceId)
	if err != nil {
		var apiError *kbapi.APIError
		if errors.As(err, &apiError) && apiError.Code == 404 {
			state.RemoveResource(ctx)
			return
		}

		diagnostics.AddError(fmt.Sprintf("Failed to get parameter `%s`", resourceId), err.Error())
		return
	}

	model = toModelV0(*result)

	// Set refreshed state
	diags := state.Set(ctx, &model)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return
	}
}

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

	r.readState(ctx, kibanaClient, state, &response.State, &response.Diagnostics)
}
