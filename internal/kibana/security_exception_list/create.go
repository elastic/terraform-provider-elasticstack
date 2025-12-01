package security_exception_list

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *ExceptionListResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ExceptionListModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Kibana client", err.Error())
		return
	}

	// Build the request body using model method
	body, diags := plan.toCreateRequest(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the exception list
	createResp, diags := kibana_oapi.CreateExceptionList(ctx, client, plan.SpaceID.ValueString(), *body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if createResp == nil {
		resp.Diagnostics.AddError("Failed to create exception list", "API returned empty response")
		return
	}

	/*
	 * In create/update paths we typically follow the write operation with a read, and then set the state from the read.
	 * We want to avoid a dirty plan immediately after an apply.
	 */
	// Read back the created resource to get the final state
	readParams := &kbapi.ReadExceptionListParams{
		Id: (*kbapi.SecurityExceptionsAPIExceptionListId)(&createResp.Id),
	}

	// Include namespace_type if specified (required for agnostic lists)
	if utils.IsKnown(plan.NamespaceType) {
		nsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(createResp.NamespaceType)
		readParams.NamespaceType = &nsType
	}

	readResp, diags := kibana_oapi.GetExceptionList(ctx, client, plan.SpaceID.ValueString(), readParams)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if readResp == nil {
		resp.State.RemoveResource(ctx)
		resp.Diagnostics.AddError("Failed to fetch exception list", "API returned empty response")
		return
	}

	// Update state with read response using model method
	diags = plan.fromAPI(ctx, readResp)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}
