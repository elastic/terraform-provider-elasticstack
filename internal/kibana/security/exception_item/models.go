package exception_item

import (
	"context"
	"encoding/json"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// exceptionItemModel is the Terraform model for the exception item resource
type exceptionItemModel struct {
	ID            types.String         `tfsdk:"id"`
	SpaceID       types.String         `tfsdk:"space_id"`
	ItemID        types.String         `tfsdk:"item_id"`
	ListID        types.String         `tfsdk:"list_id"`
	Name          types.String         `tfsdk:"name"`
	Description   types.String         `tfsdk:"description"`
	Type          types.String         `tfsdk:"type"`
	NamespaceType types.String         `tfsdk:"namespace_type"`
	Entries       jsontypes.Normalized `tfsdk:"entries"`
	Comments      types.List           `tfsdk:"comments"`
	OsTypes       types.List           `tfsdk:"os_types"`
	Tags          types.List           `tfsdk:"tags"`
	Meta          types.Object         `tfsdk:"meta"`
	ExpireTime    types.String         `tfsdk:"expire_time"`
}

// commentModel represents a comment structure
type commentModel struct {
	Comment types.String `tfsdk:"comment"`
}

func (c commentModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"comment": types.StringType,
	}
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
func (m *exceptionItemModel) toCreateRequest(ctx context.Context) (kbapi.CreateExceptionListItemJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := kbapi.CreateExceptionListItemJSONRequestBody{
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
		Type:        kbapi.SecurityExceptionsAPIExceptionListItemType(m.Type.ValueString()),
		ListId:      kbapi.SecurityExceptionsAPIExceptionListHumanId(m.ListID.ValueString()),
	}

	// Parse entries JSON
	if utils.IsKnown(m.Entries) && !m.Entries.IsNull() {
		var apiEntries kbapi.SecurityExceptionsAPIExceptionListItemEntryArray
		if err := json.Unmarshal([]byte(m.Entries.ValueString()), &apiEntries); err != nil {
			diags.AddError("Failed to parse entries", err.Error())
			return req, diags
		}
		req.Entries = apiEntries
	}

	// Set optional ItemID
	if utils.IsKnown(m.ItemID) {
		itemID := kbapi.SecurityExceptionsAPIExceptionListItemHumanId(m.ItemID.ValueString())
		req.ItemId = &itemID
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
			apiTags := kbapi.SecurityExceptionsAPIExceptionListItemTags(tags)
			req.Tags = &apiTags
		}
	}

	// Set optional Meta
	if utils.IsKnown(m.Meta) && !m.Meta.IsNull() {
		var meta metaModel
		diags.Append(m.Meta.As(ctx, &meta, basetypes.ObjectAsOptions{})...)
		if !diags.HasError() && utils.IsKnown(meta.AdditionalProperties) {
			var props map[string]string
			diags.Append(meta.AdditionalProperties.ElementsAs(ctx, &props, false)...)
			if !diags.HasError() {
				apiMeta := make(kbapi.SecurityExceptionsAPIExceptionListItemMeta)
				for k, v := range props {
					apiMeta[k] = v
				}
				req.Meta = &apiMeta
			}
		}
	}

	// Set optional ExpireTime
	if utils.IsKnown(m.ExpireTime) {
		expTime, err := time.Parse(time.RFC3339, m.ExpireTime.ValueString())
		if err != nil {
			diags.AddError("Failed to parse expire_time", err.Error())
		} else {
			expireTime := (kbapi.SecurityExceptionsAPIExceptionListItemExpireTime)(expTime)
					req.ExpireTime = &expireTime
		}
	}

	// Set optional Comments
	if utils.IsKnown(m.Comments) && !m.Comments.IsNull() {
		var comments []commentModel
		diags.Append(m.Comments.ElementsAs(ctx, &comments, false)...)
		if !diags.HasError() {
			apiComments := make(kbapi.SecurityExceptionsAPICreateExceptionListItemCommentArray, len(comments))
			for i, c := range comments {
				apiComments[i] = kbapi.SecurityExceptionsAPICreateExceptionListItemComment{
					Comment: kbapi.SecurityExceptionsAPINonEmptyString(c.Comment.ValueString()),
				}
			}
			req.Comments = &apiComments
		}
	}

	return req, diags
}

