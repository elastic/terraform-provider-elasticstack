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

package index

import (
	"context"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"

	indexparent "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// importHydrationPrunableFieldKeys lists tfModel tfsdk tags for Optional-only
// settings fields populated by hydrateAllSettingsFromRaw and cleared in ModifyPlan
// when absent from configuration.
var importHydrationPrunableFieldKeys []string

func init() {
	for _, key := range indexparent.AllSettingsKeys {
		if sortKeysExpandedFromNestedBlock[key] {
			continue
		}
		importHydrationPrunableFieldKeys = append(importHydrationPrunableFieldKeys, typeutils.ConvertSettingsKeyToTFFieldKey(key))
	}
	importHydrationPrunableFieldKeys = append(importHydrationPrunableFieldKeys,
		"analysis_analyzer",
		"analysis_tokenizer",
		"analysis_char_filter",
		"analysis_filter",
		"analysis_normalizer",
	)
}

// hydrateAllSettingsFromRaw parses settings_raw and populates individual setting
// fields on the model. Conversion failures for individual keys are skipped.
func hydrateAllSettingsFromRaw(ctx context.Context, model *tfModel) diag.Diagnostics {
	if !typeutils.IsKnown(model.SettingsRaw) {
		return nil
	}

	var flat map[string]json.RawMessage
	if err := json.Unmarshal([]byte(model.SettingsRaw.ValueString()), &flat); err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic("failed to unmarshal settings_raw for import hydration", err.Error()),
		}
	}

	hydrateAnalysisFromFlatSettings(model, flat)

	if raw, ok := flat["index."+indexparent.SettingQueryDefaultField]; ok {
		hydrateQueryDefaultFieldFromRaw(ctx, model, raw)
	}

	modelType := reflect.TypeFor[tfModel]()
	for _, key := range indexparent.AllSettingsKeys {
		if sortKeysSkippedOnImportHydration[key] {
			continue
		}
		if key == indexparent.SettingQueryDefaultField {
			continue
		}

		raw, ok := flat["index."+key]
		if !ok {
			continue
		}

		tfFieldKey := typeutils.ConvertSettingsKeyToTFFieldKey(key)
		setFlatSettingOnModel(ctx, model, tfFieldKey, raw, modelType)
	}

	return nil
}

const indexAnalysisFlatPrefix = "index.analysis."

// hydrateAnalysisFromFlatSettings rebuilds analysis_* model fields from flat
// settings_raw keys (index.analysis.<category>.<name>.<property>) returned by
// FlatSettings(true). A nested index.analysis object is also accepted when present.
func hydrateAnalysisFromFlatSettings(model *tfModel, flat map[string]json.RawMessage) {
	if raw, ok := flat["index.analysis"]; ok {
		hydrateAnalysisFromNestedObject(model, raw)
		return
	}

	byCategory := make(map[string]map[string]map[string]any)
	for key, raw := range flat {
		if !strings.HasPrefix(key, indexAnalysisFlatPrefix) {
			continue
		}
		suffix := strings.TrimPrefix(key, indexAnalysisFlatPrefix)
		parts := strings.Split(suffix, ".")
		if len(parts) < 3 {
			continue
		}
		category, name := parts[0], parts[1]
		propParts := parts[2:]

		value := parseFlatSettingJSONValue(coerceFlatScalar(raw))
		if value == nil {
			continue
		}
		if byCategory[category] == nil {
			byCategory[category] = make(map[string]map[string]any)
		}
		if byCategory[category][name] == nil {
			byCategory[category][name] = make(map[string]any)
		}
		setNestedMapValue(byCategory[category][name], propParts, value)
	}

	for category, target := range analysisNormalizedFieldTargets(model) {
		names, ok := byCategory[category]
		if !ok || len(names) == 0 {
			continue
		}
		bytes, err := json.Marshal(names)
		if err != nil {
			continue
		}
		*target = jsontypes.NewNormalizedValue(string(bytes))
	}
}

func hydrateAnalysisFromNestedObject(model *tfModel, raw json.RawMessage) {
	var analysis map[string]json.RawMessage
	if err := json.Unmarshal(raw, &analysis); err != nil {
		return
	}

	for subKey, target := range analysisNormalizedFieldTargets(model) {
		subRaw, ok := analysis[subKey]
		if !ok {
			continue
		}
		if normalized, ok := jsonRawToNormalized(subRaw); ok {
			*target = normalized
		}
	}
}

// coerceFlatScalar normalizes flat-settings JSON string scalars ("3", "true") to
// numeric/boolean JSON for analysis hydration. Non-string raw JSON passes through.
func coerceFlatScalar(raw json.RawMessage) json.RawMessage {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return raw
	}
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		if coerced, err := json.Marshal(json.Number(s)); err == nil {
			return coerced
		}
	}
	if b, err := strconv.ParseBool(s); err == nil {
		if coerced, err := json.Marshal(b); err == nil {
			return coerced
		}
	}
	return raw
}

func parseFlatSettingJSONValue(raw json.RawMessage) any {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil
	}
	return value
}

func setNestedMapValue(obj map[string]any, path []string, value any) {
	if len(path) == 0 {
		return
	}
	cur := obj
	for i := range len(path) - 1 {
		seg := path[i]
		next, ok := cur[seg]
		if !ok {
			child := make(map[string]any)
			cur[seg] = child
			cur = child
			continue
		}
		child, ok := next.(map[string]any)
		if !ok {
			child = make(map[string]any)
			cur[seg] = child
		}
		cur = child
	}
	cur[path[len(path)-1]] = value
}

