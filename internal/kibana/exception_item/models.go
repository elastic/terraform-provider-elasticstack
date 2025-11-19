package exception_item

import (
	"context"
	"encoding/json"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ExceptionItemData struct {
	ID               types.String         `tfsdk:"id"`
	KibanaConnection types.List           `tfsdk:"kibana_connection"`
	SpaceID          types.String         `tfsdk:"space_id"`
	ListID           types.String         `tfsdk:"list_id"`
	ItemID           types.String         `tfsdk:"item_id"`
	Name             types.String         `tfsdk:"name"`
	Description      types.String         `tfsdk:"description"`
	Type             types.String         `tfsdk:"type"`
	NamespaceType    types.String         `tfsdk:"namespace_type"`
	Entries          jsontypes.Normalized `tfsdk:"entries"`
	Comments         jsontypes.Normalized `tfsdk:"comments"`
	ExpireTime       types.String         `tfsdk:"expire_time"`
	Tags             types.List           `tfsdk:"tags"`
	OsTypes          types.List           `tfsdk:"os_types"`
}

func (model ExceptionItemData) GetID() (*clients.CompositeId, diag.Diagnostics) {
	compId, sdkDiags := clients.CompositeIdFromStr(model.ID.ValueString())
	if sdkDiags.HasError() {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic("Invalid ID", sdkDiags[0].Summary),
		}
	}
	return compId, nil
}

func (model ExceptionItemData) toAPICreateRequest(ctx context.Context) (kbapi.CreateExceptionListItemJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	body := kbapi.CreateExceptionListItemJSONRequestBody{
		Name:        kbapi.SecurityExceptionsAPIExceptionListItemName(model.Name.ValueString()),
		Description: kbapi.SecurityExceptionsAPIExceptionListItemDescription(model.Description.ValueString()),
		Type:        kbapi.SecurityExceptionsAPIExceptionListItemType(model.Type.ValueString()),
		ListId:      kbapi.SecurityExceptionsAPIExceptionListHumanId(model.ListID.ValueString()),
	}

	if utils.IsKnown(model.ItemID) {
		itemID := kbapi.SecurityExceptionsAPIExceptionListItemHumanId(model.ItemID.ValueString())
		body.ItemId = &itemID
	}

	if utils.IsKnown(model.NamespaceType) {
		namespaceType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(model.NamespaceType.ValueString())
		body.NamespaceType = &namespaceType
	}

	// Parse entries JSON
	if utils.IsKnown(model.Entries) {
		entriesJSON := model.Entries.ValueString()
		var entries []kbapi.SecurityExceptionsAPIExceptionListItemEntry
		if err := json.Unmarshal([]byte(entriesJSON), &entries); err != nil {
			diags.AddError("Invalid entries JSON", err.Error())
			return body, diags
		}
		body.Entries = entries
	}

	// Parse comments JSON if provided
	if utils.IsKnown(model.Comments) {
		commentsJSON := model.Comments.ValueString()
		var comments []kbapi.SecurityExceptionsAPICreateExceptionListItemComment
		if err := json.Unmarshal([]byte(commentsJSON), &comments); err != nil {
			diags.AddError("Invalid comments JSON", err.Error())
			return body, diags
		}
		body.Comments = &comments
	}

	if utils.IsKnown(model.ExpireTime) {
		expireTimeStr := model.ExpireTime.ValueString()
		expireTime, err := time.Parse(time.RFC3339, expireTimeStr)
		if err != nil {
			diags.AddError("Invalid expire_time format", "expire_time must be in RFC3339 format (e.g., 2006-01-02T15:04:05Z07:00)")
			return body, diags
		}
		expireTimeTyped := kbapi.SecurityExceptionsAPIExceptionListItemExpireTime(expireTime)
		body.ExpireTime = &expireTimeTyped
	}

	if utils.IsKnown(model.Tags) {
		var tags []string
		diags.Append(model.Tags.ElementsAs(ctx, &tags, false)...)
		if !diags.HasError() {
			tagsArray := kbapi.SecurityExceptionsAPIExceptionListItemTags(tags)
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
			osTypesArrayTyped := kbapi.SecurityExceptionsAPIExceptionListItemOsTypeArray(osTypesArray)
			body.OsTypes = &osTypesArrayTyped
		}
	}

	return body, diags
}

