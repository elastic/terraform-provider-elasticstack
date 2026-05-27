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
	"fmt"
	"reflect"
	"strconv"
	"strings"

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
	for _, key := range allSettingsKeys {
		if sortKeysExpandedFromNestedBlock[key] {
			continue
		}
		importHydrationPrunableFieldKeys = append(importHydrationPrunableFieldKeys, convertSettingsKeyToTFFieldKey(key))
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

	if raw, ok := flat["index.analysis"]; ok {
		hydrateAnalysisFromRaw(model, raw)
	}

	if raw, ok := flat["index."+settingQueryDefaultField]; ok {
		hydrateQueryDefaultFieldFromRaw(ctx, model, raw)
	}

	modelType := reflect.TypeFor[tfModel]()
	for _, key := range allSettingsKeys {
		if sortKeysExpandedFromNestedBlock[key] {
			continue
		}
		if key == settingQueryDefaultField {
			continue
		}

		raw, ok := flat["index."+key]
		if !ok {
			continue
		}

		tfFieldKey := convertSettingsKeyToTFFieldKey(key)
		setFlatSettingOnModel(ctx, model, tfFieldKey, raw, modelType)
	}

	return nil
}

func hydrateAnalysisFromRaw(model *tfModel, raw json.RawMessage) {
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
// value (scalar string or JSON array), mirroring extractSortSetting.
func extractStringSliceFromFlatRaw(raw json.RawMessage) []string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		trimmed := strings.TrimSpace(s)
		if strings.HasPrefix(trimmed, "[") {
			var arr []string
			if err := json.Unmarshal([]byte(trimmed), &arr); err == nil {
				return arr
			}
		}
		if trimmed != "" {
			return []string{trimmed}
		}
		return nil
	}

	var arr []string
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr
	}

	var arrAny []any
	if err := json.Unmarshal(raw, &arrAny); err == nil {
		result := make([]string, len(arrAny))
		for i, e := range arrAny {
			result[i] = fmt.Sprint(e)
		}
		return result
	}

	return nil
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
			nullVal := nullAttrValueForField(ctx, planField.Interface().(attr.Value))
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
