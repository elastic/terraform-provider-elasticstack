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

package output

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigYamlFromAPI covers the empty-string-to-null normalization that
// keeps state consistent when the Fleet API echoes back an empty
// `config_yaml` value for outputs the user never configured (issue #1856).
func TestConfigYamlFromAPI(t *testing.T) {
	t.Parallel()

	empty := ""
	value := "bulk_max_size: 100\n"

	tests := []struct {
		name     string
		input    *string
		wantNull bool
		wantVal  string
	}{
		{name: "nil pointer maps to null", input: nil, wantNull: true},
		{name: "empty string maps to null", input: &empty, wantNull: true},
		{name: "non-empty string maps to value", input: &value, wantNull: false, wantVal: value},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := configYamlFromAPI(tc.input)
			if tc.wantNull {
				assert.True(t, got.IsNull(), "expected null, got %v", got)
				return
			}
			require.False(t, got.IsNull(), "expected non-null value")
			assert.Equal(t, tc.wantVal, got.ValueString())
		})
	}
}

// TestFromAPICommonFields_ConfigYamlNormalization asserts that the shared
// reader normalizes nil and empty config_yaml from the Fleet API into a
// null state value. When the model already carries a null config_yaml
// (user removed it or initial import), the API echo of a previously
// stored value SHALL be suppressed so plan-null intent survives apply and
// the subsequent refresh — fixing the "inconsistent values for sensitive
// attribute" failure mode and avoiding perpetual drift (issue #1856).
func TestFromAPICommonFields_ConfigYamlNormalization(t *testing.T) {
	t.Parallel()

	hosts := []string{"https://example:9200"}
	empty := ""
	value := "bulk_max_size: 100\n"
	other := "queue.mem.events: 4096\n"

	// existingName mirrors what the framework gives us in Read after a
	// successful prior apply: Name is always populated. On import only the
	// SpaceImporter-populated fields are set, so Name is null.
	existingName := types.StringValue("example")
	importName := types.StringNull()

	tests := []struct {
		name         string
		api          *string
		existing     customtypes.NormalizedYamlValue
		existingName types.String
		wantNull     bool
		wantValue    string
	}{
		{
			name:         "nil API value with null existing state",
			api:          nil,
			existing:     customtypes.NewNormalizedYamlNull(),
			existingName: existingName,
			wantNull:     true,
		},
		{
			name:         "empty API value with null existing state",
			api:          &empty,
			existing:     customtypes.NewNormalizedYamlNull(),
			existingName: existingName,
			wantNull:     true,
		},
		{
			name: "non-empty API echo is suppressed when existing state is null (user removed config_yaml)",
			api:  &value,
			// Plan/state model carries null because the user removed the
			// attribute from configuration; Fleet still echoes the prior
			// value back. The reader must honour the null intent.
			existing:     customtypes.NewNormalizedYamlNull(),
			existingName: existingName,
			wantNull:     true,
		},
		{
			name:         "non-empty API value with non-null existing state surfaces the API value",
			api:          &value,
			existing:     customtypes.NewNormalizedYamlValue(value),
			existingName: existingName,
			wantNull:     false,
			wantValue:    value,
		},
		{
			name:         "non-empty API drift over a non-null existing state surfaces the drift",
			api:          &other,
			existing:     customtypes.NewNormalizedYamlValue(value),
			existingName: existingName,
			wantNull:     false,
			wantValue:    other,
		},
		{
			name: "import populates config_yaml from the API even when existing state is null",
			api:  &value,
			// Import scenario: only output_id is pre-populated, all
			// other fields (including Name) are null. The reader must
			// take the API value so the imported state matches Fleet.
			existing:     customtypes.NewNormalizedYamlNull(),
			existingName: importName,
			wantNull:     false,
			wantValue:    value,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			model := outputModel{
				ConfigYaml: tc.existing,
				Name:       tc.existingName,
			}
			diags := model.fromAPICommonFields(context.Background(), commonOutputReadData{
				name:       "example",
				outputType: "elasticsearch",
				hosts:      hosts,
				configYaml: tc.api,
			})
			require.False(t, diags.HasError(), "unexpected diags: %v", diags)

			if tc.wantNull {
				assert.True(t, model.ConfigYaml.IsNull(), "expected null, got %v", model.ConfigYaml)
				return
			}
			require.False(t, model.ConfigYaml.IsNull())
			assert.Equal(t, tc.wantValue, model.ConfigYaml.ValueString())
		})
	}
}

// TestConfigYaml_SemanticEquality locks in that the resource's config_yaml
// attribute uses NormalizedYamlValue, so semantically equivalent YAML
// (different key order, whitespace, etc.) does not register as a change.
// Without this, the API re-emitting normalized YAML would trigger spurious
// `inconsistent values for sensitive attribute` errors on every apply.
func TestConfigYaml_SemanticEquality(t *testing.T) {
	t.Parallel()

	planned := customtypes.NewNormalizedYamlValue("a: 1\nb: 2\n")
	apiEcho := customtypes.NewNormalizedYamlValue("b: 2\na: 1\n")

	equal, diags := planned.StringSemanticEquals(context.Background(), basetypes.StringValuable(apiEcho))
	require.False(t, diags.HasError(), "unexpected diags: %v", diags)
	assert.True(t, equal, "expected key-reordered YAML to be semantically equal")
}
