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

package transform

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// baseModel returns a minimal valid tfModel with required fields populated.
func baseModel() tfModel {
	return tfModel{
		Name: types.StringValue("test-transform"),
		Source: &tfModelSource{
			Indices: []types.String{types.StringValue("src-index")},
			Query:   jsontypes.NewNormalizedNull(),
		},
		Destination: &tfModelDestination{
			Index: types.StringValue("dest-index"),
		},
		Pivot: jsontypes.NewNormalizedValue(`{"group_by":{"customer_id":{"terms":{"field":"customer_id"}}}}`),
	}
}

func TestToAPIModel_VersionGating_DestinationAliases(t *testing.T) {
	t.Parallel()

	model := baseModel()
	model.Destination.Aliases = []tfModelAlias{
		{Alias: types.StringValue("alias-1"), MoveOnCreation: types.BoolValue(true)},
	}

	cases := []struct {
		name      string
		version   string
		wantAlias bool
	}{
		{"omitted below 8.8", "8.7.0", false},
		{"included at 8.8", "8.8.0", true},
		{"included above 8.8", "8.9.0", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := version.Must(version.NewVersion(tc.version))
			api, diags := toAPIModel(context.Background(), model, v)
			require.False(t, diags.HasError(), "%s", diags)
			require.NotNil(t, api.Destination)
			if tc.wantAlias {
				assert.Len(t, api.Destination.Aliases, 1)
				assert.Equal(t, "alias-1", api.Destination.Aliases[0].Alias)
			} else {
				assert.Empty(t, api.Destination.Aliases)
			}
		})
	}
}

func TestToAPIModel_VersionGating_Settings(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		version      string
		settingField string
		wantPresent  bool
		setupModel   func(*tfModel)
		assertField  func(t *testing.T, s *models.TransformSettings)
	}{
		{
			name:         "deduce_mappings omitted below 8.1",
			version:      "8.0.0",
			settingField: "deduce_mappings",
			wantPresent:  false,
			setupModel: func(m *tfModel) {
				m.DeduceMappings = types.BoolValue(true)
			},
			assertField: func(t *testing.T, s *models.TransformSettings) {
				assert.Nil(t, s.DeduceMappings)
			},
		},
		{
			name:         "deduce_mappings included at 8.1",
			version:      "8.1.0",
			settingField: "deduce_mappings",
			wantPresent:  true,
			setupModel: func(m *tfModel) {
				m.DeduceMappings = types.BoolValue(true)
			},
			assertField: func(t *testing.T, s *models.TransformSettings) {
				require.NotNil(t, s.DeduceMappings)
				assert.True(t, *s.DeduceMappings)
			},
		},
		{
			name:         "num_failure_retries omitted below 8.4",
			version:      "8.3.0",
			settingField: "num_failure_retries",
			wantPresent:  false,
			setupModel: func(m *tfModel) {
				m.NumFailureRetries = types.Int64Value(5)
			},
			assertField: func(t *testing.T, s *models.TransformSettings) {
				assert.Nil(t, s.NumFailureRetries)
			},
		},
		{
			name:         "num_failure_retries included at 8.4",
			version:      "8.4.0",
			settingField: "num_failure_retries",
			wantPresent:  true,
			setupModel: func(m *tfModel) {
				m.NumFailureRetries = types.Int64Value(5)
			},
			assertField: func(t *testing.T, s *models.TransformSettings) {
				require.NotNil(t, s.NumFailureRetries)
				assert.Equal(t, 5, *s.NumFailureRetries)
			},
		},
		{
			name:         "unattended omitted below 8.5",
			version:      "8.4.0",
			settingField: "unattended",
			wantPresent:  false,
			setupModel: func(m *tfModel) {
				m.Unattended = types.BoolValue(true)
			},
			assertField: func(t *testing.T, s *models.TransformSettings) {
				assert.Nil(t, s.Unattended)
			},
		},
		{
			name:         "unattended included at 8.5",
			version:      "8.5.0",
			settingField: "unattended",
			wantPresent:  true,
			setupModel: func(m *tfModel) {
				m.Unattended = types.BoolValue(true)
			},
			assertField: func(t *testing.T, s *models.TransformSettings) {
				require.NotNil(t, s.Unattended)
				assert.True(t, *s.Unattended)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			model := baseModel()
			tc.setupModel(&model)

			v := version.Must(version.NewVersion(tc.version))
			api, diags := toAPIModel(context.Background(), model, v)
			require.False(t, diags.HasError(), "%s", diags)

			if tc.wantPresent {
				require.NotNil(t, api.Settings, "expected settings to be present")
				tc.assertField(t, api.Settings)
			} else if api.Settings != nil {
				// Settings may be nil or the specific field may be nil.
				tc.assertField(t, api.Settings)
			}
		})
	}
}

func TestToAPIModel_VersionGating_UnsetFieldsNoWarning(t *testing.T) {
	t.Parallel()

	// When a version-gated field is not configured, it should not be present
	// in the API model and we should not get validation diagnostics.
	model := baseModel()
	// None of the gated settings are configured.

	v := version.Must(version.NewVersion("7.17.0"))
	api, diags := toAPIModel(context.Background(), model, v)
	require.False(t, diags.HasError(), "%s", diags)
	assert.Nil(t, api.Settings)
	assert.Empty(t, api.Destination.Aliases)
}

func TestToAPIModel_JSONFields(t *testing.T) {
	t.Parallel()

	model := baseModel()
	model.Pivot = jsontypes.NewNormalizedValue(`{"group_by":{"customer_id":{"terms":{"field":"customer_id"}}}}`)
	model.Latest = jsontypes.NewNormalizedNull()

	v := version.Must(version.NewVersion("8.9.0"))
	api, diags := toAPIModel(context.Background(), model, v)
	require.False(t, diags.HasError(), "%s", diags)

	pivotBytes, err := json.Marshal(api.Pivot)
	require.NoError(t, err)
	assert.JSONEq(t, `{"group_by":{"customer_id":{"terms":{"field":"customer_id"}}}}`, string(pivotBytes))
	assert.Nil(t, api.Latest)
}
