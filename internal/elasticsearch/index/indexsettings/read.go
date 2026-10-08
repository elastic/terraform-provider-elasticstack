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
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"

	esttypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	indexparent "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// readIndexSettings populates only the typed dynamic-setting attributes and
// settings_json keys present in the previously stored state (the declared
// subset); settings Elasticsearch reports that are not part of the declared
// subset are ignored. A tracked setting the API no longer reports is nulled /
// dropped so the drift shows.
func readIndexSettings(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	resourceID string,
	state tfModel,
) (tfModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	indexName := resourceID

	apiIndex, getDiags := elasticsearch.GetIndex(ctx, client, indexName)
	diags.Append(getDiags...)
	if diags.HasError() {
		return state, false, diags
	}
	if apiIndex == nil {
		return state, false, diags
	}

	state.Index = types.StringValue(indexName)

	flat, flatDiags := flatSettingsByCanonicalKey(apiIndex.Settings)
	diags.Append(flatDiags...)
	if diags.HasError() {
		return state, false, diags
	}

	readDeclaredTypedSettings(ctx, &state, flat)
	diags.Append(readDeclaredSettingsJSON(&state, flat)...)
	if diags.HasError() {
		return state, false, diags
	}

	return state, true, diags
}

// postReadIndexSettings runs after [readIndexSettings] on every read path. It
// performs the one-shot import hydration: when the import private-state flag
// set by ImportState is present, it clears the flag and populates the known
// dynamic typed attributes from a fresh API response, leaving settings_json
// unset so static and metadata settings are never adopted into state.
// Hydration is keyed on the flag and never on an empty tracked-settings set,
// which is also reachable via a legitimate configuration after an external
// reset of the last tracked setting.
func postReadIndexSettings(ctx context.Context, req entitycore.ElasticsearchPostReadRequest[tfModel]) (tfModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !importHydrationRequested(ctx, req.Private) {
		return req.State, diags
	}
	diags.Append(clearImportHydrationFlag(ctx, req.Private)...)
	if diags.HasError() {
		return req.State, diags
	}

	apiIndex, getDiags := elasticsearch.GetIndex(ctx, req.Client, req.State.Index.ValueString())
	diags.Append(getDiags...)
	if diags.HasError() {
		return req.State, diags
	}
	if apiIndex == nil {
		return req.State, diags
	}

	flat, flatDiags := flatSettingsByCanonicalKey(apiIndex.Settings)
	diags.Append(flatDiags...)
	if diags.HasError() {
		return req.State, diags
	}

	hydrateDynamicTypedSettings(ctx, &req.State, flat)

	return req.State, diags
}

// flatSettingsByCanonicalKey marshals the typed client's settings back to the
// flat form GetIndex requested (index.<key>: "<value>") and reindexes the keys
// by canonical (prefix-stripped) settings key.
func flatSettingsByCanonicalKey(settings *esttypes.IndexSettings) (map[string]json.RawMessage, diag.Diagnostics) {
	if settings == nil {
		return map[string]json.RawMessage{}, nil
	}

	settingsBytes, err := json.Marshal(settings)
	if err != nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic("failed to marshal index settings", err.Error()),
		}
	}

	var rawFlat map[string]json.RawMessage
	if err := json.Unmarshal(settingsBytes, &rawFlat); err != nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic("failed to unmarshal index settings", err.Error()),
		}
	}

	flat := make(map[string]json.RawMessage, len(rawFlat))
	for key, raw := range rawFlat {
		flat[normalizeSettingsKey(key)] = raw
	}
	return flat, nil
}

// readDeclaredTypedSettings refreshes the typed dynamic-setting attributes
// tracked in state (non-null values) from the flat API settings. A tracked
// setting the API no longer reports is set to null so the drift shows.
func readDeclaredTypedSettings(ctx context.Context, state *tfModel, flat map[string]json.RawMessage) {
	modelType := reflect.TypeFor[tfModel]()
	stateValue := reflect.ValueOf(state).Elem()

	for attribute, settingKey := range dynamicSettingKeyByAttribute {
		current, found := indexparent.GetFieldValueByTagValue(stateValue, modelType, attribute)
		if !found || !typeutils.IsKnown(current) {
			continue
		}

		raw, ok := flat[settingKey]
		if !ok {
			indexparent.SetFieldValueByTagValue(reflect.ValueOf(state), modelType, attribute, nullTypedSettingValue(current))
			continue
		}

		if value, ok := flatSettingAsTypedValue(ctx, current, raw); ok {
			indexparent.SetFieldValueByTagValue(reflect.ValueOf(state), modelType, attribute, value)
		}
	}
}

