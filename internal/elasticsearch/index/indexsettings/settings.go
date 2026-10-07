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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	indexparent "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// indexSettingsKeyPrefix is the prefix Elasticsearch puts on index settings keys
// in flat settings responses (e.g. `index.number_of_replicas`). Setting keys
// may be written with or without the prefix in `settings_json`.
const indexSettingsKeyPrefix = "index."

// normalizeSettingsKey canonicalizes a flat settings key by stripping the
// optional `index.` prefix, so `settings_json` keys, typed-attribute settings
// keys and GetIndex flat keys compare equal regardless of spelling. Shared by
// the overlap validator, update diffing and read.
func normalizeSettingsKey(key string) string {
	return strings.TrimPrefix(key, indexSettingsKeyPrefix)
}

// staticSettingsKeys is the set of canonical (normalized) creation-time-only
// setting keys.
var staticSettingsKeys = func() map[string]struct{} {
	keys := make(map[string]struct{}, len(indexparent.StaticSettingsKeys))
	for _, key := range indexparent.StaticSettingsKeys {
		keys[key] = struct{}{}
	}
	return keys
}()

// settingKeyIsStatic reports whether the given key (with or without the
// `index.` prefix) is a creation-time-only setting that this resource cannot
// update on an existing index.
func settingKeyIsStatic(key string) bool {
	_, ok := staticSettingsKeys[normalizeSettingsKey(key)]
	return ok
}

// dynamicSettingKeyByAttribute maps the Terraform attribute name of every
// shared dynamic-setting attribute to its canonical Elasticsearch settings key.
var dynamicSettingKeyByAttribute = func() map[string]string {
	m := make(map[string]string, len(indexparent.DynamicSettingsKeys))
	for _, key := range indexparent.DynamicSettingsKeys {
		m[typeutils.ConvertSettingsKeyToTFFieldKey(key)] = key
	}
	return m
}()

// declaredSettingsPayload returns the flat settings map declared by the model:
// every configured typed dynamic-setting attribute plus every settings_json
// entry. Typed attributes use the full flat key spelling (index.<canonical
// key>); settings_json keys are sent as declared.
func declaredSettingsPayload(model tfModel) (map[string]any, diag.Diagnostics) {
	payload := map[string]any{}

	for attribute, settingKey := range dynamicSettingKeyByAttribute {
		value, found := indexparent.GetFieldValueByTagValue(reflect.ValueOf(model), reflect.TypeFor[tfModel](), attribute)
		if !found || !typeutils.IsKnown(value) {
			continue
		}
		payload[indexSettingsKeyPrefix+settingKey] = typedSettingValue(value)
	}

	if typeutils.IsKnown(model.SettingsJSON) {
		var settings map[string]json.RawMessage
		if err := json.Unmarshal([]byte(model.SettingsJSON.ValueString()), &settings); err != nil {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic("Invalid settings_json", "`settings_json` must be a valid JSON object string."),
			}
		}
		for key, raw := range settings {
			var value any
			if err := json.Unmarshal(raw, &value); err != nil {
				return nil, diag.Diagnostics{
					diag.NewErrorDiagnostic("Invalid settings_json", fmt.Sprintf("invalid value for setting %q.", key)),
				}
			}
			payload[key] = value
		}
	}

	return payload, nil
}

// typedSettingValue converts a known typed dynamic-setting attribute value to
// its Elasticsearch settings payload representation.
func typedSettingValue(value attr.Value) any {
	switch a := value.(type) {
	case types.String:
		return a.ValueString()
	case types.Bool:
		return a.ValueBool()
	case types.Int64:
		return a.ValueInt64()
	case types.Set:
		elems := []string{}
		for _, elem := range a.Elements() {
			elems = append(elems, elem.(types.String).ValueString())
		}
		return elems
	default:
		return nil
	}
}
