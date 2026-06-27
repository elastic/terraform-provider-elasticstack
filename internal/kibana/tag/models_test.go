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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagModel_GetVersionRequirements(t *testing.T) {
	t.Parallel()

	reqs, diags := tagModel{}.GetVersionRequirements(context.Background())
	require.False(t, diags.HasError())
	require.Len(t, reqs, 1)
	require.Equal(t, *tagMinVersion, reqs[0].MinVersion)
	require.Contains(t, reqs[0].ErrorMessage, "9.5.0")
}

func TestTagModel_ImplementsEntityCoreContracts(t *testing.T) {
	t.Parallel()

	var _ entitycore.KibanaResourceModel = tagModel{}
	var _ entitycore.WithVersionRequirements = tagModel{}
	var _ entitycore.WithVersionRequirements = (*tagsDataSourceModel)(nil)
}

func TestTagModel_toAPIModel(t *testing.T) {
	t.Parallel()

	t.Run("omits nil color and description", func(t *testing.T) {
		model := tagModel{
			tagBaseModel: tagBaseModel{
				Name: types.StringValue("staging"),
			},
		}

		body := model.toAPIModel()
		assert.Equal(t, "staging", body.Name)
		assert.Nil(t, body.Color)
		assert.Nil(t, body.Description)
	})

	t.Run("includes color when set", func(t *testing.T) {
		model := tagModel{
			tagBaseModel: tagBaseModel{
				Name:  types.StringValue("staging"),
				Color: types.StringValue("#FF0000"),
			},
		}

		body := model.toAPIModel()
		require.NotNil(t, body.Color)
		assert.Equal(t, "#FF0000", *body.Color)
	})

	t.Run("normalizes empty description to absent", func(t *testing.T) {
		model := tagModel{
			tagBaseModel: tagBaseModel{
				Name:        types.StringValue("staging"),
				Description: types.StringValue("   "),
			},
		}

		body := model.toAPIModel()
		assert.Nil(t, body.Description)
	})

	t.Run("includes non-empty description", func(t *testing.T) {
		model := tagModel{
			tagBaseModel: tagBaseModel{
				Name:        types.StringValue("staging"),
				Description: types.StringValue("prod workloads"),
			},
		}

		body := model.toAPIModel()
		require.NotNil(t, body.Description)
		assert.Equal(t, "prod workloads", *body.Description)
	})
}

func TestTagModel_toUpdateAPIModel(t *testing.T) {
	t.Parallel()

	t.Run("preserves prior color when plan color is unknown", func(t *testing.T) {
		plan := tagModel{
			tagBaseModel: tagBaseModel{
				Name:  types.StringValue("staging-v2"),
				Color: types.StringUnknown(),
			},
		}
		prior := &tagModel{
			tagBaseModel: tagBaseModel{
				Color: types.StringValue("#AABBCC"),
			},
		}

		body := plan.toUpdateAPIModel(prior)
		require.NotNil(t, body.Color)
		assert.Equal(t, "#AABBCC", *body.Color)
	})

	t.Run("uses plan color when known", func(t *testing.T) {
		plan := tagModel{
			tagBaseModel: tagBaseModel{
				Name:  types.StringValue("staging-v2"),
				Color: types.StringValue("#112233"),
			},
		}
		prior := &tagModel{
			tagBaseModel: tagBaseModel{
				Color: types.StringValue("#AABBCC"),
			},
		}

		body := plan.toUpdateAPIModel(prior)
		require.NotNil(t, body.Color)
		assert.Equal(t, "#112233", *body.Color)
	})

	t.Run("omits color when plan and prior color are unknown", func(t *testing.T) {
		plan := tagModel{
			tagBaseModel: tagBaseModel{
				Name:  types.StringValue("staging-v2"),
				Color: types.StringUnknown(),
			},
		}
		prior := &tagModel{
			tagBaseModel: tagBaseModel{
				Color: types.StringUnknown(),
			},
		}

		body := plan.toUpdateAPIModel(prior)
		assert.Nil(t, body.Color)
	})

	t.Run("clears prior description when plan omits description", func(t *testing.T) {
		plan := tagModel{
			tagBaseModel: tagBaseModel{
				Name:        types.StringValue("staging-v2"),
				Description: types.StringNull(),
			},
		}
		prior := &tagModel{
			tagBaseModel: tagBaseModel{
				Description: types.StringValue("old description"),
			},
		}

		body := plan.toUpdateAPIModel(prior)
		require.NotNil(t, body.Description)
		assert.Empty(t, *body.Description)
	})
}

func TestTagModel_populateFromAPI(t *testing.T) {
	t.Parallel()

	t.Run("populated response", func(t *testing.T) {
		desc := "Production"
		createdAt := "2026-01-01T00:00:00.000Z"
		updatedAt := "2026-01-02T00:00:00.000Z"
		managed := false

		var model tagBaseModel
		model.populateFromAPI("ops", &kibanaoapi.TagDetail{
			ID:          "abc-123",
			Name:        "prod",
			Color:       "#112233",
			Description: &desc,
			CreatedAt:   &createdAt,
			UpdatedAt:   &updatedAt,
			Managed:     &managed,
		})

		assert.Equal(t, "ops/abc-123", model.ID.ValueString())
		assert.Equal(t, "abc-123", model.TagID.ValueString())
		assert.Equal(t, "ops", model.SpaceID.ValueString())
		assert.Equal(t, "prod", model.Name.ValueString())
		assert.Equal(t, "#112233", model.Color.ValueString())
		assert.Equal(t, "Production", model.Description.ValueString())
		assert.Equal(t, createdAt, model.CreatedAt.ValueString())
		assert.Equal(t, updatedAt, model.UpdatedAt.ValueString())
	})

	t.Run("sparse response", func(t *testing.T) {
		var model tagBaseModel
		model.populateFromAPI("", &kibanaoapi.TagDetail{
			ID:    "abc-123",
			Name:  "prod",
			Color: "#112233",
		})

		assert.Equal(t, "default/abc-123", model.ID.ValueString())
		assert.True(t, model.Description.IsNull())
		assert.True(t, model.CreatedAt.IsNull())
		assert.True(t, model.UpdatedAt.IsNull())
	})
}

func TestTagItemFromAPI(t *testing.T) {
	t.Parallel()

	desc := "Production"
	managed := true
	item := tagItemFromAPI(kibanaoapi.TagDetail{
		ID:          "abc-123",
		Name:        "prod",
		Color:       "#112233",
		Description: &desc,
		Managed:     &managed,
	})

	assert.Equal(t, "abc-123", item.ID.ValueString())
	assert.Equal(t, "prod", item.Name.ValueString())
	assert.Equal(t, "#112233", item.Color.ValueString())
	assert.Equal(t, "Production", item.Description.ValueString())
	assert.True(t, item.Managed.ValueBool())
}
