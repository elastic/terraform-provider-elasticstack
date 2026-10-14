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

package casefielddefinition

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

// minVersion is the first Elastic Stack version with the Cases field library API.
var minVersion = version.Must(version.NewVersion("9.6.0-SNAPSHOT"))

type fieldDefinitionModel struct {
	entitycore.KibanaConnectionField
	entitycore.ResourceTimeoutsField
	ID                types.String                    `tfsdk:"id"`
	SpaceID           types.String                    `tfsdk:"space_id"`
	FieldDefinitionID types.String                    `tfsdk:"field_definition_id"`
	Owner             types.String                    `tfsdk:"owner"`
	Name              types.String                    `tfsdk:"name"`
	Definition        customtypes.NormalizedYamlValue `tfsdk:"definition"`
	Description       types.String                    `tfsdk:"description"`
	IsGlobal          types.Bool                      `tfsdk:"is_global"`
	DisplayOrder      types.Int64                     `tfsdk:"display_order"`
}

func (m fieldDefinitionModel) GetID() types.String         { return m.ID }
func (m fieldDefinitionModel) GetResourceID() types.String { return m.FieldDefinitionID }
func (m fieldDefinitionModel) GetSpaceID() types.String    { return m.SpaceID }

// GetVersionRequirements enforces the 9.6.0 minimum on every lifecycle operation.
func (m fieldDefinitionModel) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return entitycore.SingleVersionRequirement(
		*minVersion,
		"The Kibana Cases field library requires Elastic Stack v"+minVersion.Core().String()+" or later.",
	), nil
}

var (
	_ entitycore.KibanaResourceModel     = fieldDefinitionModel{}
	_ entitycore.WithVersionRequirements = fieldDefinitionModel{}
)

func (m *fieldDefinitionModel) populateFromAPI(_ context.Context, spaceID string, data *kbapi.CasesFieldDefinitionResponse) diag.Diagnostics {
	if data == nil {
		return nil
	}

	spaceID = clients.EffectiveSpaceID(spaceID)

	m.ID = clients.CompositeIDValue(spaceID, data.FieldDefinitionId)
	m.SpaceID = types.StringValue(spaceID)
	m.FieldDefinitionID = types.StringValue(data.FieldDefinitionId)
	m.Owner = types.StringValue(data.Owner)
	m.Name = types.StringValue(data.Name)
	m.Definition = customtypes.NewNormalizedYamlValue(data.Definition)
	m.Description = types.StringPointerValue(data.Description)
	m.IsGlobal = types.BoolPointerValue(data.IsGlobal)
	m.DisplayOrder = typeutils.IntPointerToInt64Value(data.DisplayOrder)

	return nil
}

func (m fieldDefinitionModel) toAPICreateModel() kbapi.CasesFieldDefinitionWriteRequest {
	return kbapi.CasesFieldDefinitionWriteRequest{
		Owner:       m.Owner.ValueString(),
		Definition:  m.Definition.ValueString(),
		Name:        typeutils.ValueStringPointer(m.Name),
		Description: typeutils.ValueStringPointer(m.Description),
		IsGlobal:    typeutils.OptionalBool(m.IsGlobal),
	}
}

func (m fieldDefinitionModel) toAPIUpdateModel() kbapi.CasesFieldDefinitionPutRequest {
	return kbapi.CasesFieldDefinitionPutRequest{
		Owner:       m.Owner.ValueString(),
		Definition:  m.Definition.ValueString(),
		Name:        typeutils.ValueStringPointer(m.Name),
		Description: typeutils.ValueStringPointer(m.Description),
		IsGlobal:    typeutils.OptionalBool(m.IsGlobal),
	}
}