func (model ExceptionItemData) toAPIUpdateRequest(ctx context.Context, version string) (kbapi.UpdateExceptionListItemJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	body := kbapi.UpdateExceptionListItemJSONRequestBody{
		Name:        kbapi.SecurityExceptionsAPIExceptionListItemName(model.Name.ValueString()),
		Description: kbapi.SecurityExceptionsAPIExceptionListItemDescription(model.Description.ValueString()),
		Type:        kbapi.SecurityExceptionsAPIExceptionListItemType(model.Type.ValueString()),
	}

	if version != "" {
		body.UnderscoreVersion = &version
	}

	if utils.IsKnown(model.ItemID) {
		itemID := kbapi.SecurityExceptionsAPIExceptionListItemHumanId(model.ItemID.ValueString())
		body.ItemId = &itemID
	}

	if utils.IsKnown(model.ListID) {
		listID := kbapi.SecurityExceptionsAPIExceptionListHumanId(model.ListID.ValueString())
		body.ListId = &listID
	}

	if utils.IsKnown(model.NamespaceType) {
		namespaceType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(model.NamespaceType.ValueString())
		body.NamespaceType = &namespaceType
	}

	// Parse entries JSON
	if utils.IsKnown(model.Entries) {
		entriesJSON := model.Entries.ValueString()
		var entries []kbapi.SecurityExceptionsAPIExceptionListItemEntry
		if err := json.Unmarshal([]byte(entriesJSON), &entries); err != nil {
			diags.AddError("Invalid entries JSON", err.Error())
			return body, diags
		}
		body.Entries = entries
	}

	// Parse comments JSON if provided
	if utils.IsKnown(model.Comments) {
		commentsJSON := model.Comments.ValueString()
		var comments []kbapi.SecurityExceptionsAPIUpdateExceptionListItemComment
		if err := json.Unmarshal([]byte(commentsJSON), &comments); err != nil {
			diags.AddError("Invalid comments JSON", err.Error())
			return body, diags
		}
		body.Comments = &comments
	}

	if utils.IsKnown(model.ExpireTime) {
		expireTimeStr := model.ExpireTime.ValueString()
		expireTime, err := time.Parse(time.RFC3339, expireTimeStr)
		if err != nil {
			diags.AddError("Invalid expire_time format", "expire_time must be in RFC3339 format (e.g., 2006-01-02T15:04:05Z07:00)")
			return body, diags
		}
		expireTimeTyped := kbapi.SecurityExceptionsAPIExceptionListItemExpireTime(expireTime)
		body.ExpireTime = &expireTimeTyped
	}

	if utils.IsKnown(model.Tags) {
		var tags []string
		diags.Append(model.Tags.ElementsAs(ctx, &tags, false)...)
		if !diags.HasError() {
			tagsArray := kbapi.SecurityExceptionsAPIExceptionListItemTags(tags)
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
			osTypesArrayTyped := kbapi.SecurityExceptionsAPIExceptionListItemOsTypeArray(osTypesArray)
			body.OsTypes = &osTypesArrayTyped
		}
	}

	return body, diags
}

func (model *ExceptionItemData) populateFromAPI(ctx context.Context, apiModel *kbapi.SecurityExceptionsAPIExceptionListItem, compositeID *clients.CompositeId) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue(compositeID.String())
	model.SpaceID = types.StringValue(compositeID.ClusterId)
	model.ListID = types.StringValue(string(apiModel.ListId))
	model.ItemID = types.StringValue(string(apiModel.ItemId))
	model.Name = types.StringValue(string(apiModel.Name))
	model.Description = types.StringValue(string(apiModel.Description))
	model.Type = types.StringValue(string(apiModel.Type))
	model.NamespaceType = types.StringValue(string(apiModel.NamespaceType))

	// Serialize entries to JSON
	entriesJSON, err := json.Marshal(apiModel.Entries)
	if err != nil {
		diags.AddError("Failed to marshal entries", err.Error())
		return diags
	}
	model.Entries = jsontypes.NewNormalizedValue(string(entriesJSON))

	// Serialize comments to JSON if present
	if len(apiModel.Comments) > 0 {
		commentsJSON, err := json.Marshal(apiModel.Comments)
		if err != nil {
			diags.AddError("Failed to marshal comments", err.Error())
			return diags
		}
		model.Comments = jsontypes.NewNormalizedValue(string(commentsJSON))
	} else {
		model.Comments = jsontypes.NewNormalizedNull()
	}

	if apiModel.ExpireTime != nil {
		expireTime := time.Time(*apiModel.ExpireTime)
		model.ExpireTime = types.StringValue(expireTime.Format(time.RFC3339))
	} else {
		model.ExpireTime = types.StringNull()
	}

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
