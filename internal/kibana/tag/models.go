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

package tag

import (
	"context"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tagBaseModel struct {
	entitycore.KibanaConnectionField
	ID          types.String `tfsdk:"id"`
	TagID       types.String `tfsdk:"tag_id"`
	SpaceID     types.String `tfsdk:"space_id"`
	Name        types.String `tfsdk:"name"`
	Color       types.String `tfsdk:"color"`
	Description types.String `tfsdk:"description"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

type tagModel struct {
	entitycore.ResourceTimeoutsField
	tagBaseModel
}

var (
	_ entitycore.KibanaResourceModel     = tagModel{}
	_ entitycore.WithVersionRequirements = tagModel{}

	tagMinVersion = version.Must(version.NewVersion("9.5.0-SNAPSHOT"))
)

func (m tagBaseModel) GetID() types.String         { return m.ID }
func (m tagBaseModel) GetResourceID() types.String { return m.TagID }
func (m tagBaseModel) GetSpaceID() types.String    { return m.SpaceID }

func (m *tagBaseModel) setCompositeIdentity(spaceID, tagID string) {
	m.ID = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: tagID}).String())
	m.TagID = types.StringValue(tagID)
	m.SpaceID = types.StringValue(spaceID)
}

func (tagBaseModel) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{
		{
			MinVersion:   *tagMinVersion,
			ErrorMessage: "Kibana tags require Elastic Stack v9.5.0 or later (introduced in Kibana 9.5).",
		},
	}, nil
}

func (m *tagBaseModel) populateFromAPI(spaceID string, detail *kibanaoapi.TagDetail) {
	if detail == nil {
		return
	}

	if spaceID == "" {
		spaceID = clients.DefaultSpaceID
	}

	m.setCompositeIdentity(spaceID, detail.ID)
	m.Name = types.StringValue(detail.Name)
	m.Color = types.StringValue(detail.Color)
	m.Description = optionalStringPointerValue(detail.Description)
	m.CreatedAt = types.StringPointerValue(detail.CreatedAt)
	m.UpdatedAt = types.StringPointerValue(detail.UpdatedAt)
}

func optionalStringPointerValue(value *string) types.String {
	if value == nil || strings.TrimSpace(*value) == "" {
		return types.StringNull()
	}
	return types.StringValue(*value)
}

func (m tagModel) toAPIModel(includeDescription bool) kbapi.KibanaHTTPAPIsKbnTagsRequestAttributes {
	body := kbapi.KibanaHTTPAPIsKbnTagsRequestAttributes{
		Name: m.Name.ValueString(),
	}

	if color, ok := m.resolvedColor(nil); ok {
		body.Color = &color
	}

	if includeDescription && typeutils.IsKnown(m.Description) {
		desc := strings.TrimSpace(m.Description.ValueString())
		if desc != "" {
			body.Description = &desc
		}
	}

	return body
}

func (m tagModel) toUpdateAPIModel(prior *tagModel) kbapi.KibanaHTTPAPIsKbnTagsRequestAttributes {
	return m.withResolvedColor(prior).toAPIModel(true)
}

func (m tagModel) withResolvedColor(prior *tagModel) tagModel {
	if typeutils.IsKnown(m.Color) || prior == nil {
		return m
	}

	if !typeutils.IsKnown(prior.Color) {
		return m
	}

	updated := m
	updated.Color = prior.Color
	return updated
}

func (m tagModel) resolvedColor(prior *tagModel) (string, bool) {
	model := m.withResolvedColor(prior)
	if !typeutils.IsKnown(model.Color) {
		return "", false
	}
	return model.Color.ValueString(), true
}

type tagCreateAction int

const (
	tagCreateActionPOST tagCreateAction = iota
	tagCreateActionPUT
	tagCreateActionRejectExisting
)

func tagCreateActionForExisting(explicitTagID bool, exists bool) tagCreateAction {
	if !explicitTagID {
		return tagCreateActionPOST
	}
	if exists {
		return tagCreateActionRejectExisting
	}
	return tagCreateActionPUT
}
