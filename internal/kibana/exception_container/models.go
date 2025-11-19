package exception_container

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ExceptionContainerData struct {
	ID                types.String `tfsdk:"id"`
	KibanaConnection  types.List   `tfsdk:"kibana_connection"`
	SpaceID           types.String `tfsdk:"space_id"`
	ListID            types.String `tfsdk:"list_id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	Type              types.String `tfsdk:"type"`
	NamespaceType     types.String `tfsdk:"namespace_type"`
	Tags              types.List   `tfsdk:"tags"`
	OsTypes           types.List   `tfsdk:"os_types"`
}

func (model ExceptionContainerData) GetID() (*clients.CompositeId, diag.Diagnostics) {
	compId, sdkDiags := clients.CompositeIdFromStr(model.ID.ValueString())
	if sdkDiags.HasError() {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic("Invalid ID", sdkDiags[0].Summary),
		}
	}
	return compId, nil
}

func (model ExceptionContainerData) toAPICreateRequest(ctx context.Context) (kbapi.CreateExceptionListJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	body := kbapi.CreateExceptionListJSONRequestBody{
		Name:        kbapi.SecurityExceptionsAPIExceptionListName(model.Name.ValueString()),
		Description: kbapi.SecurityExceptionsAPIExceptionListDescription(model.Description.ValueString()),
		Type:        kbapi.SecurityExceptionsAPIExceptionListType(model.Type.ValueString()),
	}

	if utils.IsKnown(model.ListID) {
		listID := kbapi.SecurityExceptionsAPIExceptionListHumanId(model.ListID.ValueString())
		body.ListId = &listID
	}

	if utils.IsKnown(model.NamespaceType) {
		namespaceType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(model.NamespaceType.ValueString())
		body.NamespaceType = &namespaceType
	}

	if utils.IsKnown(model.Tags) {
		var tags []string
		diags.Append(model.Tags.ElementsAs(ctx, &tags, false)...)
		if !diags.HasError() {
			tagsArray := kbapi.SecurityExceptionsAPIExceptionListTags(tags)
			body.Tags = &tagsArray
		}
	}

	if !model.OsTypes.IsNull() && !model.OsTypes.IsUnknown() {
		var osTypes []string
		diags.Append(model.OsTypes.ElementsAs(ctx, &osTypes, false)...)
		if !diags.HasError() {
			var osTypesArray []kbapi.SecurityExceptionsAPIExceptionListOsType
			for _, osType := range osTypes {
				osTypesArray = append(osTypesArray, kbapi.SecurityExceptionsAPIExceptionListOsType(osType))
			}
			osTypesArrayTyped := kbapi.SecurityExceptionsAPIExceptionListOsTypeArray(osTypesArray)
			body.OsTypes = &osTypesArrayTyped
		}
	}

	return body, diags
}

func (model ExceptionContainerData) toAPIUpdateRequest(ctx context.Context, version string) (kbapi.UpdateExceptionListJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	body := kbapi.UpdateExceptionListJSONRequestBody{
		Name:        kbapi.SecurityExceptionsAPIExceptionListName(model.Name.ValueString()),
		Description: kbapi.SecurityExceptionsAPIExceptionListDescription(model.Description.ValueString()),
		Type:        kbapi.SecurityExceptionsAPIExceptionListType(model.Type.ValueString()),
	}

	if version != "" {
		body.UnderscoreVersion = &version
	}

	if utils.IsKnown(model.ListID) {
		listID := kbapi.SecurityExceptionsAPIExceptionListHumanId(model.ListID.ValueString())
		body.ListId = &listID
	}

	if utils.IsKnown(model.NamespaceType) {
		namespaceType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(model.NamespaceType.ValueString())
		body.NamespaceType = &namespaceType
	}

	if utils.IsKnown(model.Tags) {
		var tags []string
		diags.Append(model.Tags.ElementsAs(ctx, &tags, false)...)
		if !diags.HasError() {
			tagsArray := kbapi.SecurityExceptionsAPIExceptionListTags(tags)
			body.Tags = &tagsArray
		}
	}

	if utils.IsKnown(model.OsTypes) {
		var osTypes []string
		diags.Append(model.OsTypes.ElementsAs(ctx, &osTypes, false)...)
		if !diags.HasError() {
			var osTypesArray []kbapi.SecurityExceptionsAPIExceptionListOsType
			for _, osType := range osTypes {
				osTypesArray = append(osTypesArray, kbapi.SecurityExceptionsAPIExceptionListOsType(osType))
			}
			osTypesArrayTyped := kbapi.SecurityExceptionsAPIExceptionListOsTypeArray(osTypesArray)
			body.OsTypes = &osTypesArrayTyped
		}
	}

	return body, diags
}

func (model *ExceptionContainerData) populateFromAPI(ctx context.Context, apiModel *kbapi.SecurityExceptionsAPIExceptionList, compositeID *clients.CompositeId) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue(compositeID.String())
	model.SpaceID = types.StringValue(compositeID.ClusterId)
	model.ListID = types.StringValue(string(apiModel.ListId))
	model.Name = types.StringValue(string(apiModel.Name))
	model.Description = types.StringValue(string(apiModel.Description))
	model.Type = types.StringValue(string(apiModel.Type))
	model.NamespaceType = types.StringValue(string(apiModel.NamespaceType))

	if apiModel.Tags != nil {
		tags := []string(*apiModel.Tags)
		if len(tags) == 0 {
			model.Tags = types.ListNull(types.StringType)
		} else {
			tagsList, d := types.ListValueFrom(ctx, types.StringType, tags)
			diags.Append(d...)
			if !diags.HasError() {
				model.Tags = tagsList
			}
		}
	} else {
		model.Tags = types.ListNull(types.StringType)
	}

	if apiModel.OsTypes != nil {
		var osTypes []string
		for _, osType := range []kbapi.SecurityExceptionsAPIExceptionListOsType(*apiModel.OsTypes) {
			osTypes = append(osTypes, string(osType))
		}
		if len(osTypes) == 0 {
			model.OsTypes = types.ListNull(types.StringType)
		} else {
			osTypesList, d := types.ListValueFrom(ctx, types.StringType, osTypes)
			diags.Append(d...)
			if !diags.HasError() {
				model.OsTypes = osTypesList
			}
		}
	} else {
		model.OsTypes = types.ListNull(types.StringType)
	}

	return diags
}
