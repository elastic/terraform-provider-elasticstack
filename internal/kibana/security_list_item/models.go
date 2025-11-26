package security_list_item

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SecurityListItemModel struct {
	ID        types.String `tfsdk:"id"`
	SpaceID   types.String `tfsdk:"space_id"`
	ListID    types.String `tfsdk:"list_id"`
	Value     types.String `tfsdk:"value"`
	Meta      types.String `tfsdk:"meta"`
	CreatedAt types.String `tfsdk:"created_at"`
	CreatedBy types.String `tfsdk:"created_by"`
	UpdatedAt types.String `tfsdk:"updated_at"`
	UpdatedBy types.String `tfsdk:"updated_by"`
	Version   types.String `tfsdk:"version"`
}

// toAPICreateModel converts the Terraform model to the API create request body
func (m *SecurityListItemModel) toAPICreateModel(ctx context.Context) (*kbapi.CreateListItemJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	body := &kbapi.CreateListItemJSONRequestBody{
		ListId: kbapi.SecurityListsAPIListId(m.ListID.ValueString()),
		Value:  kbapi.SecurityListsAPIListItemValue(m.Value.ValueString()),
	}

	// Set optional ID if specified
	if !m.ID.IsNull() && !m.ID.IsUnknown() {
		id := kbapi.SecurityListsAPIListItemId(m.ID.ValueString())
		body.Id = &id
	}

	// Set optional meta if specified
	if !m.Meta.IsNull() && !m.Meta.IsUnknown() {
		var meta kbapi.SecurityListsAPIListItemMetadata
		if err := json.Unmarshal([]byte(m.Meta.ValueString()), &meta); err != nil {
			diags.AddError("Failed to parse meta JSON", err.Error())
			return nil, diags
		}
		body.Meta = &meta
	}

	return body, diags
}

// toAPIUpdateModel converts the Terraform model to the API update request body
func (m *SecurityListItemModel) toAPIUpdateModel(ctx context.Context) (*kbapi.UpdateListItemJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	body := &kbapi.UpdateListItemJSONRequestBody{
		Id:    kbapi.SecurityListsAPIListItemId(m.ID.ValueString()),
		Value: kbapi.SecurityListsAPIListItemValue(m.Value.ValueString()),
	}

	// Set optional version if available
	if !m.Version.IsNull() && !m.Version.IsUnknown() {
		version := kbapi.SecurityListsAPIListVersionId(m.Version.ValueString())
		body.UnderscoreVersion = &version
	}

	// Set optional meta if specified
	if !m.Meta.IsNull() && !m.Meta.IsUnknown() {
		var meta kbapi.SecurityListsAPIListItemMetadata
		if err := json.Unmarshal([]byte(m.Meta.ValueString()), &meta); err != nil {
			diags.AddError("Failed to parse meta JSON", err.Error())
			return nil, diags
		}
		body.Meta = &meta
	}

	return body, diags
}

// fromAPIModel populates the Terraform model from an API response
func (m *SecurityListItemModel) fromAPIModel(ctx context.Context, apiItem *kbapi.SecurityListsAPIListItem) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.StringValue(string(apiItem.Id))
	m.ListID = types.StringValue(string(apiItem.ListId))
	m.Value = types.StringValue(string(apiItem.Value))
	m.CreatedAt = types.StringValue(apiItem.CreatedAt.Format("2006-01-02T15:04:05.000Z"))
	m.CreatedBy = types.StringValue(apiItem.CreatedBy)
	m.UpdatedAt = types.StringValue(apiItem.UpdatedAt.Format("2006-01-02T15:04:05.000Z"))
	m.UpdatedBy = types.StringValue(apiItem.UpdatedBy)

	// Set version if available
	if apiItem.UnderscoreVersion != nil {
		m.Version = types.StringValue(string(*apiItem.UnderscoreVersion))
	} else {
		m.Version = types.StringNull()
	}

	// Set meta if available
	if apiItem.Meta != nil {
		metaJSON, err := json.Marshal(apiItem.Meta)
		if err != nil {
			diags.AddError("Failed to serialize meta", err.Error())
			return diags
		}
		m.Meta = types.StringValue(string(metaJSON))
	} else {
		m.Meta = types.StringNull()
	}

	return diags
}
