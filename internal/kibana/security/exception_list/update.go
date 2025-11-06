package exception_list

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
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

	// Build the update request body
	id := kbapi.SecurityExceptionsAPIExceptionListId(plan.ID.ValueString())
	body := kbapi.UpdateExceptionListJSONRequestBody{
		Id:          &id,
		Name:        kbapi.SecurityExceptionsAPIExceptionListName(plan.Name.ValueString()),
		Description: kbapi.SecurityExceptionsAPIExceptionListDescription(plan.Description.ValueString()),
	}

	// Set optional namespace_type (should not change, but include it)
	if utils.IsKnown(plan.NamespaceType) {
		nsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(plan.NamespaceType.ValueString())
		body.NamespaceType = &nsType
	}

	// Set optional os_types
	if utils.IsKnown(plan.OsTypes) && !plan.OsTypes.IsNull() {
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
	if utils.IsKnown(plan.Tags) && !plan.Tags.IsNull() {
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
	if utils.IsKnown(plan.Meta) && !plan.Meta.IsNull() {
		var meta kbapi.SecurityExceptionsAPIExceptionListMeta
		if err := json.Unmarshal([]byte(plan.Meta.ValueString()), &meta); err != nil {
			resp.Diagnostics.AddError("Failed to parse meta JSON", err.Error())
			return
		}
		body.Meta = &meta
	}

	// Update the exception list
	updateResp, diags := kibana_oapi.UpdateExceptionList(ctx, client, body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if updateResp == nil || updateResp.JSON200 == nil {
		resp.Diagnostics.AddError("Failed to update exception list", "API returned empty response")
		return
	}

	// Update state with response
	diags = r.updateStateFromAPIResponse(ctx, &plan, updateResp.JSON200)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}
