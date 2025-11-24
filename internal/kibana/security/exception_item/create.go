package exception_item

import (
	"context"
	"encoding/json"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *ExceptionItemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ExceptionItemModel

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

	// Convert entries from Terraform model to API model
	entries, diags := convertEntriesToAPI(ctx, plan.Entries)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the request body
	body := kbapi.CreateExceptionListItemJSONRequestBody{
		ListId:      kbapi.SecurityExceptionsAPIExceptionListHumanId(plan.ListID.ValueString()),
		Name:        kbapi.SecurityExceptionsAPIExceptionListItemName(plan.Name.ValueString()),
		Description: kbapi.SecurityExceptionsAPIExceptionListItemDescription(plan.Description.ValueString()),
		Type:        kbapi.SecurityExceptionsAPIExceptionListItemType(plan.Type.ValueString()),
		Entries:     entries,
	}

	// Set optional item_id
	if utils.IsKnown(plan.ItemID) && !plan.ItemID.IsNull() {
		itemID := kbapi.SecurityExceptionsAPIExceptionListItemHumanId(plan.ItemID.ValueString())
		body.ItemId = &itemID
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
			osTypesArray := make(kbapi.SecurityExceptionsAPIExceptionListItemOsTypeArray, len(osTypes))
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
			tagsArray := kbapi.SecurityExceptionsAPIExceptionListItemTags(tags)
			body.Tags = &tagsArray
		}
	}

	// Set optional meta
	if utils.IsKnown(plan.Meta) && !plan.Meta.IsNull() {
		var meta kbapi.SecurityExceptionsAPIExceptionListItemMeta
		if err := json.Unmarshal([]byte(plan.Meta.ValueString()), &meta); err != nil {
			resp.Diagnostics.AddError("Failed to parse meta JSON", err.Error())
			return
		}
		body.Meta = &meta
	}

	// Set optional comments
	if utils.IsKnown(plan.Comments) && !plan.Comments.IsNull() {
		var comments []CommentModel
		diags := plan.Comments.ElementsAs(ctx, &comments, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(comments) > 0 {
			commentsArray := make(kbapi.SecurityExceptionsAPICreateExceptionListItemCommentArray, len(comments))
			for i, comment := range comments {
				commentsArray[i] = kbapi.SecurityExceptionsAPICreateExceptionListItemComment{
					Comment: kbapi.SecurityExceptionsAPINonEmptyString(comment.Comment.ValueString()),
				}
			}
			body.Comments = &commentsArray
		}
	}

	// Set optional expire_time
	if utils.IsKnown(plan.ExpireTime) && !plan.ExpireTime.IsNull() {
		expireTime, err := time.Parse(time.RFC3339, plan.ExpireTime.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to parse expire_time", err.Error())
			return
		}
		expireTimeAPI := kbapi.SecurityExceptionsAPIExceptionListItemExpireTime(expireTime)
		body.ExpireTime = &expireTimeAPI
	}

	// Create the exception item
	createResp, diags := kibana_oapi.CreateExceptionListItem(ctx, client, plan.SpaceID.ValueString(), body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if createResp == nil || createResp.JSON200 == nil {
		resp.Diagnostics.AddError("Failed to create exception item", "API returned empty response")
		return
	}

	/*
	 * In create/update paths we typically follow the write operation with a read, and then set the state from the read.
	 * We want to avoid a dirty plan immediately after an apply.
	 */
	// Read back the created resource to get the final state
	readParams := &kbapi.ReadExceptionListItemParams{
		Id: (*kbapi.SecurityExceptionsAPIExceptionListItemId)(&createResp.JSON200.Id),
	}

	readResp, diags := kibana_oapi.GetExceptionListItem(ctx, client, plan.SpaceID.ValueString(), readParams)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if readResp == nil || readResp.JSON200 == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state with read response
	diags = r.updateStateFromAPIResponse(ctx, &plan, readResp.JSON200)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ExceptionItemResource) updateStateFromAPIResponse(ctx context.Context, model *ExceptionItemModel, apiResp *kbapi.SecurityExceptionsAPIExceptionListItem) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue(string(apiResp.Id))
	model.ItemID = types.StringValue(string(apiResp.ItemId))
	model.ListID = types.StringValue(string(apiResp.ListId))
	model.Name = types.StringValue(string(apiResp.Name))
	model.Description = types.StringValue(string(apiResp.Description))
	model.Type = types.StringValue(string(apiResp.Type))
	model.NamespaceType = types.StringValue(string(apiResp.NamespaceType))
	model.CreatedAt = types.StringValue(apiResp.CreatedAt.Format("2006-01-02T15:04:05.000Z"))
	model.CreatedBy = types.StringValue(apiResp.CreatedBy)
	model.UpdatedAt = types.StringValue(apiResp.UpdatedAt.Format("2006-01-02T15:04:05.000Z"))
	model.UpdatedBy = types.StringValue(apiResp.UpdatedBy)
	model.TieBreakerID = types.StringValue(apiResp.TieBreakerId)

	// Set optional expire_time
	if apiResp.ExpireTime != nil {
		model.ExpireTime = types.StringValue(time.Time(*apiResp.ExpireTime).Format(time.RFC3339))
	} else {
		model.ExpireTime = types.StringNull()
	}

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

	// Set entries (convert from API model to Terraform model)
	entriesList, d := convertEntriesFromAPI(ctx, apiResp.Entries)
	diags.Append(d...)
	model.Entries = entriesList

	// Set optional comments
	if len(apiResp.Comments) > 0 {
		comments := make([]CommentModel, len(apiResp.Comments))
		for i, comment := range apiResp.Comments {
			comments[i] = CommentModel{
				ID:      types.StringValue(string(comment.Id)),
				Comment: types.StringValue(string(comment.Comment)),
			}
		}
		list, d := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"id":      types.StringType,
				"comment": types.StringType,
			},
		}, comments)
		diags.Append(d...)
		model.Comments = list
	} else {
		model.Comments = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"id":      types.StringType,
				"comment": types.StringType,
			},
		})
	}

	return diags
}
