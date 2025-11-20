package exception_item

import (
	"context"
	"encoding/json"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *ExceptionItemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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

	// Parse entries JSON
	var entries kbapi.SecurityExceptionsAPIExceptionListItemEntryArray
	if err := json.Unmarshal([]byte(plan.Entries.ValueString()), &entries); err != nil {
		resp.Diagnostics.AddError("Failed to parse entries JSON", err.Error())
		return
	}

	// Build the update request body
	id := kbapi.SecurityExceptionsAPIExceptionListItemId(plan.ID.ValueString())
	body := kbapi.UpdateExceptionListItemJSONRequestBody{
		Id:          &id,
		Name:        kbapi.SecurityExceptionsAPIExceptionListItemName(plan.Name.ValueString()),
		Description: kbapi.SecurityExceptionsAPIExceptionListItemDescription(plan.Description.ValueString()),
		Type:        kbapi.SecurityExceptionsAPIExceptionListItemType(plan.Type.ValueString()),
		Entries:     entries,
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
			commentsArray := make(kbapi.SecurityExceptionsAPIUpdateExceptionListItemCommentArray, len(comments))
			for i, comment := range comments {
				commentsArray[i] = kbapi.SecurityExceptionsAPIUpdateExceptionListItemComment{
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

	// Update the exception item
	updateResp, diags := kibana_oapi.UpdateExceptionListItem(ctx, client, plan.SpaceID.ValueString(), body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if updateResp == nil || updateResp.JSON200 == nil {
		resp.Diagnostics.AddError("Failed to update exception item", "API returned empty response")
		return
	}

	/*
	 * In create/update paths we typically follow the write operation with a read, and then set the state from the read.
	 * We want to avoid a dirty plan immediately after an apply.
	 */
	// Read back the updated resource to get the final state
	readParams := &kbapi.ReadExceptionListItemParams{
		Id: (*kbapi.SecurityExceptionsAPIExceptionListItemId)(&updateResp.JSON200.Id),
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
