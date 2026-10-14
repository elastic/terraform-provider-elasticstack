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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPopulateFromAPI(t *testing.T) {
	t.Run("all fields present", func(t *testing.T) {
		var m fieldDefinitionModel
		diags := m.populateFromAPI(context.Background(), "ops", &kbapi.CasesFieldDefinitionResponse{
			FieldDefinitionId: "fd-1",
			Owner:             "cases",
			Name:              "f",
			Definition:        "name: f\ntype: keyword",
			Description:       new("d"),
			IsGlobal:          new(true),
			DisplayOrder:      new(3),
		})
		require.False(t, diags.HasError())
		assert.Equal(t, "ops/fd-1", m.ID.ValueString())
		assert.Equal(t, "ops", m.SpaceID.ValueString())
		assert.Equal(t, "fd-1", m.FieldDefinitionID.ValueString())
		assert.Equal(t, "d", m.Description.ValueString())
		assert.True(t, m.IsGlobal.ValueBool())
		assert.Equal(t, int64(3), m.DisplayOrder.ValueInt64())
	})

	t.Run("nil optional pointers become null", func(t *testing.T) {
		var m fieldDefinitionModel
		diags := m.populateFromAPI(context.Background(), "", &kbapi.CasesFieldDefinitionResponse{
			FieldDefinitionId: "fd-1", Owner: "cases", Name: "f", Definition: "name: f",
		})
		require.False(t, diags.HasError())
		assert.Equal(t, "default/fd-1", m.ID.ValueString())
		assert.True(t, m.Description.IsNull())
		assert.True(t, m.IsGlobal.IsNull())
		assert.True(t, m.DisplayOrder.IsNull())
	})
}

func TestToAPIModels(t *testing.T) {
	base := fieldDefinitionModel{
		Owner:      types.StringValue("cases"),
		Definition: customtypes.NewNormalizedYamlValue("name: f\ntype: keyword"),
	}

	t.Run("unknown and null optionals are omitted", func(t *testing.T) {
		m := base
		m.Name = types.StringUnknown()
		m.Description = types.StringNull()
		m.IsGlobal = types.BoolUnknown()

		create := m.toAPICreateModel()
		assert.Equal(t, "cases", create.Owner)
		assert.Nil(t, create.Name)
		assert.Nil(t, create.Description)
		assert.Nil(t, create.IsGlobal)

		update := m.toAPIUpdateModel()
		assert.Nil(t, update.Name)
		assert.Nil(t, update.Description)
		assert.Nil(t, update.IsGlobal)
	})

	t.Run("known optionals are sent", func(t *testing.T) {
		m := base
		m.Name = types.StringValue("f")
		m.Description = types.StringValue("d")
		m.IsGlobal = types.BoolValue(false)

		create := m.toAPICreateModel()
		assert.Equal(t, "f", *create.Name)
		assert.Equal(t, "d", *create.Description)
		assert.False(t, *create.IsGlobal)

		update := m.toAPIUpdateModel()
		assert.Equal(t, "f", *update.Name)
		assert.Equal(t, "d", *update.Description)
		assert.False(t, *update.IsGlobal)
	})
}

func TestVersionRequirement(t *testing.T) {
	reqs, diags := fieldDefinitionModel{}.GetVersionRequirements(context.Background())
	require.False(t, diags.HasError())
	require.Len(t, reqs, 1)
	assert.Equal(t, "9.5.0", reqs[0].MinVersion.String())
	assert.Contains(t, reqs[0].ErrorMessage, "9.5.0")
}

func TestSchema(t *testing.T) {
	s := getResourceSchema(context.Background())
	assert.Contains(t, s.MarkdownDescription, "Technical Preview")
	for _, attr := range []string{"id", "space_id", "field_definition_id", "owner", "name", "definition", "description", "is_global", "display_order"} {
		assert.Contains(t, s.Attributes, attr)
	}
	assert.True(t, s.Attributes["description"].IsComputed() && s.Attributes["description"].IsOptional())
	assert.True(t, s.Attributes["is_global"].IsComputed() && s.Attributes["is_global"].IsOptional())
	assert.IsType(t, customtypes.NormalizedYamlType{}, s.Attributes["definition"].GetType())
}
