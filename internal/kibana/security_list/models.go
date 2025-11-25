package security_list

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SecurityListModel struct {
	ID           types.String `tfsdk:"id"`
	SpaceID      types.String `tfsdk:"space_id"`
	ListID       types.String `tfsdk:"list_id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Type         types.String `tfsdk:"type"`
	Deserializer types.String `tfsdk:"deserializer"`
	Serializer   types.String `tfsdk:"serializer"`
	Meta         types.String `tfsdk:"meta"`
	Version      types.Int64  `tfsdk:"version"`
	VersionID    types.String `tfsdk:"version_id"`
	Immutable    types.Bool   `tfsdk:"immutable"`
	CreatedAt    types.String `tfsdk:"created_at"`
	CreatedBy    types.String `tfsdk:"created_by"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
	UpdatedBy    types.String `tfsdk:"updated_by"`
	TieBreakerID types.String `tfsdk:"tie_breaker_id"`
}

// toCreateRequest converts the Terraform model to API create request
func (m *SecurityListModel) toCreateRequest() (*kbapi.CreateListJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := &kbapi.CreateListJSONRequestBody{
		Name:        kbapi.SecurityListsAPIListName(m.Name.ValueString()),
		Description: kbapi.SecurityListsAPIListDescription(m.Description.ValueString()),
		Type:        kbapi.SecurityListsAPIListType(m.Type.ValueString()),
	}

	// Set optional fields
	if !m.ListID.IsNull() && !m.ListID.IsUnknown() {
		id := kbapi.SecurityListsAPIListId(m.ListID.ValueString())
		req.Id = &id
	}

	if !m.Deserializer.IsNull() && !m.Deserializer.IsUnknown() {
		deserializer := kbapi.SecurityListsAPIListDeserializer(m.Deserializer.ValueString())
		req.Deserializer = &deserializer
	}

	if !m.Serializer.IsNull() && !m.Serializer.IsUnknown() {
		serializer := kbapi.SecurityListsAPIListSerializer(m.Serializer.ValueString())
		req.Serializer = &serializer
	}

	if !m.Meta.IsNull() && !m.Meta.IsUnknown() {
		var metaMap kbapi.SecurityListsAPIListMetadata
		if err := json.Unmarshal([]byte(m.Meta.ValueString()), &metaMap); err != nil {
			diags.AddError("Invalid meta JSON", err.Error())
			return nil, diags
		}
		req.Meta = &metaMap
	}

	if !m.Version.IsNull() && !m.Version.IsUnknown() {
		version := int(m.Version.ValueInt64())
		req.Version = &version
	}

	return req, diags
}

// toUpdateRequest converts the Terraform model to API update request
// versionID should be passed from the current state for optimistic locking
func (m *SecurityListModel) toUpdateRequest(versionID string) (*kbapi.UpdateListJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := &kbapi.UpdateListJSONRequestBody{
		Id:          kbapi.SecurityListsAPIListId(m.ListID.ValueString()),
		Name:        kbapi.SecurityListsAPIListName(m.Name.ValueString()),
		Description: kbapi.SecurityListsAPIListDescription(m.Description.ValueString()),
	}

	// Set version ID from state for optimistic locking
	if versionID != "" {
		versionIDValue := kbapi.SecurityListsAPIListVersionId(versionID)
		req.UnderscoreVersion = &versionIDValue
	}

	if !m.Meta.IsNull() && !m.Meta.IsUnknown() {
		var metaMap kbapi.SecurityListsAPIListMetadata
		if err := json.Unmarshal([]byte(m.Meta.ValueString()), &metaMap); err != nil {
			diags.AddError("Invalid meta JSON", err.Error())
			return nil, diags
		}
		req.Meta = &metaMap
	}

	// Note: Version field is not sent in update requests as it's managed server-side

	return req, diags
}

// fromAPI converts the API response to Terraform model
func (m *SecurityListModel) fromAPI(ctx context.Context, apiList *kbapi.SecurityListsAPIList) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.StringValue(string(apiList.Id))
	m.ListID = types.StringValue(string(apiList.Id))
	m.Name = types.StringValue(string(apiList.Name))
	m.Description = types.StringValue(string(apiList.Description))
	m.Type = types.StringValue(string(apiList.Type))
	m.Immutable = types.BoolValue(apiList.Immutable)
	m.Version = types.Int64Value(int64(apiList.Version))
	m.TieBreakerID = types.StringValue(apiList.TieBreakerId)
	m.CreatedAt = types.StringValue(apiList.CreatedAt.String())
	m.CreatedBy = types.StringValue(apiList.CreatedBy)
	m.UpdatedAt = types.StringValue(apiList.UpdatedAt.String())
	m.UpdatedBy = types.StringValue(apiList.UpdatedBy)

	// Set optional _version field
	if apiList.UnderscoreVersion != nil {
		m.VersionID = types.StringValue(string(*apiList.UnderscoreVersion))
	} else {
		m.VersionID = types.StringNull()
	}

	if apiList.Deserializer != nil {
		m.Deserializer = types.StringValue(string(*apiList.Deserializer))
	} else {
		m.Deserializer = types.StringNull()
	}

	if apiList.Serializer != nil {
		m.Serializer = types.StringValue(string(*apiList.Serializer))
	} else {
		m.Serializer = types.StringNull()
	}

	if apiList.Meta != nil {
		metaBytes, err := json.Marshal(apiList.Meta)
		if err != nil {
			diags.AddError("Failed to marshal meta", err.Error())
			return diags
		}
		m.Meta = types.StringValue(string(metaBytes))
	} else {
		m.Meta = types.StringNull()
	}

	return diags
}
