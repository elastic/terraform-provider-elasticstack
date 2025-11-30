package security_exception_list

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ExceptionListModel struct {
	ID            types.String         `tfsdk:"id"`
	SpaceID       types.String         `tfsdk:"space_id"`
	ListID        types.String         `tfsdk:"list_id"`
	Name          types.String         `tfsdk:"name"`
	Description   types.String         `tfsdk:"description"`
	Type          types.String         `tfsdk:"type"`
	NamespaceType types.String         `tfsdk:"namespace_type"`
	OsTypes       types.Set            `tfsdk:"os_types"`
	Tags          types.Set            `tfsdk:"tags"`
	Meta          jsontypes.Normalized `tfsdk:"meta"`
	CreatedAt     types.String         `tfsdk:"created_at"`
	CreatedBy     types.String         `tfsdk:"created_by"`
	UpdatedAt     types.String         `tfsdk:"updated_at"`
	UpdatedBy     types.String         `tfsdk:"updated_by"`
	Immutable     types.Bool           `tfsdk:"immutable"`
	TieBreakerID  types.String         `tfsdk:"tie_breaker_id"`
}

// toCreateRequest converts the Terraform model to API create request
func (m *ExceptionListModel) toCreateRequest(ctx context.Context) (*kbapi.CreateExceptionListJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := &kbapi.CreateExceptionListJSONRequestBody{
		Name:        kbapi.SecurityExceptionsAPIExceptionListName(m.Name.ValueString()),
		Description: kbapi.SecurityExceptionsAPIExceptionListDescription(m.Description.ValueString()),
		Type:        kbapi.SecurityExceptionsAPIExceptionListType(m.Type.ValueString()),
	}

	// Set optional list_id
	if utils.IsKnown(m.ListID) {
		listId := kbapi.SecurityExceptionsAPIExceptionListHumanId(m.ListID.ValueString())
		req.ListId = &listId
	}

	// Set optional namespace_type
	if utils.IsKnown(m.NamespaceType) {
		nsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(m.NamespaceType.ValueString())
		req.NamespaceType = &nsType
	}

	// Set optional os_types
	if utils.IsKnown(m.OsTypes) {
		var osTypes []string
		unmarshalDiags := m.OsTypes.ElementsAs(ctx, &osTypes, false)
		diags.Append(unmarshalDiags...)
		if diags.HasError() {
			return nil, diags
		}
		if len(osTypes) > 0 {
			osTypesArray := make(kbapi.SecurityExceptionsAPIExceptionListOsTypeArray, len(osTypes))
			for i, osType := range osTypes {
				osTypesArray[i] = kbapi.SecurityExceptionsAPIExceptionListOsType(osType)
			}
			req.OsTypes = &osTypesArray
		}
	}

	// Set optional tags
	if utils.IsKnown(m.Tags) {
		var tags []string
		unmarshalDiags := m.Tags.ElementsAs(ctx, &tags, false)
		diags.Append(unmarshalDiags...)
		if diags.HasError() {
			return nil, diags
		}
		if len(tags) > 0 {
			tagsArray := kbapi.SecurityExceptionsAPIExceptionListTags(tags)
			req.Tags = &tagsArray
		}
	}

	// Set optional meta
	if utils.IsKnown(m.Meta) {
		var meta kbapi.SecurityExceptionsAPIExceptionListMeta
		unmarshalDiags := m.Meta.Unmarshal(&meta)
		diags.Append(unmarshalDiags...)
		if diags.HasError() {
			return nil, diags
		}
		req.Meta = &meta
	}

	return req, diags
}

