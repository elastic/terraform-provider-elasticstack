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

package indexsettings

import (
	"context"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func validateSettingsJSON(t *testing.T, raw string) (*validator.StringResponse, string) {
	t.Helper()

	resp := &validator.StringResponse{}
	settingsJSONValidator{}.ValidateString(context.Background(), validator.StringRequest{
		Path:        path.Root("settings_json"),
		ConfigValue: types.StringValue(raw),
	}, resp)

	details := make([]string, 0, len(resp.Diagnostics))
	for _, d := range resp.Diagnostics {
		details = append(details, d.Detail())
	}
	return resp, strings.Join(details, "\n")
}

func TestSettingsJSONValidator(t *testing.T) {
	tests := []struct {
		name       string
		raw        string
		wantErr    bool
		wantDetail string
	}{
		{
			name:    "flat dynamic key",
			raw:     `{"number_of_replicas": 2}`,
			wantErr: false,
		},
		{
			name:    "index-prefixed dynamic key",
			raw:     `{"index.max_result_window": 20000}`,
			wantErr: false,
		},
		{
			name:    "unknown key outside AllSettingsKeys is permitted",
			raw:     `{"some.future.dynamic.setting": "value"}`,
			wantErr: false,
		},
		{
			name:       "empty object is rejected",
			raw:        `{}`,
			wantErr:    true,
			wantDetail: "at least one setting",
		},
		{
			name:       "static key is rejected",
			raw:        `{"number_of_shards": 3}`,
			wantErr:    true,
			wantDetail: "can only be set at index creation time",
		},
		{
			name:       "index-prefixed static key is rejected",
			raw:        `{"index.number_of_shards": 3}`,
			wantErr:    true,
			wantDetail: "can only be set at index creation time",
		},
		{
			name:       "nested object value is rejected",
			raw:        `{"index": {"number_of_replicas": 2}}`,
			wantErr:    true,
			wantDetail: "flat dotted",
		},
		{
			name:       "explicit null value is rejected",
			raw:        `{"refresh_interval": null, "max_result_window": 20000}`,
			wantErr:    true,
			wantDetail: "null",
		},
		{
			name:    "scalar array is permitted",
			raw:     `{"index.query.default_field": ["title", "body"]}`,
			wantErr: false,
		},
		{
			name:    "empty array is permitted",
			raw:     `{"some.future.array.setting": []}`,
			wantErr: false,
		},
		{
			name:       "null array element is rejected",
			raw:        `{"index.query.default_field": ["title", null]}`,
			wantErr:    true,
			wantDetail: "may only contain scalar elements",
		},
		{
			name:       "object array element is rejected",
			raw:        `{"index.query.default_field": [{"field": "title"}]}`,
			wantErr:    true,
			wantDetail: "may only contain scalar elements",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, detail := validateSettingsJSON(t, tt.raw)

			if !tt.wantErr {
				assert.False(t, resp.Diagnostics.HasError(), "unexpected validation error: %s", detail)
				return
			}
			if !resp.Diagnostics.HasError() {
				t.Fatal("expected validation error, got none")
			}
			assert.Contains(t, detail, tt.wantDetail)
		})
	}
}

func TestSettingsJSONValidator_nullAndUnknownPass(t *testing.T) {
	resp := &validator.StringResponse{}
	settingsJSONValidator{}.ValidateString(context.Background(), validator.StringRequest{
		Path:        path.Root("settings_json"),
		ConfigValue: types.StringNull(),
	}, resp)
	assert.False(t, resp.Diagnostics.HasError(), "null settings_json must pass")

	resp = &validator.StringResponse{}
	settingsJSONValidator{}.ValidateString(context.Background(), validator.StringRequest{
		Path:        path.Root("settings_json"),
		ConfigValue: types.StringUnknown(),
	}, resp)
	assert.False(t, resp.Diagnostics.HasError(), "unknown settings_json must pass")
}

func TestValidateDeclaredSettings(t *testing.T) {
	tests := []struct {
		name         string
		replicas     types.Int64
		settingsJSON jsontypes.Normalized
		wantErr      bool
		wantDetail   string
	}{
		{
			name:         "typed attribute and settings_json key overlap",
			replicas:     types.Int64Value(1),
			settingsJSON: jsontypes.NewNormalizedValue(`{"number_of_replicas": 2}`),
			wantErr:      true,
			wantDetail:   "number_of_replicas",
		},
		{
			name:         "overlap after index prefix canonicalization",
			replicas:     types.Int64Value(1),
			settingsJSON: jsontypes.NewNormalizedValue(`{"index.number_of_replicas": 2}`),
			wantErr:      true,
			wantDetail:   "number_of_replicas",
		},
		{
			name:         "distinct keys do not overlap",
			replicas:     types.Int64Value(1),
			settingsJSON: jsontypes.NewNormalizedValue(`{"max_result_window": 20000}`),
			wantErr:      false,
		},
		{
			name:         "settings_json only",
			settingsJSON: jsontypes.NewNormalizedValue(`{"max_result_window": 20000}`),
			wantErr:      false,
		},
		{
			name:     "typed attribute only",
			replicas: types.Int64Value(1),
			wantErr:  false,
		},
		{
			name:    "index-only configuration is valid",
			wantErr: false,
		},
		{
			name:         "unknown settings_json value defers the overlap check",
			settingsJSON: jsontypes.NewNormalizedUnknown(),
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := tfModel{NumberOfReplicas: tt.replicas}
			switch {
			case tt.settingsJSON.IsUnknown() || tt.settingsJSON.ValueString() != "":
				model.SettingsJSON = tt.settingsJSON
			default:
				model.SettingsJSON = jsontypes.NewNormalizedNull()
			}

			diags := validateDeclaredSettings(model)

			if !tt.wantErr {
				assert.False(t, diags.HasError(), "unexpected validation error: %v", diags)
				return
			}
			if !diags.HasError() {
				t.Fatal("expected validation error, got none")
			}
			var detail string
			for _, d := range diags {
				detail += d.Detail() + "\n"
			}
			assert.Contains(t, detail, tt.wantDetail)
		})
	}
}

func TestSettingKeyIsStatic(t *testing.T) {
	assert.True(t, settingKeyIsStatic(index.SettingNumberOfShards))
	assert.True(t, settingKeyIsStatic("index."+index.SettingCodec))
	assert.False(t, settingKeyIsStatic(index.SettingNumberOfReplicas))
	assert.False(t, settingKeyIsStatic("some.unknown.key"))
}
