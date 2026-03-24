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

package filter

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTFModel_toAPICreateModel(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		model       TFModel
		expectedAPI *CreateAPIModel
	}{
		{
			name: "with description and items",
			model: TFModel{
				Description: types.StringValue("safe domains"),
				Items:       mustStringSet(ctx, t, []string{"*.example.com", "trusted.org"}),
			},
			expectedAPI: &CreateAPIModel{
				Description: "safe domains",
				Items:       []string{"*.example.com", "trusted.org"},
			},
		},
		{
			name: "with description only",
			model: TFModel{
				Description: types.StringValue("empty filter"),
				Items:       types.SetNull(types.StringType),
			},
			expectedAPI: &CreateAPIModel{
				Description: "empty filter",
				Items:       nil,
			},
		},
		{
			name: "null description and null items",
			model: TFModel{
				Description: types.StringNull(),
				Items:       types.SetNull(types.StringType),
			},
			expectedAPI: &CreateAPIModel{
				Description: "",
				Items:       nil,
			},
		},
		{
			name: "empty items set",
			model: TFModel{
				Description: types.StringValue("no items"),
				Items:       mustStringSet(ctx, t, []string{}),
			},
			expectedAPI: &CreateAPIModel{
				Description: "no items",
				Items:       []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, diags := tt.model.toAPICreateModel(ctx)
			require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
			assert.Equal(t, tt.expectedAPI.Description, result.Description)
			assert.Equal(t, tt.expectedAPI.Items, result.Items)
		})
	}
}

func TestTFModel_fromAPIModel(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name             string
		initialItems     types.Set
		apiModel         *APIModel
		expectedFilterID string
		expectedDesc     types.String
		expectItemsNull  bool
		expectedItems    []string
	}{
		{
			name:         "full API response",
			initialItems: mustStringSet(ctx, t, []string{"old-item"}),
			apiModel: &APIModel{
				FilterID:    "my-filter",
				Description: "A safe domains filter",
				Items:       []string{"*.example.com", "trusted.org"},
			},
			expectedFilterID: "my-filter",
			expectedDesc:     types.StringValue("A safe domains filter"),
			expectedItems:    []string{"*.example.com", "trusted.org"},
		},
		{
			name:         "empty description from API becomes null",
			initialItems: types.SetNull(types.StringType),
			apiModel: &APIModel{
				FilterID:    "my-filter",
				Description: "",
				Items:       []string{},
			},
			expectedFilterID: "my-filter",
			expectedDesc:     types.StringNull(),
			expectItemsNull:  true,
		},
		{
			name:         "empty items with non-null TF state becomes empty set",
			initialItems: mustStringSet(ctx, t, []string{}),
			apiModel: &APIModel{
				FilterID: "my-filter",
				Items:    []string{},
			},
			expectedFilterID: "my-filter",
			expectedDesc:     types.StringNull(),
			expectedItems:    []string{},
		},
		{
			name:         "empty items with null TF state stays null",
			initialItems: types.SetNull(types.StringType),
			apiModel: &APIModel{
				FilterID: "my-filter",
				Items:    []string{},
			},
			expectedFilterID: "my-filter",
			expectedDesc:     types.StringNull(),
			expectItemsNull:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &TFModel{
				Items: tt.initialItems,
			}

			diags := model.fromAPIModel(ctx, tt.apiModel)
			require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

			assert.Equal(t, tt.expectedFilterID, model.FilterID.ValueString())
			assert.Equal(t, tt.expectedDesc, model.Description)

			if tt.expectItemsNull {
				assert.True(t, model.Items.IsNull(), "expected Items to be null")
			} else {
				var items []string
				diags = model.Items.ElementsAs(ctx, &items, false)
				require.False(t, diags.HasError())
				assert.ElementsMatch(t, tt.expectedItems, items)
			}
		})
	}
}

func mustStringSet(ctx context.Context, t *testing.T, vals []string) types.Set {
	t.Helper()
	s, diags := types.SetValueFrom(ctx, types.StringType, vals)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	return s
}
