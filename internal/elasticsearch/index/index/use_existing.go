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
	"sort"
	"strconv"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// staticSettingMismatch records one static index setting that differs between
// the Terraform plan and an existing index returned by Get Index.
type staticSettingMismatch struct {
	Attribute  string
	Configured string
	Actual     string
}

// compareStaticSettings walks staticSettingsKeys and reports mismatches for
// settings that are explicitly set in the plan (Known and non-null) but differ
// from existing.Settings. Plan fields that are null or unknown are skipped.
//
// Elasticsearch flat settings commonly use keys prefixed with "index."; bare
// keys are also checked.
func compareStaticSettings(ctx context.Context, plan *tfModel, existing models.Index) ([]staticSettingMismatch, diag.Diagnostics) {
	if plan == nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic("compareStaticSettings", "plan is nil"),
		}
	}

	var diags diag.Diagnostics
	var mismatches []staticSettingMismatch

	modelType := reflect.TypeFor[tfModel]()
	settings := existing.Settings
	if settings == nil {
		settings = map[string]any{}
	}

	for _, key := range staticSettingsKeys {
		tfFieldKey := convertSettingsKeyToTFFieldKey(key)
		planField, ok := plan.getFieldValueByTagValue(tfFieldKey, modelType)
		if !ok {
			diags.AddError(
				"failed to find setting field",
				fmt.Sprintf("expected field with tfsdk tag %s", tfFieldKey),
			)
			return nil, diags
		}

		if planField == nil || planField.IsNull() || planField.IsUnknown() {
			continue
		}

		actualRaw, found := lookupExistingSetting(settings, key)
		if !found {
			mismatches = append(mismatches, staticSettingMismatch{
				Attribute:  tfFieldKey,
				Configured: configuredStringFromPlan(ctx, planField),
				Actual:     "<absent>",
			})
			continue
		}

		if mm := compareStaticSettingValue(ctx, key, tfFieldKey, planField, actualRaw); mm != nil {
			mismatches = append(mismatches, *mm)
		}
	}

	return mismatches, diags
}

func lookupExistingSetting(settings map[string]any, key string) (any, bool) {
	prefixed := "index." + key
	if v, ok := settings[prefixed]; ok && v != nil {
		return v, true
	}
	if v, ok := settings[key]; ok && v != nil {
		return v, true
	}
	return nil, false
}

func configuredStringFromPlan(ctx context.Context, planVal attr.Value) string {
	switch v := planVal.(type) {
	case types.Int64:
		return strconv.FormatInt(v.ValueInt64(), 10)
	case types.Bool:
		return strconv.FormatBool(v.ValueBool())
	case types.String:
		return v.ValueString()
	case types.Set:
		var elems []string
		_ = v.ElementsAs(ctx, &elems, true)
		sort.Strings(elems)
		return strings.Join(elems, ", ")
	case types.List:
		var elems []string
		_ = v.ElementsAs(ctx, &elems, true)
		return strings.Join(elems, ", ")
	default:
		return fmt.Sprintf("%v", planVal)
	}
}

