package exception_list

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

	// Build the request body
	body := kbapi.CreateExceptionListJSONRequestBody{
		ListId:      (*kbapi.SecurityExceptionsAPIExceptionListHumanId)(plan.ListID.ValueStringPointer()),
		Name:        kbapi.SecurityExceptionsAPIExceptionListName(plan.Name.ValueString()),
		Description: kbapi.SecurityExceptionsAPIExceptionListDescription(plan.Description.ValueString()),
		Type:        kbapi.SecurityExceptionsAPIExceptionListType(plan.Type.ValueString()),
	}

	// Set optional namespace_type
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

	// Create the exception list
	createResp, diags := kibana_oapi.CreateExceptionList(ctx, client, body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if createResp == nil || createResp.JSON200 == nil {
		resp.Diagnostics.AddError("Failed to create exception list", "API returned empty response")
		return
	}

	// Update state with create response
	diags = r.updateStateFromAPIResponse(ctx, &plan, createResp.JSON200)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ExceptionListResource) updateStateFromAPIResponse(ctx context.Context, model *ExceptionListModel, apiResp *kbapi.SecurityExceptionsAPIExceptionList) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue(string(apiResp.Id))
	model.ListID = types.StringValue(string(apiResp.ListId))
	model.Name = types.StringValue(string(apiResp.Name))
	model.Description = types.StringValue(string(apiResp.Description))
	model.Type = types.StringValue(string(apiResp.Type))
	model.NamespaceType = types.StringValue(string(apiResp.NamespaceType))
	model.CreatedAt = types.StringValue(apiResp.CreatedAt.Format("2006-01-02T15:04:05.000Z"))
	model.CreatedBy = types.StringValue(apiResp.CreatedBy)
	model.UpdatedAt = types.StringValue(apiResp.UpdatedAt.Format("2006-01-02T15:04:05.000Z"))
	model.UpdatedBy = types.StringValue(apiResp.UpdatedBy)
	model.Immutable = types.BoolValue(apiResp.Immutable)
	model.TieBreakerID = types.StringValue(apiResp.TieBreakerId)

	// Set optional os_types
	if apiResp.OsTypes != nil && len(*apiResp.OsTypes) > 0 {
		osTypes := make([]string, len(*apiResp.OsTypes))
		for i, osType := range *apiResp.OsTypes {
			osTypes[i] = string(osType)
		}
		list, d := types.ListValueFrom(ctx, types.StringType, osTypes)
		diags.Append(d...)
		model.OsTypes = list
	} else {
		model.OsTypes = types.ListNull(types.StringType)
	}

	// Set optional tags
	if apiResp.Tags != nil && len(*apiResp.Tags) > 0 {
		list, d := types.ListValueFrom(ctx, types.StringType, *apiResp.Tags)
		diags.Append(d...)
		model.Tags = list
	} else {
		model.Tags = types.ListNull(types.StringType)
	}

	// Set optional meta
	if apiResp.Meta != nil {
		metaJSON, err := json.Marshal(apiResp.Meta)
		if err != nil {
			diags.AddError("Failed to serialize meta", err.Error())
			return diags
		}
		model.Meta = types.StringValue(string(metaJSON))
	} else {
		model.Meta = types.StringNull()
	}

	return diags
}
