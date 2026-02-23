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

package securitylist

import (
	"context"
	"encoding/json"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Model struct {
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
func (m *Model) toCreateRequest() (*kbapi.CreateListJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := &kbapi.CreateListJSONRequestBody{
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
		Type:        kbapi.SecurityListsAPIListType(m.Type.ValueString()),
	}

	// Set optional fields
	if typeutils.IsKnown(m.ListID) {
		id := m.ListID.ValueString()
		req.Id = &id
	}

	if typeutils.IsKnown(m.Deserializer) {
		deserializer := m.Deserializer.ValueString()
		req.Deserializer = &deserializer
	}

	if typeutils.IsKnown(m.Serializer) {
		serializer := m.Serializer.ValueString()
		req.Serializer = &serializer
	}

	if typeutils.IsKnown(m.Meta) {
		var metaMap kbapi.SecurityListsAPIListMetadata
		unmarshalDiags := m.Meta.Unmarshal(&metaMap)
		diags.Append(unmarshalDiags...)
		if diags.HasError() {
			return nil, diags
		}
		req.Meta = &metaMap
	}

	if typeutils.IsKnown(m.Version) {
		version := int(m.Version.ValueInt64())
		req.Version = &version
	}

	return req, diags
}

// toUpdateRequest converts the Terraform model to API update request
func (m *Model) toUpdateRequest() (*kbapi.UpdateListJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := &kbapi.UpdateListJSONRequestBody{
		Id:          m.ListID.ValueString(),
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
	}

	// Set optional fields
	if typeutils.IsKnown(m.VersionID) {
		versionID := m.VersionID.ValueString()
		req.UnderscoreVersion = &versionID
	}

	if typeutils.IsKnown(m.Meta) {
		var metaMap kbapi.SecurityListsAPIListMetadata
		unmarshalDiags := m.Meta.Unmarshal(&metaMap)
		diags.Append(unmarshalDiags...)
		if diags.HasError() {
			return nil, diags
		}
		req.Meta = &metaMap
	}

	if typeutils.IsKnown(m.Version) {
		version := int(m.Version.ValueInt64())
		req.Version = &version
	}

	return req, diags
}

// fromAPI converts the API response to Terraform model
func (m *Model) fromAPI(ctx context.Context, apiList *kbapi.SecurityListsAPIList) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	// Create composite ID from space_id and list_id
	compID := clients.CompositeID{
		ClusterID:  m.SpaceID.ValueString(),
		ResourceID: apiList.Id,
	}
	m.ID = types.StringValue(compID.String())

	m.ListID = typeutils.StringishValue(apiList.Id)
	m.Name = typeutils.StringishValue(apiList.Name)
	m.Description = typeutils.StringishValue(apiList.Description)
	m.Type = typeutils.StringishValue(apiList.Type)
	m.Immutable = types.BoolValue(apiList.Immutable)
	m.Version = types.Int64Value(int64(apiList.Version))
	m.TieBreakerID = types.StringValue(apiList.TieBreakerId)
	m.CreatedAt = types.StringValue(apiList.CreatedAt.Format(time.RFC3339))
	m.CreatedBy = types.StringValue(apiList.CreatedBy)
	m.UpdatedAt = types.StringValue(apiList.UpdatedAt.Format(time.RFC3339))
	m.UpdatedBy = types.StringValue(apiList.UpdatedBy)

	// Set optional _version field
	m.VersionID = typeutils.StringishPointerValue(apiList.UnderscoreVersion)

	m.Deserializer = typeutils.StringishPointerValue(apiList.Deserializer)

	m.Serializer = typeutils.StringishPointerValue(apiList.Serializer)

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