// toUpdateRequest converts the Terraform model to API update request
func (m *ExceptionListModel) toUpdateRequest(ctx context.Context, resourceId string) (*kbapi.UpdateExceptionListJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics
	id := kbapi.SecurityExceptionsAPIExceptionListId(resourceId)
	req := &kbapi.UpdateExceptionListJSONRequestBody{
		Id:          &id,
		Name:        kbapi.SecurityExceptionsAPIExceptionListName(m.Name.ValueString()),
		Description: kbapi.SecurityExceptionsAPIExceptionListDescription(m.Description.ValueString()),
		// Type is required by the API even though it has RequiresReplace in the schema
		// The API will reject updates without this field, even though the value cannot change
		Type: kbapi.SecurityExceptionsAPIExceptionListType(m.Type.ValueString()),
	}

	// Set optional namespace_type
	if utils.IsKnown(m.NamespaceType) {
		nsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(m.NamespaceType.ValueString())
		req.NamespaceType = &nsType
	}

	// Set optional os_types
	if utils.IsKnown(m.OsTypes) {
		var osTypes []string
		unmarshalDiags := m.OsTypes.ElementsAs(ctx, &osTypes, false)
		diags.Append(unmarshalDiags...)
		if diags.HasError() {
			return nil, diags
		}
		if len(osTypes) > 0 {
			osTypesArray := make(kbapi.SecurityExceptionsAPIExceptionListOsTypeArray, len(osTypes))
			for i, osType := range osTypes {
				osTypesArray[i] = kbapi.SecurityExceptionsAPIExceptionListOsType(osType)
			}
			req.OsTypes = &osTypesArray
		}
	}

	// Set optional tags
	if utils.IsKnown(m.Tags) {
		var tags []string
		unmarshalDiags := m.Tags.ElementsAs(ctx, &tags, false)
		diags.Append(unmarshalDiags...)
		if diags.HasError() {
			return nil, diags
		}
		if len(tags) > 0 {
			tagsArray := kbapi.SecurityExceptionsAPIExceptionListTags(tags)
			req.Tags = &tagsArray
		}
	}

	// Set optional meta
	if utils.IsKnown(m.Meta) {
		var meta kbapi.SecurityExceptionsAPIExceptionListMeta
		unmarshalDiags := m.Meta.Unmarshal(&meta)
		diags.Append(unmarshalDiags...)
		if diags.HasError() {
			return nil, diags
		}
		req.Meta = &meta
	}

	return req, diags
}

// fromAPI converts the API response to Terraform model
func (m *ExceptionListModel) fromAPI(ctx context.Context, apiList *kbapi.SecurityExceptionsAPIExceptionList) diag.Diagnostics {
	var diags diag.Diagnostics

	// Create composite ID from space_id and list id
	compId := clients.CompositeId{
		ClusterId:  m.SpaceID.ValueString(),
		ResourceId: string(apiList.Id),
	}
	m.ID = types.StringValue(compId.String())

	m.ListID = utils.StringishValue(apiList.ListId)
	m.Name = utils.StringishValue(apiList.Name)
	m.Description = utils.StringishValue(apiList.Description)
	m.Type = utils.StringishValue(apiList.Type)
	m.NamespaceType = utils.StringishValue(apiList.NamespaceType)
	m.Immutable = types.BoolValue(apiList.Immutable)
	m.TieBreakerID = types.StringValue(apiList.TieBreakerId)
	m.CreatedAt = types.StringValue(apiList.CreatedAt.Format("2006-01-02T15:04:05.000Z"))
	m.CreatedBy = types.StringValue(apiList.CreatedBy)
	m.UpdatedAt = types.StringValue(apiList.UpdatedAt.Format("2006-01-02T15:04:05.000Z"))
	m.UpdatedBy = types.StringValue(apiList.UpdatedBy)

	// Set optional os_types
	if apiList.OsTypes != nil && len(*apiList.OsTypes) > 0 {
		set, d := types.SetValueFrom(ctx, types.StringType, apiList.OsTypes)
		diags.Append(d...)
		m.OsTypes = set
	} else {
		m.OsTypes = types.SetNull(types.StringType)
	}

	// Set optional tags
	if apiList.Tags != nil && len(*apiList.Tags) > 0 {
		set, d := types.SetValueFrom(ctx, types.StringType, *apiList.Tags)
		diags.Append(d...)
		m.Tags = set
	} else {
		m.Tags = types.SetNull(types.StringType)
	}

	// Set optional meta
	if apiList.Meta != nil {
		metaBytes, err := json.Marshal(apiList.Meta)
		if err != nil {
			diags.AddError("Failed to marshal meta field from API response to JSON", err.Error())
			return diags
		}
		m.Meta = jsontypes.NewNormalizedValue(string(metaBytes))
	} else {
		m.Meta = jsontypes.NewNormalizedNull()
	}

	return diags
}