// hydrateDynamicTypedSettings populates every known dynamic typed attribute
// from the flat API settings, regardless of previous state. Used only on the
// first read after terraform import, so a narrowed configuration shows a diff
// instead of silently adopting settings.
func hydrateDynamicTypedSettings(ctx context.Context, state *tfModel, flat map[string]json.RawMessage) {
	modelType := reflect.TypeFor[tfModel]()
	stateValue := reflect.ValueOf(state).Elem()

	for attribute, settingKey := range dynamicSettingKeyByAttribute {
		raw, ok := flat[settingKey]
		if !ok {
			continue
		}

		current, found := indexparent.GetFieldValueByTagValue(stateValue, modelType, attribute)
		if !found {
			continue
		}

		if value, ok := flatSettingAsTypedValue(ctx, current, raw); ok {
			indexparent.SetFieldValueByTagValue(reflect.ValueOf(state), modelType, attribute, value)
		}
	}
}

// readDeclaredSettingsJSON reconciles the settings_json keys tracked in state
// against the flat API settings: each API value is converted back to the JSON
// scalar type declared in state (GetIndex returns string values because it
// requests flat settings), and keys the API no longer reports are dropped so
// the drift shows.
func readDeclaredSettingsJSON(state *tfModel, flat map[string]json.RawMessage) diag.Diagnostics {
	if !typeutils.IsKnown(state.SettingsJSON) {
		return nil
	}

	stateSettings, diags := typeutils.UnmarshalJSONDiag[map[string]json.RawMessage](state.SettingsJSON.ValueString(), "failed to unmarshal settings_json from state")
	if diags.HasError() {
		return diags
	}

	reconciled := make(map[string]json.RawMessage, len(stateSettings))
	for key, stateRaw := range stateSettings {
		apiRaw, ok := flat[normalizeSettingsKey(key)]
		if !ok {
			continue
		}

		value, ok := reconcileSettingJSONValue(stateRaw, apiRaw)
		if !ok {
			continue
		}
		reconciled[key] = value
	}

	encoded, err := json.Marshal(reconciled)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic("failed to marshal settings_json", err.Error()),
		}
	}

	state.SettingsJSON = jsontypes.NewNormalizedValue(string(encoded))
	return nil
}

// reconcileSettingJSONValue converts an API flat-settings value (typically a
// JSON string like "20000" or "false") to the JSON scalar type declared in
// state, so an unchanged settings_json key does not produce false drift.
// Numbers are handled as exact JSON tokens (json.Number), never parsed
// through float64, so integers beyond its 2^53 precision round-trip verbatim.
func reconcileSettingJSONValue(stateRaw, apiRaw json.RawMessage) (json.RawMessage, bool) {
	declared := decodeJSONValueWithExactNumbers(stateRaw)
	if declared == nil {
		return nil, false
	}

	switch declared := declared.(type) {
	case json.Number:
		var s string
		if err := json.Unmarshal(apiRaw, &s); err == nil {
			// Echo the API string's numeric token verbatim when it is a valid
			// JSON number, so no precision is lost re-encoding it.
			if err := json.Unmarshal([]byte(s), new(json.Number)); err == nil {
				return json.RawMessage(s), true
			}
			return nil, false
		}
		var n json.Number
		if err := json.Unmarshal(apiRaw, &n); err == nil {
			return apiRaw, true
		}
		return nil, false
	case bool:
		var s string
		if err := json.Unmarshal(apiRaw, &s); err == nil {
			if b, err := strconv.ParseBool(s); err == nil {
				if encoded, err := json.Marshal(b); err == nil {
					return encoded, true
				}
			}
			return nil, false
		}
		var b bool
		if err := json.Unmarshal(apiRaw, &b); err == nil {
			return apiRaw, true
		}
		return nil, false
	case string:
		var s string
		if err := json.Unmarshal(apiRaw, &s); err == nil {
			return apiRaw, true
		}
		return nil, false
	case []any:
		return reconcileArraySettingJSONValue(declared, apiRaw)
	default:
		return nil, false
	}
}

