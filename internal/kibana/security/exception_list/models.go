package exception_list

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// exceptionListModel is the Terraform model for the exception list resource
type exceptionListModel struct {
	ID            types.String `tfsdk:"id"`
	SpaceID       types.String `tfsdk:"space_id"`
	ListID        types.String `tfsdk:"list_id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Type          types.String `tfsdk:"type"`
	NamespaceType types.String `tfsdk:"namespace_type"`
	OsTypes       types.List   `tfsdk:"os_types"`
	Tags          types.List   `tfsdk:"tags"`
	Meta          types.Object `tfsdk:"meta"`
	Version       types.Int64  `tfsdk:"version"`
}

// metaModel represents the meta object structure
type metaModel struct {
	AdditionalProperties types.Map `tfsdk:"additional_properties"`
}

func (m metaModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"additional_properties": types.MapType{ElemType: types.StringType},
	}
}

// toCreateRequest converts the Terraform model to an API create request
func (m *exceptionListModel) toCreateRequest(ctx context.Context) (kbapi.CreateExceptionListJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := kbapi.CreateExceptionListJSONRequestBody{
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
		Type:        kbapi.SecurityExceptionsAPIExceptionListType(m.Type.ValueString()),
	}

	// Set optional ListID
	if utils.IsKnown(m.ListID) {
		listID := kbapi.SecurityExceptionsAPIExceptionListHumanId(m.ListID.ValueString())
		req.ListId = &listID
	}

	// Set optional NamespaceType
	if utils.IsKnown(m.NamespaceType) {
		nsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(m.NamespaceType.ValueString())
		req.NamespaceType = &nsType
	}

	// Set optional OsTypes
	if utils.IsKnown(m.OsTypes) && !m.OsTypes.IsNull() {
		var osTypes []string
		diags.Append(m.OsTypes.ElementsAs(ctx, &osTypes, false)...)
		if !diags.HasError() {
			apiOsTypes := make(kbapi.SecurityExceptionsAPIExceptionListOsTypeArray, len(osTypes))
			for i, ot := range osTypes {
				apiOsTypes[i] = kbapi.SecurityExceptionsAPIExceptionListOsType(ot)
			}
			req.OsTypes = &apiOsTypes
		}
	}

	// Set optional Tags
	if utils.IsKnown(m.Tags) && !m.Tags.IsNull() {
		var tags []string
		diags.Append(m.Tags.ElementsAs(ctx, &tags, false)...)
		if !diags.HasError() {
			apiTags := kbapi.SecurityExceptionsAPIExceptionListTags(tags)
			req.Tags = &apiTags
		}
	}

	// Set optional Meta
	if utils.IsKnown(m.Meta) && !m.Meta.IsNull() {
		var meta metaModel
		diags.Append(m.Meta.As(ctx, &meta, utils.EmptyOpts())...)
		if !diags.HasError() && utils.IsKnown(meta.AdditionalProperties) {
			var props map[string]string
			diags.Append(meta.AdditionalProperties.ElementsAs(ctx, &props, false)...)
			if !diags.HasError() {
				apiMeta := kbapi.SecurityExceptionsAPIExceptionListMeta(props)
				req.Meta = &apiMeta
			}
		}
	}

	return req, diags
}

// toUpdateRequest converts the Terraform model to an API update request
func (m *exceptionListModel) toUpdateRequest(ctx context.Context) (kbapi.UpdateExceptionListJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := kbapi.UpdateExceptionListJSONRequestBody{
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
	}

	// Set ID from ListID
	if utils.IsKnown(m.ListID) {
		listID := kbapi.SecurityExceptionsAPIExceptionListHumanId(m.ListID.ValueString())
		req.ListId = &listID
	}

	// Set optional NamespaceType
	if utils.IsKnown(m.NamespaceType) {
		nsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(m.NamespaceType.ValueString())
		req.NamespaceType = &nsType
	}

	// Set optional OsTypes
	if utils.IsKnown(m.OsTypes) && !m.OsTypes.IsNull() {
		var osTypes []string
		diags.Append(m.OsTypes.ElementsAs(ctx, &osTypes, false)...)
		if !diags.HasError() {
			apiOsTypes := make(kbapi.SecurityExceptionsAPIExceptionListOsTypeArray, len(osTypes))
			for i, ot := range osTypes {
				apiOsTypes[i] = kbapi.SecurityExceptionsAPIExceptionListOsType(ot)
			}
			req.OsTypes = &apiOsTypes
		}
	}

	// Set optional Tags
	if utils.IsKnown(m.Tags) && !m.Tags.IsNull() {
		var tags []string
		diags.Append(m.Tags.ElementsAs(ctx, &tags, false)...)
		if !diags.HasError() {
			apiTags := kbapi.SecurityExceptionsAPIExceptionListTags(tags)
			req.Tags = &apiTags
		}
	}

	// Set optional Meta
	if utils.IsKnown(m.Meta) && !m.Meta.IsNull() {
		var meta metaModel
		diags.Append(m.Meta.As(ctx, &meta, utils.EmptyOpts())...)
		if !diags.HasError() && utils.IsKnown(meta.AdditionalProperties) {
			var props map[string]string
			diags.Append(meta.AdditionalProperties.ElementsAs(ctx, &props, false)...)
			if !diags.HasError() {
				apiMeta := kbapi.SecurityExceptionsAPIExceptionListMeta(props)
				req.Meta = &apiMeta
			}
		}
	}

	// Set type
	exType := kbapi.SecurityExceptionsAPIExceptionListType(m.Type.ValueString())
	req.Type = &exType

	return req, diags
}

// fromAPIResponse populates the Terraform model from an API response
func (m *exceptionListModel) fromAPIResponse(ctx context.Context, resp *kbapi.SecurityExceptionsAPIExceptionList, spaceID string) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.StringValue(resp.Id)
	m.SpaceID = types.StringValue(spaceID)
	m.ListID = types.StringValue(resp.ListId)
	m.Name = types.StringValue(resp.Name)
	m.Description = types.StringValue(resp.Description)
	m.Type = types.StringValue(string(resp.Type))
	m.NamespaceType = types.StringValue(string(resp.NamespaceType))

	// Convert version to int64 if available
	if resp.Version != nil {
		m.Version = types.Int64Value(int64(*resp.Version))
	}

	// Convert OsTypes
	if resp.OsTypes != nil {
		osTypesList := make([]attr.Value, len(*resp.OsTypes))
		for i, ot := range *resp.OsTypes {
			osTypesList[i] = types.StringValue(string(ot))
		}
		osList, d := types.ListValue(types.StringType, osTypesList)
		diags.Append(d...)
		m.OsTypes = osList
	} else {
		m.OsTypes = types.ListNull(types.StringType)
	}

	// Convert Tags
	if resp.Tags != nil {
		tagsList := make([]attr.Value, len(*resp.Tags))
		for i, tag := range *resp.Tags {
			tagsList[i] = types.StringValue(tag)
		}
		tagList, d := types.ListValue(types.StringType, tagsList)
		diags.Append(d...)
		m.Tags = tagList
	} else {
		m.Tags = types.ListNull(types.StringType)
	}

	// Convert Meta
	if resp.Meta != nil {
		metaMap := make(map[string]attr.Value)
		for k, v := range *resp.Meta {
			metaMap[k] = types.StringValue(v)
		}
		propsMap, d := types.MapValue(types.StringType, metaMap)
		diags.Append(d...)

		metaObj, d := types.ObjectValue(
			metaModel{}.attrTypes(),
			map[string]attr.Value{
				"additional_properties": propsMap,
			},
		)
		diags.Append(d...)
		m.Meta = metaObj
	} else {
		m.Meta = types.ObjectNull(metaModel{}.attrTypes())
	}

	return diags
}