func jsonRawToNormalized(raw json.RawMessage) (jsontypes.Normalized, bool) {
	if len(raw) == 0 {
		return jsontypes.Normalized{}, false
	}
	return jsontypes.NewNormalizedValue(string(raw)), true
}

func hydrateQueryDefaultFieldFromRaw(ctx context.Context, model *tfModel, raw json.RawMessage) {
	elems := extractStringSliceFromFlatRaw(raw)
	if len(elems) == 0 {
		return
	}

	fieldSet, diags := types.SetValueFrom(ctx, types.StringType, elems)
	if diags.HasError() {
		return
	}
	model.QueryDefaultField = fieldSet
}

// extractStringSliceFromFlatRaw extracts string values from a flat-settings JSON
// value (scalar string or JSON array).
func extractStringSliceFromFlatRaw(raw json.RawMessage) []string {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil
	}
	return stringSliceFromAny(v)
}

func setFlatSettingOnModel(ctx context.Context, model *tfModel, tfFieldKey string, raw json.RawMessage, modelType reflect.Type) {
	value, ok := model.getFieldValueByTagValue(tfFieldKey, modelType)
	if !ok {
		return
	}

	switch value.(type) {
	case types.Int64:
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return
		}
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return
		}
		setTFModelField(model, tfFieldKey, types.Int64Value(i))
	case types.Bool:
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return
		}
		b, err := strconv.ParseBool(s)
		if err != nil {
			return
		}
		setTFModelField(model, tfFieldKey, types.BoolValue(b))
	case types.String:
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return
		}
		setTFModelField(model, tfFieldKey, types.StringValue(s))
	case types.Set:
		elems := extractStringSliceFromFlatRaw(raw)
		if len(elems) == 0 {
			return
		}
		fieldSet, diags := types.SetValueFrom(ctx, types.StringType, elems)
		if diags.HasError() {
			return
		}
		setTFModelField(model, tfFieldKey, fieldSet)
	case types.List:
		elems := extractStringSliceFromFlatRaw(raw)
		if len(elems) == 0 {
			return
		}
		fieldList, diags := types.ListValueFrom(ctx, types.StringType, elems)
		if diags.HasError() {
			return
		}
		setTFModelField(model, tfFieldKey, fieldList)
	}
}

func setTFModelField(model *tfModel, tfFieldKey string, value attr.Value) {
	rv := reflect.ValueOf(model).Elem()
	rt := rv.Type()
	for i := range rt.NumField() {
		field := rt.Field(i)
		if field.Tag.Get("tfsdk") != tfFieldKey {
			continue
		}
		fieldVal := rv.Field(i)
		valReflect := reflect.ValueOf(value)
		if valReflect.Type().AssignableTo(fieldVal.Type()) {
			fieldVal.Set(valReflect)
		}
		return
	}
}

// populateOperationalDefaults sets provider-side defaults when each field is null.
func populateOperationalDefaults(model *tfModel) {
	if model.DeletionProtection.IsNull() || model.DeletionProtection.IsUnknown() {
		model.DeletionProtection = types.BoolValue(true)
	}
	if model.WaitForActiveShards.IsNull() || model.WaitForActiveShards.IsUnknown() {
		model.WaitForActiveShards = types.StringValue("1")
	}
	if model.MasterTimeout.IsNull() || model.MasterTimeout.IsUnknown() {
		model.MasterTimeout = customtypes.NewDurationValue("30s")
	}
	if model.Timeout.IsNull() || model.Timeout.IsUnknown() {
		model.Timeout = customtypes.NewDurationValue("30s")
	}
}

// pruneImportHydratedPlanFields nulls plan fields for Optional-only settings that
// are absent from configuration after import hydration.
func pruneImportHydratedPlanFields(ctx context.Context, plan, config *tfModel) {
	modelType := reflect.TypeFor[tfModel]()
	planVal := reflect.ValueOf(plan).Elem()

	for _, fieldKey := range importHydrationPrunableFieldKeys {
		configField, ok := config.getFieldValueByTagValue(fieldKey, modelType)
		if !ok || !configField.IsNull() {
			continue
		}

		for i := range modelType.NumField() {
			field := modelType.Field(i)
			if field.Tag.Get("tfsdk") != fieldKey {
				continue
			}
			planField := planVal.Field(i)
			planAttr, ok := planField.Interface().(attr.Value)
			if !ok {
				break
			}
			nullVal := nullAttrValueForField(ctx, planAttr)
			if nullVal != nil {
				planField.Set(reflect.ValueOf(nullVal))
			}
			break
		}
	}
}

func nullAttrValueForField(ctx context.Context, v attr.Value) attr.Value {
	switch val := v.(type) {
	case types.Int64:
		return types.Int64Null()
	case types.Bool:
		return types.BoolNull()
	case types.String:
		return types.StringNull()
	case types.Set:
		return types.SetNull(val.ElementType(ctx))
	case types.List:
		return types.ListNull(val.ElementType(ctx))
	case jsontypes.Normalized:
		return jsontypes.NewNormalizedNull()
	default:
		return nil
	}
}
