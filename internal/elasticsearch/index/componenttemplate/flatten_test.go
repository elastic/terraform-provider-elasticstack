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

package componenttemplate

import (
	"context"
	"testing"

	esindex "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/stretchr/testify/require"
)

// TestFlattenTemplateBlock_MappingsAndSettingsEmptyObject pins the empty-object
// normalisation behaviour added for issue #609: both a nil and an empty
// map[string]any decoded from Elasticsearch's response must produce a null
// Terraform value when there is no prior practitioner-authored empty-object
// value to preserve. This is the baseline "omitted field" path.
func TestFlattenTemplateBlock_MappingsAndSettingsEmptyObject(t *testing.T) {
	cases := []struct {
		name             string
		mappings         map[string]any
		settings         map[string]any
		wantMappingsNull bool
		wantSettingsNull bool
	}{
		{
			name:             "nil mappings and nil settings",
			mappings:         nil,
			settings:         nil,
			wantMappingsNull: true,
			wantSettingsNull: true,
		},
		{
			name:             "non-nil empty mappings and empty settings",
			mappings:         map[string]any{},
			settings:         map[string]any{},
			wantMappingsNull: true,
			wantSettingsNull: true,
		},
		{
			name:             "non-empty mappings and settings are emitted as values",
			mappings:         map[string]any{"properties": map[string]any{"a": map[string]any{"type": "keyword"}}},
			settings:         map[string]any{"index": map[string]any{"number_of_shards": "3"}},
			wantMappingsNull: false,
			wantSettingsNull: false,
		},
		{
			name:             "empty mappings with non-empty settings",
			mappings:         map[string]any{},
			settings:         map[string]any{"index": map[string]any{"number_of_shards": "3"}},
			wantMappingsNull: true,
			wantSettingsNull: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tpl := &models.Template{
				Mappings: tc.mappings,
				Settings: tc.settings,
			}
			obj, diags := flattenTemplateBlock(
				context.Background(),
				tpl,
				nil,
				esindex.NewMappingsNull(),
				customtypes.NewIndexSettingsNull(),
			)
			require.False(t, diags.HasError(), "unexpected diags: %v", diags)
			require.False(t, obj.IsNull(), "template object should not be null when template != nil")

			attrs := obj.Attributes()
			mappings, ok := attrs["mappings"]
			require.True(t, ok, "template.mappings attribute missing")
			settings, ok := attrs["settings"]
			require.True(t, ok, "template.settings attribute missing")

			require.Equal(t, tc.wantMappingsNull, mappings.IsNull(), "template.mappings null mismatch")
			require.Equal(t, tc.wantSettingsNull, settings.IsNull(), "template.settings null mismatch")
		})
	}
}

// TestFlattenTemplateBlock_PreservesPriorEmptyObjectOverride pins the
// practitioner-authored empty-object preservation path. When the API omits
// the mappings/settings field (or returns it as `{}`) and the prior Terraform
// value is a known, semantically-empty JSON object (for example because the
// user wrote `mappings = jsonencode({})` in HCL), the flattened state value
// SHALL be the prior value, not null.
//
// This avoids the post-apply consistency error the Plugin Framework raises
// when the planned `"{}"` value collides with a null state value, because the
// framework's ValueSemanticEquality walker short-circuits when either side is
// null and never invokes StringSemanticEquals.
func TestFlattenTemplateBlock_PreservesPriorEmptyObjectOverride(t *testing.T) {
	cases := []struct {
		name             string
		mappings         map[string]any
		settings         map[string]any
		priorMappings    esindex.MappingsValue
		priorSettings    customtypes.IndexSettingsValue
		wantMappingsNull bool
		wantSettingsNull bool
		wantMappings     string
		wantSettings     string
	}{
		{
			name:             "API empty + prior empty-object mappings preserved",
			mappings:         map[string]any{},
			settings:         nil,
			priorMappings:    esindex.NewMappingsValue(`{}`),
			priorSettings:    customtypes.NewIndexSettingsNull(),
			wantMappingsNull: false,
			wantSettingsNull: true,
			wantMappings:     `{}`,
		},
		{
			name:             "API empty + prior empty-object settings preserved",
			mappings:         nil,
			settings:         map[string]any{},
			priorMappings:    esindex.NewMappingsNull(),
			priorSettings:    customtypes.NewIndexSettingsValue(`{}`),
			wantMappingsNull: true,
			wantSettingsNull: false,
			wantSettings:     `{}`,
		},
		{
			name:             "API empty + both prior empty-object preserved",
			mappings:         map[string]any{},
			settings:         map[string]any{},
			priorMappings:    esindex.NewMappingsValue(`{}`),
			priorSettings:    customtypes.NewIndexSettingsValue(`{}`),
			wantMappingsNull: false,
			wantSettingsNull: false,
			wantMappings:     `{}`,
			wantSettings:     `{}`,
		},
		{
			name:             "API empty + null priors stay null",
			mappings:         map[string]any{},
			settings:         map[string]any{},
			priorMappings:    esindex.NewMappingsNull(),
			priorSettings:    customtypes.NewIndexSettingsNull(),
			wantMappingsNull: true,
			wantSettingsNull: true,
		},
		{
			// A prior non-empty value must NOT be preserved if the API response
			// is empty — that would mask genuine out-of-band deletion (the API
			// is authoritative for non-empty content).
			name:             "API empty + prior non-empty mappings not preserved",
			mappings:         map[string]any{},
			settings:         nil,
			priorMappings:    esindex.NewMappingsValue(`{"properties":{"f":{"type":"keyword"}}}`),
			priorSettings:    customtypes.NewIndexSettingsNull(),
			wantMappingsNull: true,
			wantSettingsNull: true,
		},
		{
			// When the API actually returns non-empty content, the prior is
			// ignored and the API value wins.
			name:             "API non-empty + prior empty-object ignored",
			mappings:         map[string]any{"properties": map[string]any{"f": map[string]any{"type": "keyword"}}},
			settings:         nil,
			priorMappings:    esindex.NewMappingsValue(`{}`),
			priorSettings:    customtypes.NewIndexSettingsNull(),
			wantMappingsNull: false,
			wantSettingsNull: true,
			wantMappings:     `{"properties":{"f":{"type":"keyword"}}}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tpl := &models.Template{
				Mappings: tc.mappings,
				Settings: tc.settings,
			}
			obj, diags := flattenTemplateBlock(
				context.Background(),
				tpl,
				nil,
				tc.priorMappings,
				tc.priorSettings,
			)
			require.False(t, diags.HasError(), "unexpected diags: %v", diags)
			require.False(t, obj.IsNull(), "template object should not be null when template != nil")

			attrs := obj.Attributes()
			mappings := attrs["mappings"].(esindex.MappingsValue)
			settings := attrs["settings"].(customtypes.IndexSettingsValue)

			require.Equal(t, tc.wantMappingsNull, mappings.IsNull(), "template.mappings null mismatch")
			require.Equal(t, tc.wantSettingsNull, settings.IsNull(), "template.settings null mismatch")
			if !tc.wantMappingsNull && tc.wantMappings != "" {
				require.JSONEq(t, tc.wantMappings, mappings.ValueString())
			}
			if !tc.wantSettingsNull && tc.wantSettings != "" {
				require.JSONEq(t, tc.wantSettings, settings.ValueString())
			}
		})
	}
}
