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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPopulateFromAPIMinimal(t *testing.T) {
	var m templateModel
	diags := m.populateFromAPI(context.Background(), "", &kbapi.CasesTemplateV2Response{
		TemplateId:       "tpl-1",
		TemplateVersion:  1,
		Name:             "t",
		Owner:            "cases",
		DefinitionString: "name: t\n",
	})
	require.False(t, diags.HasError())

	assert.Equal(t, "default/tpl-1", m.ID.ValueString())
	assert.Equal(t, "default", m.SpaceID.ValueString())
	assert.Equal(t, "tpl-1", m.TemplateID.ValueString())
	assert.Equal(t, int64(1), m.TemplateVersion.ValueInt64())
	assert.Equal(t, "t", m.Name.ValueString())
	assert.Equal(t, "cases", m.Owner.ValueString())
	assert.Equal(t, "name: t\n", m.DefinitionYAML.ValueString())
	assert.True(t, m.Description.IsNull())
	assert.True(t, m.Tags.IsNull())
	assert.True(t, m.IsEnabled.IsNull())
	assert.True(t, m.FieldCount.IsNull())
	assert.True(t, m.UsageCount.IsNull())
	assert.True(t, m.IsDefault.IsNull())
}

func TestPopulateFromAPIAllFields(t *testing.T) {
	description := "desc"
	tags := kbapi.CasesTemplateTags{"z", "a", "z"}
	enabled, isDefault := false, true
	fieldCount, usageCount := 3, 7

	var m templateModel
	diags := m.populateFromAPI(context.Background(), "ops", &kbapi.CasesTemplateV2Response{
		TemplateId:       "tpl-1",
		TemplateVersion:  4,
		Name:             "t",
		Owner:            "securitySolution",
		DefinitionString: "name: t\n",
		Description:      &description,
		Tags:             &tags,
		IsEnabled:        &enabled,
		IsDefault:        &isDefault,
		FieldCount:       &fieldCount,
		UsageCount:       &usageCount,
	})
	require.False(t, diags.HasError())

	assert.Equal(t, "ops/tpl-1", m.ID.ValueString())
	assert.Equal(t, "desc", m.Description.ValueString())
	assert.Equal(t, []attr.Value{types.StringValue("z"), types.StringValue("a"), types.StringValue("z")}, m.Tags.Elements())
	assert.False(t, m.IsEnabled.ValueBool())
	assert.True(t, m.IsDefault.ValueBool())
	assert.Equal(t, int64(3), m.FieldCount.ValueInt64())
	assert.Equal(t, int64(7), m.UsageCount.ValueInt64())
}

func TestPopulateFromAPIEmptyValuesStayDistinctFromNull(t *testing.T) {
	description := ""
	tags := kbapi.CasesTemplateTags{}

	var m templateModel
	diags := m.populateFromAPI(context.Background(), "default", &kbapi.CasesTemplateV2Response{
		TemplateId:  "tpl-1",
		Description: &description,
		Tags:        &tags,
	})
	require.False(t, diags.HasError())

	assert.False(t, m.Description.IsNull())
	assert.Equal(t, "", m.Description.ValueString())
	assert.False(t, m.Tags.IsNull())
	assert.Empty(t, m.Tags.Elements())
}

func TestToAPIModel(t *testing.T) {
	t.Run("omitted optional attributes are not sent", func(t *testing.T) {
		m := templateModel{
			Owner:          types.StringValue("cases"),
			Name:           types.StringValue("t"),
			DefinitionYAML: customtypes.NewNormalizedYamlValue("name: t\n"),
			Description:    types.StringNull(),
			Tags:           types.ListNull(types.StringType),
			IsEnabled:      types.BoolUnknown(),
		}
		req := m.toAPIModel()

		assert.Equal(t, kbapi.CasesOwnerCases, req.Owner)
		assert.Equal(t, "t", *req.Name)
		assert.Equal(t, "name: t\n", req.Definition)
		assert.Nil(t, req.Description)
		assert.Nil(t, req.Tags)
		assert.Nil(t, req.IsEnabled)
	})

	t.Run("explicit empty values are sent", func(t *testing.T) {
		m := templateModel{
			Owner:          types.StringValue("observability"),
			Name:           types.StringValue("t"),
			DefinitionYAML: customtypes.NewNormalizedYamlValue("name: t\n"),
			Description:    types.StringValue(""),
			Tags:           types.ListValueMust(types.StringType, []attr.Value{}),
			IsEnabled:      types.BoolValue(false),
		}
		req := m.toAPIModel()

		require.NotNil(t, req.Description)
		assert.Equal(t, "", *req.Description)
		require.NotNil(t, req.Tags)
		assert.Empty(t, *req.Tags)
		assert.False(t, *req.IsEnabled)
	})

	t.Run("tags keep order and duplicates", func(t *testing.T) {
		m := templateModel{
			Tags: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("z"), types.StringValue("a"), types.StringValue("z")}),
		}
		assert.Equal(t, kbapi.CasesTemplateTags{"z", "a", "z"}, *m.toAPIModel().Tags)
	})
}

func TestVersionRequirement(t *testing.T) {
	reqs, diags := templateModel{}.GetVersionRequirements(context.Background())
	require.False(t, diags.HasError())
	require.Len(t, reqs, 1)
	assert.Equal(t, "9.5.2", reqs[0].MinVersion.String())
	assert.Contains(t, reqs[0].ErrorMessage, "9.5.2")
}