// reconcileArraySettingJSONValue reconciles an API flat-settings array value
// (a real JSON array or a JSON-encoded array string) element-wise against the
// element types declared in state, preserving element order. When the API
// array length differs from the declared one the API array is returned as-is
// so the drift shows instead of the tracked array being dropped.
func reconcileArraySettingJSONValue(declared []any, apiRaw json.RawMessage) (json.RawMessage, bool) {
	elements, ok := arrayElementsFromFlatRaw(apiRaw)
	if !ok {
		return nil, false
	}

	if len(elements) != len(declared) {
		encoded, err := json.Marshal(elements)
		if err != nil {
			return nil, false
		}
		return encoded, true
	}

	reconciled := make([]json.RawMessage, len(elements))
	for i, apiElement := range elements {
		declaredRaw, err := json.Marshal(declared[i])
		if err != nil {
			return nil, false
		}
		value, ok := reconcileSettingJSONValue(declaredRaw, apiElement)
		if !ok {
			return nil, false
		}
		reconciled[i] = value
	}

	encoded, err := json.Marshal(reconciled)
	if err != nil {
		return nil, false
	}
	return encoded, true
}

// decodeJSONValueWithExactNumbers decodes a raw JSON value with json.Number
// for numbers, so large integers keep their exact token; nil when invalid.
func decodeJSONValueWithExactNumbers(raw json.RawMessage) any {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil
	}
	return value
}

// arrayElementsFromFlatRaw extracts the elements of a flat API settings array
// value that arrives either as a real JSON array or as a JSON-encoded array
// string. ok is false when the value is neither array form.
func arrayElementsFromFlatRaw(raw json.RawMessage) ([]json.RawMessage, bool) {
	var elements []json.RawMessage
	if err := json.Unmarshal(raw, &elements); err == nil && elements != nil {
		return elements, true
	}

	var encoded string
	if err := json.Unmarshal(raw, &encoded); err != nil {
		return nil, false
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(encoded)), &elements); err != nil || elements == nil {
		return nil, false
	}
	return elements, true
}

// flatSettingAsTypedValue converts a flat API settings value to the declared
// attribute type, mirroring index/index/settings_read.go's conversion of flat
// string values to typed attributes. ok is false when the value cannot be
// represented in the declared type.
func flatSettingAsTypedValue(ctx context.Context, current attr.Value, raw json.RawMessage) (attr.Value, bool) {
	switch current.(type) {
	case types.Int64:
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return types.Int64Null(), false
		}
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return types.Int64Null(), false
		}
		return types.Int64Value(i), true
	case types.Bool:
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return types.BoolNull(), false
		}
		b, err := strconv.ParseBool(s)
		if err != nil {
			return types.BoolNull(), false
		}
		return types.BoolValue(b), true
	case types.String:
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return types.StringNull(), false
		}
		return types.StringValue(s), true
	case types.Set:
		elems, ok := stringElemsFromFlatRaw(raw)
		if !ok {
			return types.SetNull(types.StringType), false
		}
		set, setDiags := types.SetValueFrom(ctx, types.StringType, elems)
		if setDiags.HasError() {
			return types.SetNull(types.StringType), false
		}
		return set, true
	default:
		return nil, false
	}
}

// stringElemsFromFlatRaw extracts string elements from a flat settings value
// that may be a scalar string or a JSON array of strings (for example
// index.query.default_field). ok distinguishes a value that cannot be
// represented as string elements from a valid empty array, which is a legal
// tracked value and must round-trip as an empty set.
func stringElemsFromFlatRaw(raw json.RawMessage) ([]string, bool) {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, false
	}

	switch value := v.(type) {
	case string:
		// A string value may itself be a JSON-encoded string array (for
		// example index.query.default_field), mirroring the index resource's
		// stringSliceFromAny decoding.
		trimmed := strings.TrimSpace(value)
		if strings.HasPrefix(trimmed, "[") {
			var arr []string
			if err := json.Unmarshal([]byte(trimmed), &arr); err == nil {
				return arr, true
			}
		}
		return []string{value}, true
	case []any:
		elems := make([]string, 0, len(value))
		for _, elem := range value {
			s, ok := elem.(string)
			if !ok {
				return nil, false
			}
			elems = append(elems, s)
		}
		return elems, true
	default:
		return nil, false
	}
}

// nullTypedSettingValue returns the null value matching the declared type of
// the current attribute value.
func nullTypedSettingValue(current attr.Value) attr.Value {
	switch current.(type) {
	case types.Int64:
		return types.Int64Null()
	case types.Bool:
		return types.BoolNull()
	case types.String:
		return types.StringNull()
	case types.Set:
		return types.SetNull(types.StringType)
	default:
		return current
	}
}