func compareStaticSettingValue(ctx context.Context, esKey, tfAttr string, planVal attr.Value, actualRaw any) *staticSettingMismatch {
	switch esKey {
	case "number_of_shards", "number_of_routing_shards", "routing_partition_size":
		pInt, ok := planVal.(types.Int64)
		if !ok {
			return mismatch(tfAttr, configuredStringFromPlan(ctx, planVal), fmt.Sprintf("(unexpected plan type %T)", planVal))
		}
		planI := pInt.ValueInt64()
		cfg := strconv.FormatInt(planI, 10)
		actI, actStr, ok := int64FromAny(actualRaw)
		if !ok {
			return mismatch(tfAttr, cfg, actStr)
		}
		if planI != actI {
			return mismatch(tfAttr, cfg, actStr)
		}
		return nil

	case "codec", "shard.check_on_startup":
		pStr, ok := planVal.(types.String)
		if !ok {
			return mismatch(tfAttr, configuredStringFromPlan(ctx, planVal), fmt.Sprintf("(unexpected plan type %T)", planVal))
		}
		planS := pStr.ValueString()
		actS := stringFromAny(actualRaw)
		if planS != actS {
			return mismatch(tfAttr, planS, actS)
		}
		return nil

	case "load_fixed_bitset_filters_eagerly", "mapping.coerce":
		pBool, ok := planVal.(types.Bool)
		if !ok {
			return mismatch(tfAttr, configuredStringFromPlan(ctx, planVal), fmt.Sprintf("(unexpected plan type %T)", planVal))
		}
		planB := pBool.ValueBool()
		cfg := strconv.FormatBool(planB)
		actB, actStr, ok := boolFromAny(actualRaw)
		if !ok {
			return mismatch(tfAttr, cfg, actStr)
		}
		if planB != actB {
			return mismatch(tfAttr, cfg, actStr)
		}
		return nil

	case "sort.field":
		pSet, ok := planVal.(types.Set)
		if !ok {
			return mismatch(tfAttr, configuredStringFromPlan(ctx, planVal), fmt.Sprintf("(unexpected plan type %T)", planVal))
		}
		var planElems []string
		_ = pSet.ElementsAs(ctx, &planElems, true)
		cfg := configuredStringFromPlan(ctx, planVal)
		actSlice := stringSliceFromSortFieldAny(actualRaw)
		if !equalAsStringSets(planElems, actSlice) {
			return mismatch(tfAttr, cfg, formatAsSortedUniqueSet(actSlice))
		}
		return nil

	case "sort.order":
		pList, ok := planVal.(types.List)
		if !ok {
			return mismatch(tfAttr, configuredStringFromPlan(ctx, planVal), fmt.Sprintf("(unexpected plan type %T)", planVal))
		}
		var planElems []string
		_ = pList.ElementsAs(ctx, &planElems, true)
		cfg := configuredStringFromPlan(ctx, planVal)
		actSlice := stringSliceOrderedFromAny(actualRaw)
		if !slicesEqual(planElems, actSlice) {
			return mismatch(tfAttr, cfg, strings.Join(actSlice, ", "))
		}
		return nil

	default:
		return mismatch(tfAttr, configuredStringFromPlan(ctx, planVal), fmt.Sprintf("(internal: unknown static key %q)", esKey))
	}
}

func mismatch(attr, configured, actual string) *staticSettingMismatch {
	return &staticSettingMismatch{
		Attribute:  attr,
		Configured: configured,
		Actual:     actual,
	}
}

func int64FromAny(v any) (val int64, display string, ok bool) {
	switch x := v.(type) {
	case string:
		i, err := strconv.ParseInt(x, 10, 64)
		if err != nil {
			return 0, x, false
		}
		return i, x, true
	case json.Number:
		i, err := x.Int64()
		if err != nil {
			return 0, x.String(), false
		}
		return i, x.String(), true
	case float64:
		i := int64(x)
		if float64(i) != x {
			return 0, fmt.Sprintf("%v", x), false
		}
		return i, strconv.FormatInt(i, 10), true
	case int:
		return int64(x), strconv.Itoa(x), true
	case int64:
		return x, strconv.FormatInt(x, 10), true
	case int32:
		return int64(x), strconv.FormatInt(int64(x), 10), true
	default:
		return 0, fmt.Sprintf("%v", v), false
	}
}

func boolFromAny(v any) (val bool, display string, ok bool) {
	switch x := v.(type) {
	case string:
		b, err := strconv.ParseBool(x)
		if err != nil {
			return false, x, false
		}
		return b, x, true
	case bool:
		return x, strconv.FormatBool(x), true
	default:
		return false, fmt.Sprintf("%v", v), false
	}
}

func stringFromAny(v any) string {
	switch x := v.(type) {
	case string:
		return x
	default:
		return fmt.Sprintf("%v", v)
	}
}

func stringSliceFromSortFieldAny(v any) []string {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case string:
		return []string{x}
	case []string:
		return append([]string(nil), x...)
	case []any:
		out := make([]string, 0, len(x))
		for _, e := range x {
			out = append(out, elemToString(e))
		}
		return out
	default:
		return []string{fmt.Sprintf("%v", v)}
	}
}

func stringSliceOrderedFromAny(v any) []string {
	// Same shapes as sort.field; order preserved for list-like values.
	return stringSliceFromSortFieldAny(v)
}

func elemToString(e any) string {
	if e == nil {
		return ""
	}
	switch t := e.(type) {
	case string:
		return t
	default:
		return fmt.Sprint(t)
	}
}

func sortUniqueStrings(s []string) []string {
	m := make(map[string]struct{}, len(s))
	for _, x := range s {
		m[x] = struct{}{}
	}
	out := make([]string, 0, len(m))
	for x := range m {
		out = append(out, x)
	}
	sort.Strings(out)
	return out
}

func equalAsStringSets(a, b []string) bool {
	sa := sortUniqueStrings(a)
	sb := sortUniqueStrings(b)
	if len(sa) != len(sb) {
		return false
	}
	for i := range sa {
		if sa[i] != sb[i] {
			return false
		}
	}
	return true
}

func formatAsSortedUniqueSet(elems []string) string {
	su := sortUniqueStrings(elems)
	return strings.Join(su, ", ")
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
