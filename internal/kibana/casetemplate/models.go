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

package casetemplate

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// minVersion is the first Elastic Stack patch release that includes the standalone Cases templates API.
var minVersion = version.Must(version.NewVersion("9.5.2"))

type templateModel struct {
	entitycore.KibanaConnectionField
	entitycore.ResourceTimeoutsField
	ID              types.String                    `tfsdk:"id"`
	SpaceID         types.String                    `tfsdk:"space_id"`
	TemplateID      types.String                    `tfsdk:"template_id"`
	Owner           types.String                    `tfsdk:"owner"`
	Name            types.String                    `tfsdk:"name"`
	DefinitionYAML  customtypes.NormalizedYamlValue `tfsdk:"definition_yaml"`
	Description     types.String                    `tfsdk:"description"`
	Tags            types.List                      `tfsdk:"tags"`
	IsEnabled       types.Bool                      `tfsdk:"is_enabled"`
	TemplateVersion types.Int64                     `tfsdk:"template_version"`
	FieldCount      types.Int64                     `tfsdk:"field_count"`
	UsageCount      types.Int64                     `tfsdk:"usage_count"`
	IsDefault       types.Bool                      `tfsdk:"is_default"`
}

func (m templateModel) GetID() types.String         { return m.ID }
func (m templateModel) GetResourceID() types.String { return m.TemplateID }
func (m templateModel) GetSpaceID() types.String    { return m.SpaceID }

// GetVersionRequirements enforces the 9.5.2 minimum on every lifecycle operation.
func (m templateModel) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return entitycore.SingleVersionRequirement(
		*minVersion,
		"The Kibana Cases templates API requires Elastic Stack v"+minVersion.Core().String()+" or later.",
	), nil
}

var (
	_ entitycore.KibanaResourceModel     = templateModel{}
	_ entitycore.WithVersionRequirements = templateModel{}
)

func (m *templateModel) populateFromAPI(_ context.Context, spaceID string, data *kbapi.CasesTemplateV2Response) diag.Diagnostics {
	if data == nil {
		return nil
	}

	spaceID = clients.EffectiveSpaceID(spaceID)

	m.ID = clients.CompositeIDValue(spaceID, data.TemplateId)
	m.SpaceID = types.StringValue(spaceID)
	m.TemplateID = types.StringValue(data.TemplateId)
	m.Owner = types.StringValue(data.Owner)
	m.Name = types.StringValue(data.Name)
	m.DefinitionYAML = customtypes.NewNormalizedYamlValue(data.DefinitionString)
	m.Description = types.StringPointerValue(data.Description)
	m.Tags = tagsToList(data.Tags)
	m.IsEnabled = types.BoolPointerValue(data.IsEnabled)
	m.TemplateVersion = types.Int64Value(int64(data.TemplateVersion))
	m.FieldCount = typeutils.IntPointerToInt64Value(data.FieldCount)
	m.UsageCount = typeutils.IntPointerToInt64Value(data.UsageCount)
	m.IsDefault = types.BoolPointerValue(data.IsDefault)

	return nil
}

func tagsToList(tags *kbapi.CasesTemplateTags) types.List {
	if tags == nil {
		return types.ListNull(types.StringType)
	}
	return typeutils.StringsToListMust(*tags)
}

// toAPIModel builds the create and full-replace update body. Null attributes
// are omitted, which the API treats as clearing the stored value on update.
func (m templateModel) toAPIModel() kbapi.CasesTemplateWriteRequest {
	req := kbapi.CasesTemplateWriteRequest{
		Owner:       kbapi.CasesOwner(m.Owner.ValueString()),
		Definition:  m.DefinitionYAML.ValueString(),
		Name:        typeutils.ValueStringPointer(m.Name),
		Description: typeutils.ValueStringPointer(m.Description),
		IsEnabled:   typeutils.OptionalBool(m.IsEnabled),
	}
	if typeutils.IsKnown(m.Tags) {
		tags := typeutils.ListToStringsMust(m.Tags)
		if tags == nil {
			tags = kbapi.CasesTemplateTags{}
		}
		req.Tags = &tags
	}
	return req
}
