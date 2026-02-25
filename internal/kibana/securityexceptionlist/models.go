// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package securityexceptionlist

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
		Type:        kbapi.SecurityExceptionsAPIExceptionListType(m.Type.ValueString()),
	}

	// Set optional list_id
	if typeutils.IsKnown(m.ListID) {
		listID := m.ListID.ValueString()
		req.ListId = &listID
	}

	// Set optional namespace_type
	if typeutils.IsKnown(m.NamespaceType) {
		nsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(m.NamespaceType.ValueString())
		req.NamespaceType = &nsType
	}

	// Set optional os_types
	if typeutils.IsKnown(m.OsTypes) {
		osTypes := typeutils.SetTypeAs[string](ctx, m.OsTypes, path.Empty(), &diags)
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
	if typeutils.IsKnown(m.Tags) {
		tags := typeutils.SetTypeAs[string](ctx, m.Tags, path.Empty(), &diags)
		if diags.HasError() {
			return nil, diags
		}
		if len(tags) > 0 {
			req.Tags = &tags
		}
	}

	// Set optional meta
	if typeutils.IsKnown(m.Meta) {
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
func (m *ExceptionListModel) toUpdateRequest(ctx context.Context, resourceID string) (*kbapi.UpdateExceptionListJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics
	id := resourceID
	req := &kbapi.UpdateExceptionListJSONRequestBody{
		Id:          &id,
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
		// Type is required by the API even though it has RequiresReplace in the schema
		// The API will reject updates without this field, even though the value cannot change
		Type: kbapi.SecurityExceptionsAPIExceptionListType(m.Type.ValueString()),
	}

	// Set optional namespace_type
	if typeutils.IsKnown(m.NamespaceType) {
		nsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(m.NamespaceType.ValueString())
		req.NamespaceType = &nsType
	}

	// Set optional os_types
	if typeutils.IsKnown(m.OsTypes) {
		osTypes := typeutils.SetTypeAs[string](ctx, m.OsTypes, path.Empty(), &diags)
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
	if typeutils.IsKnown(m.Tags) {
		tags := typeutils.SetTypeAs[string](ctx, m.Tags, path.Empty(), &diags)
		if diags.HasError() {
			return nil, diags
		}
		if len(tags) > 0 {
			tagsArray := tags
			req.Tags = &tagsArray
		}
	}

	// Set optional meta
	if typeutils.IsKnown(m.Meta) {
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
	compID := clients.CompositeID{
		ClusterID:  m.SpaceID.ValueString(),
		ResourceID: typeutils.StringishValue(apiList.Id).ValueString(),
	}
	m.ID = types.StringValue(compID.String())

	m.ListID = typeutils.StringishValue(apiList.ListId)
	m.Name = typeutils.StringishValue(apiList.Name)
	m.Description = typeutils.StringishValue(apiList.Description)
	m.Type = typeutils.StringishValue(apiList.Type)
	m.NamespaceType = typeutils.StringishValue(apiList.NamespaceType)
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
	} else if m.OsTypes.IsUnknown() {
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
