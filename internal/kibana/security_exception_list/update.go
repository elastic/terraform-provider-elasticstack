package security_exception_list

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *ExceptionListResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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

	// Parse composite ID to get space_id and resource_id
	compId, compIdDiags := clients.CompositeIdFromStrFw(plan.ID.ValueString())
	resp.Diagnostics.Append(compIdDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the update request body
	id := kbapi.SecurityExceptionsAPIExceptionListId(compId.ResourceId)
	body := kbapi.UpdateExceptionListJSONRequestBody{
		Id:          &id,
		Name:        kbapi.SecurityExceptionsAPIExceptionListName(plan.Name.ValueString()),
		Description: kbapi.SecurityExceptionsAPIExceptionListDescription(plan.Description.ValueString()),
		// Type is required by the API even though it has RequiresReplace in the schema
		// The API will reject updates without this field, even though the value cannot change
		Type: kbapi.SecurityExceptionsAPIExceptionListType(plan.Type.ValueString()),
	}

	// Set optional namespace_type (should not change, but include it)
	if utils.IsKnown(plan.NamespaceType) {
		nsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(plan.NamespaceType.ValueString())
		body.NamespaceType = &nsType
	}

	// Set optional os_types
	if utils.IsKnown(plan.OsTypes) {
		var osTypes []string
		diags := plan.OsTypes.ElementsAs(ctx, &osTypes, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(osTypes) > 0 {
			osTypesArray := make(kbapi.SecurityExceptionsAPIExceptionListOsTypeArray, len(osTypes))
			for i, osType := range osTypes {
				osTypesArray[i] = kbapi.SecurityExceptionsAPIExceptionListOsType(osType)
			}
			body.OsTypes = &osTypesArray
		}
	}

	// Set optional tags
	if utils.IsKnown(plan.Tags) {
		var tags []string
		diags := plan.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(tags) > 0 {
			tagsArray := kbapi.SecurityExceptionsAPIExceptionListTags(tags)
			body.Tags = &tagsArray
		}
	}

	// Set optional meta
	if utils.IsKnown(plan.Meta) {
		var meta kbapi.SecurityExceptionsAPIExceptionListMeta
		unmarshalDiags := plan.Meta.Unmarshal(&meta)
		resp.Diagnostics.Append(unmarshalDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		body.Meta = &meta
	}

	// Update the exception list
	updateResp, diags := kibana_oapi.UpdateExceptionList(ctx, client, plan.SpaceID.ValueString(), body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if updateResp == nil {
		resp.Diagnostics.AddError("Failed to update exception list", "API returned empty response")
		return
	}

	/*
	 * In create/update paths we typically follow the write operation with a read, and then set the state from the read.
	 * We want to avoid a dirty plan immediately after an apply.
	 */
	// Read back the updated resource to get the final state
	readParams := &kbapi.ReadExceptionListParams{
		Id: (*kbapi.SecurityExceptionsAPIExceptionListId)(&updateResp.Id),
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

	// Update state with read response
	diags = r.updateStateFromAPIResponse(ctx, &plan, readResp)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}
