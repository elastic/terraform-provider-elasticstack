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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/kbschema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (m ExceptionListModel) GetID() types.String { return m.ID }
func (m ExceptionListModel) GetResourceID() types.String {
	if compID, _ := clients.CompositeIDFromStr(m.ID.ValueString()); compID != nil {
		return types.StringValue(compID.ResourceID)
	}
	return types.StringValue("")
}
func (m ExceptionListModel) GetSpaceID() types.String        { return m.SpaceID }
func (m ExceptionListModel) GetKibanaConnection() types.List { return m.KibanaConnection }

var _ entitycore.KibanaResourceModel = ExceptionListModel{}

type ExceptionListModel struct {
	entitycore.ResourceTimeoutsField
	ID               types.String         `tfsdk:"id"`
	KibanaConnection types.List           `tfsdk:"kibana_connection"`
	SpaceID          types.String         `tfsdk:"space_id"`
	ListID           types.String         `tfsdk:"list_id"`
	Name             types.String         `tfsdk:"name"`
	Description      types.String         `tfsdk:"description"`
	Type             types.String         `tfsdk:"type"`
	NamespaceType    types.String         `tfsdk:"namespace_type"`
	OsTypes          types.Set            `tfsdk:"os_types"`
	Tags             types.Set            `tfsdk:"tags"`
	Meta             jsontypes.Normalized `tfsdk:"meta"`
	CreatedAt        types.String         `tfsdk:"created_at"`
	CreatedBy        types.String         `tfsdk:"created_by"`
	UpdatedAt        types.String         `tfsdk:"updated_at"`
	UpdatedBy        types.String         `tfsdk:"updated_by"`
	Immutable        types.Bool           `tfsdk:"immutable"`
	TieBreakerID     types.String         `tfsdk:"tie_breaker_id"`
}

type exceptionListOptionalFields struct {
	namespaceType *kbapi.SecurityExceptionsAPIExceptionNamespaceType
	osTypes       *kbapi.SecurityExceptionsAPIExceptionListOsTypeArray
	tags          *kbapi.SecurityExceptionsAPIExceptionListTags
	meta          *kbapi.SecurityExceptionsAPIExceptionListMeta
}

func (m *ExceptionListModel) optionalExceptionListFields(ctx context.Context) (exceptionListOptionalFields, diag.Diagnostics) {
	var diags diag.Diagnostics
	var out exceptionListOptionalFields

	if typeutils.IsKnown(m.NamespaceType) {
		nsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(m.NamespaceType.ValueString())
		out.namespaceType = &nsType
	}

	if typeutils.IsKnown(m.OsTypes) {
		osTypes := typeutils.SetTypeAs[string](ctx, m.OsTypes, path.Empty(), &diags)
		if diags.HasError() {
			return out, diags
		}
		osTypesArray := make(kbapi.SecurityExceptionsAPIExceptionListOsTypeArray, len(osTypes))
		for i, osType := range osTypes {
			osTypesArray[i] = kbapi.SecurityExceptionsAPIExceptionListOsType(osType)
		}
		out.osTypes = &osTypesArray
	}

	if typeutils.IsKnown(m.Tags) {
		tags := typeutils.SetTypeAs[string](ctx, m.Tags, path.Empty(), &diags)
		if diags.HasError() {
			return out, diags
		}
		out.tags = &tags
	}

	if typeutils.IsKnown(m.Meta) {
		var meta kbapi.SecurityExceptionsAPIExceptionListMeta
		diags.Append(m.Meta.Unmarshal(&meta)...)
		if diags.HasError() {
			return out, diags
		}
		out.meta = &meta
	}

	return out, diags
}

// toCreateRequest converts the Terraform model to API create request
func (m *ExceptionListModel) toCreateRequest(ctx context.Context) (*kbapi.CreateExceptionListJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := &kbapi.CreateExceptionListJSONRequestBody{
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
		Type:        kbapi.SecurityExceptionsAPIExceptionListType(m.Type.ValueString()),
	}

	if typeutils.IsKnown(m.ListID) {
		listID := m.ListID.ValueString()
		req.ListId = &listID
	}

	opt, d := m.optionalExceptionListFields(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	req.NamespaceType = opt.namespaceType
	req.OsTypes = opt.osTypes
	req.Tags = opt.tags
	req.Meta = opt.meta

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

	opt, d := m.optionalExceptionListFields(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	req.NamespaceType = opt.namespaceType
	req.OsTypes = opt.osTypes
	req.Tags = opt.tags
	req.Meta = opt.meta

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
	m.CreatedAt = types.StringValue(apiList.CreatedAt.Format(kbschema.KibanaTimestampLayout))
	m.CreatedBy = types.StringValue(apiList.CreatedBy)
	m.UpdatedAt = types.StringValue(apiList.UpdatedAt.Format(kbschema.KibanaTimestampLayout))
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
	} else if m.Tags.IsUnknown() {
		// Same as os_types above (fixed in #1740): only collapse to null when
		// the plan value was Unknown. Preserving a Known-empty Set avoids
		// "produced inconsistent result after apply: .tags: was null, but now
		// ..." when config sets `tags = []`.
		m.Tags = types.SetNull(types.StringType)
	}

	// Set optional meta
	m.Meta = typeutils.MarshalToNormalized(apiList.Meta, path.Root("meta"), &diags)

	return diags
}
