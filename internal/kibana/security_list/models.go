package security_list

import (
	"context"
	"encoding/json"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SecurityListModel struct {
	ID           types.String         `tfsdk:"id"`
	SpaceID      types.String         `tfsdk:"space_id"`
	ListID       types.String         `tfsdk:"list_id"`
	Name         types.String         `tfsdk:"name"`
	Description  types.String         `tfsdk:"description"`
	Type         types.String         `tfsdk:"type"`
	Deserializer types.String         `tfsdk:"deserializer"`
	Serializer   types.String         `tfsdk:"serializer"`
	Meta         jsontypes.Normalized `tfsdk:"meta"`
	Version      types.Int64          `tfsdk:"version"`
	VersionID    types.String         `tfsdk:"version_id"`
	Immutable    types.Bool           `tfsdk:"immutable"`
	CreatedAt    types.String         `tfsdk:"created_at"`
	CreatedBy    types.String         `tfsdk:"created_by"`
	UpdatedAt    types.String         `tfsdk:"updated_at"`
	UpdatedBy    types.String         `tfsdk:"updated_by"`
	TieBreakerID types.String         `tfsdk:"tie_breaker_id"`
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
	if utils.IsKnown(m.ListID) {
		id := kbapi.SecurityListsAPIListId(m.ListID.ValueString())
		req.Id = &id
	}

	if utils.IsKnown(m.Deserializer) {
		deserializer := kbapi.SecurityListsAPIListDeserializer(m.Deserializer.ValueString())
		req.Deserializer = &deserializer
	}

	if utils.IsKnown(m.Serializer) {
		serializer := kbapi.SecurityListsAPIListSerializer(m.Serializer.ValueString())
		req.Serializer = &serializer
	}

	if utils.IsKnown(m.Meta) {
		var metaMap kbapi.SecurityListsAPIListMetadata
		unmarshalDiags := m.Meta.Unmarshal(&metaMap)
		diags.Append(unmarshalDiags...)
		if diags.HasError() {
			return nil, diags
		}
		req.Meta = &metaMap
	}

	if utils.IsKnown(m.Version) {
		version := int(m.Version.ValueInt64())
		req.Version = &version
	}

	return req, diags
}

// toUpdateRequest converts the Terraform model to API update request
func (m *SecurityListModel) toUpdateRequest() (*kbapi.UpdateListJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := &kbapi.UpdateListJSONRequestBody{
		Id:          kbapi.SecurityListsAPIListId(m.ListID.ValueString()),
		Name:        kbapi.SecurityListsAPIListName(m.Name.ValueString()),
		Description: kbapi.SecurityListsAPIListDescription(m.Description.ValueString()),
	}

	// Set optional fields
	if utils.IsKnown(m.VersionID) {
		versionID := kbapi.SecurityListsAPIListVersionId(m.VersionID.ValueString())
		req.UnderscoreVersion = &versionID
	}

	if utils.IsKnown(m.Meta) {
		var metaMap kbapi.SecurityListsAPIListMetadata
		unmarshalDiags := m.Meta.Unmarshal(&metaMap)
		diags.Append(unmarshalDiags...)
		if diags.HasError() {
			return nil, diags
		}
		req.Meta = &metaMap
	}

	if utils.IsKnown(m.Version) {
		version := kbapi.SecurityListsAPIListVersion(int(m.Version.ValueInt64()))
		req.Version = &version
	}

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
	m.CreatedAt = types.StringValue(apiList.CreatedAt.Format(time.RFC3339))
	m.CreatedBy = types.StringValue(apiList.CreatedBy)
	m.UpdatedAt = types.StringValue(apiList.UpdatedAt.Format(time.RFC3339))
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
			diags.AddError("Failed to marshal meta field from API response to JSON", err.Error())
			return diags
		}
		m.Meta = jsontypes.NewNormalizedValue(string(metaBytes))
	} else {
		m.Meta = jsontypes.NewNormalizedNull()
	}

	return diags
}
