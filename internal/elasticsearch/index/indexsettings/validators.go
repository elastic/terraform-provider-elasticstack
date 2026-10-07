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
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	indexparent "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// settingsJSONValidator validates `settings_json` at plan time. The value must
// be a non-empty flat JSON object whose top-level keys are complete settings
// paths: creation-time-only (static) keys, nested object values, explicit null
// values, the empty object and object or null array elements are all rejected.
// Values are scalars or arrays of scalars (including empty arrays). Keys outside
// indexparent.AllSettingsKeys are permitted so new Elasticsearch dynamic
// settings work before they are modeled as typed attributes.
type settingsJSONValidator struct{}

func (settingsJSONValidator) Description(context.Context) string {
	return "Validates that settings_json is a non-empty flat JSON object of dynamic index settings"
}

func (v settingsJSONValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (settingsJSONValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var settings map[string]json.RawMessage
	if err := json.Unmarshal([]byte(req.ConfigValue.ValueString()), &settings); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid settings_json",
			"`settings_json` must be a valid JSON object string.",
		)
		return
	}

	if len(settings) == 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid settings_json",
			"`settings_json` must declare at least one setting; an empty object declares no settings.",
		)
		return
	}

	if detail := duplicateCanonicalKeyDetail(settings); detail != "" {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid settings_json", detail)
		return
	}

	for key, raw := range settings {
		if err := validateSettingsJSONEntry(key, raw); err != nil {
			resp.Diagnostics.AddAttributeError(req.Path, "Invalid settings_json", *err)
			return
		}
	}
}

// duplicateCanonicalKeyDetail returns the deterministic error detail to report
// when two settings_json keys canonicalize to the same setting (for example
// `number_of_replicas` and `index.number_of_replicas`), or "" when every
// canonical key is unique. Elasticsearch stores a single value per setting,
// so two spellings of the same key could never converge.
func duplicateCanonicalKeyDetail(settings map[string]json.RawMessage) string {
	keys := make([]string, 0, len(settings))
	for key := range settings {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	seen := map[string]string{}
	for _, key := range keys {
		canonical := normalizeSettingsKey(key)
		if previous, ok := seen[canonical]; ok {
			return "`settings_json` keys " + previous + " and " + key +
				" declare the same setting twice with different spellings; a setting may be declared in only one place."
		}
		seen[canonical] = key
	}
	return ""
}

// validateSettingsJSONEntry validates a single settings_json top-level key and
// raw value, returning the error detail to report or nil when the entry is valid.
func validateSettingsJSONEntry(key string, raw json.RawMessage) *string {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "null" {
		detail := "`settings_json` value for key " + key + " is null. " +
			"Explicit `null` values are not allowed; a setting is reset by omitting it from `settings_json`."
		return &detail
	}

	var nested map[string]json.RawMessage
	if err := json.Unmarshal(raw, &nested); err == nil && nested != nil {
		detail := "`settings_json` must use flat dotted setting keys (e.g. \"index.number_of_replicas\"), " +
			"not nested objects. Nested key: " + key + "."
		return &detail
	}

	var array []json.RawMessage
	if err := json.Unmarshal(raw, &array); err == nil && array != nil {
		for _, element := range array {
			if !scalarSettingsJSONValue(element) {
				detail := "`settings_json` array values may only contain scalar elements (string, number, or boolean); " +
					"object or null elements are not allowed. Invalid element in key: " + key + "."
				return &detail
			}
		}
	}

	if settingKeyIsStatic(key) {
		detail := "Setting key " + key + " can only be set at index creation time; " +
			"this resource updates settings on an existing index and cannot change it."
		return &detail
	}

	return nil
}

// scalarSettingsJSONValue reports whether a raw settings_json value is a
// scalar (string, number, or boolean); null values are not scalars.
func scalarSettingsJSONValue(raw json.RawMessage) bool {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return false
	}
	switch value.(type) {
	case string, float64, bool:
		return true
	default:
		return false
	}
}

// validateDeclaredSettings returns plan-time diagnostics when a settings_json
// key overlaps a configured typed dynamic-setting attribute (after canonicalizing
// key spellings with or without the `index.` prefix). An index-only configuration
// (no typed attribute, no settings_json) is valid, so no minimum number of declared
// settings is enforced; values unknown at plan time defer their checks.
func validateDeclaredSettings(model tfModel) diag.Diagnostics {
	var diags diag.Diagnostics

	declaredTyped := map[string]bool{}
	for attribute, settingKey := range dynamicSettingKeyByAttribute {
		value, found := indexparent.GetFieldValueByTagValue(reflect.ValueOf(model), reflect.TypeFor[tfModel](), attribute)
		if !found || !typeutils.IsKnown(value) {
			continue
		}
		declaredTyped[settingKey] = true
	}

	var settingsJSONKeys map[string]bool
	if typeutils.IsKnown(model.SettingsJSON) {
		var settings map[string]json.RawMessage
		if err := json.Unmarshal([]byte(model.SettingsJSON.ValueString()), &settings); err != nil {
			diags.AddError("Invalid settings_json", "`settings_json` must be a valid JSON object string.")
			return diags
		}
		settingsJSONKeys = make(map[string]bool, len(settings))
		for key := range settings {
			settingsJSONKeys[normalizeSettingsKey(key)] = true
		}
	}

	for key := range settingsJSONKeys {
		if declaredTyped[key] {
			diags.AddError(
				"Conflicting index settings declaration",
				fmt.Sprintf("Setting key %q is declared both via a typed dynamic-setting attribute and via `settings_json`; declare it in one place only.", key),
			)
		}
	}

	return diags
}

var _ validator.String = settingsJSONValidator{}
