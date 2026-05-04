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
	"sort"
	"strconv"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// staticSettingMismatch records one static index setting that differs between
// the Terraform plan and an existing index returned by Get Index.
type staticSettingMismatch struct {
	Attribute  string
	Configured string
	Actual     string
}

// compareStaticSettings walks staticSettingsKeys and reports mismatches for
// settings explicitly present in the plan after merging typed attributes and
// the deprecated `settings` block (see tfModel.toIndexSettings). Keys absent
// from that merged map are skipped. Values are compared to existing.Settings,
// using keys with an `index.` prefix first, then bare keys (Elasticsearch flat
// settings).
func compareStaticSettings(ctx context.Context, plan *tfModel, existing models.Index) ([]staticSettingMismatch, diag.Diagnostics) {
	if plan == nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic("compareStaticSettings", "plan is nil"),
		}
	}

	planSettings, diags := plan.toIndexSettings(ctx)
	if diags.HasError() {
		return nil, diags
	}

	settings := existing.Settings
	if settings == nil {
		settings = map[string]any{}
	}

	var mismatches []staticSettingMismatch

	for _, key := range staticSettingsKeys {
		planVal, ok := planSettings[key]
		if !ok {
			continue
		}

		tfAttr := convertSettingsKeyToTFFieldKey(key)
		actualRaw, found := lookupExistingSetting(settings, key)
		if !found {
			mismatches = append(mismatches, staticSettingMismatch{
				Attribute:  tfAttr,
				Configured: configuredDisplayFromPlanValue(key, planVal),
				Actual:     "<absent>",
			})
			continue
		}

		if mm := compareStaticPlanAndES(tfAttr, key, planVal, actualRaw); mm != nil {
			mismatches = append(mismatches, *mm)
		}
	}

	return mismatches, nil
}

// formatStaticSettingMismatchesDetail builds the error detail string for adopt-time
// static setting mismatches (used by Create).
func formatStaticSettingMismatchesDetail(concreteName string, mismatches []staticSettingMismatch) string {
	var b strings.Builder
	fmt.Fprintf(&b, "concrete_name: %s\n\n", concreteName)
	for _, m := range mismatches {
		fmt.Fprintf(&b, "%s: configured=%s, actual=%s\n", m.Attribute, m.Configured, m.Actual)
	}
	return strings.TrimSuffix(b.String(), "\n")
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

func configuredDisplayFromPlanValue(key string, planVal any) string {
	switch key {
	case "number_of_shards", "number_of_routing_shards", "routing_partition_size":
		i, s, ok := int64FromAny(planVal)
		if ok {
			return strconv.FormatInt(i, 10)
		}
		return s
	case "codec", "shard.check_on_startup":
		return stringFromPlanScalar(planVal)
	case "load_fixed_bitset_filters_eagerly", "mapping.coerce":
		b, s, ok := boolFromAny(planVal)
		if ok {
			return strconv.FormatBool(b)
		}
		return s
	case "sort.field":
		_, d := planStringSliceForSortFieldSet(planVal)
		return d
	case "sort.order":
		_, d := planStringSliceForSortOrder(planVal)
		return d
	default:
		return fmt.Sprint(planVal)
	}
}

func compareStaticPlanAndES(tfAttr, key string, planVal, actualRaw any) *staticSettingMismatch {
	switch key {
	case "number_of_shards", "number_of_routing_shards", "routing_partition_size":
		planI, cfg, okP := int64FromAny(planVal)
		if !okP {
			_, actStr, _ := int64FromAny(actualRaw)
			return mismatch(tfAttr, cfg, actStr)
		}
		actI, actStr, okA := int64FromAny(actualRaw)
		if !okA {
			return mismatch(tfAttr, cfg, actStr)
		}
		if planI != actI {
			return mismatch(tfAttr, cfg, actStr)
		}
		return nil

	case "codec", "shard.check_on_startup":
		planS := stringFromPlanScalar(planVal)
		actS := stringFromAny(actualRaw)
		if planS != actS {
			return mismatch(tfAttr, planS, actS)
		}
		return nil

	case "load_fixed_bitset_filters_eagerly", "mapping.coerce":
		planB, cfg, okP := boolFromAny(planVal)
		if !okP {
			_, actStr, _ := boolFromAny(actualRaw)
			return mismatch(tfAttr, cfg, actStr)
		}
		actB, actStr, okA := boolFromAny(actualRaw)
		if !okA {
			return mismatch(tfAttr, cfg, actStr)
		}
		if planB != actB {
			return mismatch(tfAttr, cfg, actStr)
		}
		return nil

	case "sort.field":
		planElems, cfg := planStringSliceForSortFieldSet(planVal)
		actSlice := stringSliceFromSortFieldAny(actualRaw)
		if !equalAsStringSets(planElems, actSlice) {
			return mismatch(tfAttr, cfg, formatAsSortedUniqueSet(actSlice))
		}
		return nil

	case "sort.order":
		planElems, cfg := planStringSliceForSortOrder(planVal)
		actSlice := stringSliceOrderedFromAny(actualRaw)
		if !slicesEqual(planElems, actSlice) {
			return mismatch(tfAttr, cfg, strings.Join(actSlice, ", "))
		}
		return nil

	default:
		return mismatch(tfAttr, configuredDisplayFromPlanValue(key, planVal), fmt.Sprintf("(internal: unknown static key %q)", key))
	}
}

func stringFromPlanScalar(v any) string {
	switch x := v.(type) {
	case string:
		return x
	default:
		return fmt.Sprintf("%v", v)
	}
}

func planStringSliceForSortFieldSet(v any) ([]string, string) {
	switch x := v.(type) {
	case []string:
		cp := append([]string(nil), x...)
		return cp, formatAsSortedUniqueSet(cp)
	case []any:
		out := make([]string, 0, len(x))
		for _, e := range x {
			out = append(out, elemToString(e))
		}
		return out, formatAsSortedUniqueSet(out)
	case string:
		trimmed := strings.TrimSpace(x)
		if trimmed == "" {
			return nil, ""
		}
		var arr []string
		if err := json.Unmarshal([]byte(trimmed), &arr); err == nil {
			if len(arr) == 0 {
				return nil, ""
			}
			return arr, formatAsSortedUniqueSet(arr)
		}
		one := []string{trimmed}
		return one, formatAsSortedUniqueSet(one)
	default:
		s := []string{fmt.Sprint(v)}
		return s, formatAsSortedUniqueSet(s)
	}
}

func planStringSliceForSortOrder(v any) ([]string, string) {
	switch x := v.(type) {
	case []string:
		cp := append([]string(nil), x...)
		return cp, strings.Join(cp, ", ")
	case []any:
		out := make([]string, 0, len(x))
		for _, e := range x {
			out = append(out, elemToString(e))
		}
		return out, strings.Join(out, ", ")
	case string:
		trimmed := strings.TrimSpace(x)
		if trimmed == "" {
			return nil, ""
		}
		var arr []string
		if err := json.Unmarshal([]byte(trimmed), &arr); err == nil {
			if len(arr) == 0 {
				return nil, ""
			}
			return arr, strings.Join(arr, ", ")
		}
		return []string{trimmed}, trimmed
	default:
		s := fmt.Sprint(v)
		return []string{s}, s
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
