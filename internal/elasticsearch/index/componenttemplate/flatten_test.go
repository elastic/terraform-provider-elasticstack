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

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/stretchr/testify/require"
)

// TestFlattenTemplateBlock_MappingsAndSettingsEmptyObject pins the empty-object
// normalisation behaviour added for issue #609: both a nil and an empty
// map[string]any decoded from Elasticsearch's response must produce a null
// Terraform value, so that a plan without `mappings`/`settings` blocks does
// not drift against a state populated from a `"mappings": {}`/`"settings": {}`
// API response.
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
			obj, diags := flattenTemplateBlock(context.Background(), tpl, nil)
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