// toUpdateRequest converts the Terraform model to an API update request
func (m *exceptionItemModel) toUpdateRequest(ctx context.Context) (kbapi.UpdateExceptionListItemJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := kbapi.UpdateExceptionListItemJSONRequestBody{
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
		Type:        kbapi.SecurityExceptionsAPIExceptionListItemType(m.Type.ValueString()),
	}

	// Parse entries JSON
	if utils.IsKnown(m.Entries) && !m.Entries.IsNull() {
		var apiEntries kbapi.SecurityExceptionsAPIExceptionListItemEntryArray
		if err := json.Unmarshal([]byte(m.Entries.ValueString()), &apiEntries); err != nil {
			diags.AddError("Failed to parse entries", err.Error())
			return req, diags
		}
		req.Entries = apiEntries
	}

	// Set ItemID
	if utils.IsKnown(m.ItemID) {
		itemID := kbapi.SecurityExceptionsAPIExceptionListItemHumanId(m.ItemID.ValueString())
		req.ItemId = &itemID
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
			apiTags := kbapi.SecurityExceptionsAPIExceptionListItemTags(tags)
			req.Tags = &apiTags
		}
	}

	// Set optional Meta
	if utils.IsKnown(m.Meta) && !m.Meta.IsNull() {
		var meta metaModel
		diags.Append(m.Meta.As(ctx, &meta, basetypes.ObjectAsOptions{})...)
		if !diags.HasError() && utils.IsKnown(meta.AdditionalProperties) {
			var props map[string]string
			diags.Append(meta.AdditionalProperties.ElementsAs(ctx, &props, false)...)
			if !diags.HasError() {
				apiMeta := make(kbapi.SecurityExceptionsAPIExceptionListItemMeta)
				for k, v := range props {
					apiMeta[k] = v
				}
				req.Meta = &apiMeta
			}
		}
	}

	// Set optional ExpireTime
	if utils.IsKnown(m.ExpireTime) {
		expTime, err := time.Parse(time.RFC3339, m.ExpireTime.ValueString())
		if err != nil {
			diags.AddError("Failed to parse expire_time", err.Error())
		} else {
			expireTime := (kbapi.SecurityExceptionsAPIExceptionListItemExpireTime)(expTime)
					req.ExpireTime = &expireTime
		}
	}

	// Set optional Comments - for updates, we use UpdateExceptionListItemComment
	if utils.IsKnown(m.Comments) && !m.Comments.IsNull() {
		var comments []commentModel
		diags.Append(m.Comments.ElementsAs(ctx, &comments, false)...)
		if !diags.HasError() {
			apiComments := make(kbapi.SecurityExceptionsAPIUpdateExceptionListItemCommentArray, len(comments))
			for i, c := range comments {
				apiComments[i] = kbapi.SecurityExceptionsAPIUpdateExceptionListItemComment{
					Comment: kbapi.SecurityExceptionsAPINonEmptyString(c.Comment.ValueString()),
				}
			}
			req.Comments = &apiComments
		}
	}

	return req, diags
}

// fromAPIResponse populates the Terraform model from an API response
func (m *exceptionItemModel) fromAPIResponse(ctx context.Context, resp *kbapi.SecurityExceptionsAPIExceptionListItem, spaceID string) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.StringValue(resp.Id)
	m.SpaceID = types.StringValue(spaceID)
	m.ItemID = types.StringValue(resp.ItemId)
	m.ListID = types.StringValue(resp.ListId)
	m.Name = types.StringValue(resp.Name)
	m.Description = types.StringValue(resp.Description)
	m.Type = types.StringValue(string(resp.Type))
	m.NamespaceType = types.StringValue(string(resp.NamespaceType))

	// Convert entries to JSON string
	if len(resp.Entries) > 0 {
		entriesJSON, err := json.Marshal(resp.Entries)
		if err != nil {
			diags.AddError("Failed to marshal entries", err.Error())
		} else {
			m.Entries = jsontypes.NewNormalizedValue(string(entriesJSON))
		}
	} else {
		m.Entries = jsontypes.NewNormalizedNull()
	}

	// Convert Comments
	if len(resp.Comments) > 0 {
		commentsList := make([]attr.Value, len(resp.Comments))
		for i, c := range resp.Comments {
			commentObj, d := types.ObjectValue(
				commentModel{}.attrTypes(),
				map[string]attr.Value{
					"comment": types.StringValue(c.Comment),
				},
			)
			diags.Append(d...)
			commentsList[i] = commentObj
		}
		commList, d := types.ListValue(types.ObjectType{AttrTypes: commentModel{}.attrTypes()}, commentsList)
		diags.Append(d...)
		m.Comments = commList
	} else {
		m.Comments = types.ListNull(types.ObjectType{AttrTypes: commentModel{}.attrTypes()})
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
			if strVal, ok := v.(string); ok {
				metaMap[k] = types.StringValue(strVal)
			}
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

	// Convert ExpireTime
	if resp.ExpireTime != nil {
		m.ExpireTime = types.StringValue((*resp.ExpireTime).Format(time.RFC3339))
	} else {
		m.ExpireTime = types.StringNull()
	}

	return diags
}
