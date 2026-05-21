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

package rolemapping

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable                    = NormalizedRulesType{}
	_ basetypes.StringValuable                   = NormalizedRulesValue{}
	_ basetypes.StringValuableWithSemanticEquals = NormalizedRulesValue{}
)

// NormalizedRulesType is the attr.Type companion for NormalizedRulesValue.
type NormalizedRulesType struct {
	jsontypes.NormalizedType
}

// String returns a human readable string of the type name.
func (t NormalizedRulesType) String() string {
	return "rolemapping.NormalizedRulesType"
}

// ValueType returns the Value type.
func (t NormalizedRulesType) ValueType(_ context.Context) attr.Value {
	return NormalizedRulesValue{}
}

// Equal returns true if the given type is equivalent.
func (t NormalizedRulesType) Equal(o attr.Type) bool {
	other, ok := o.(NormalizedRulesType)
	if !ok {
		return false
	}
	return t.NormalizedType.Equal(other.NormalizedType)
}

// ValueFromString returns a StringValuable type given a StringValue.
func (t NormalizedRulesType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return NormalizedRulesValue{Normalized: jsontypes.Normalized{StringValue: in}}, nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.
func (t NormalizedRulesType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.NormalizedType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	norm, ok := attrValue.(jsontypes.Normalized)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	return NormalizedRulesValue{Normalized: norm}, nil
}

// NormalizedRulesValue is a jsontypes.Normalized subtype that treats
// single-element arrays and plain strings as semantically equal inside
// "field" rule objects to handle the ES normalization behavior.
type NormalizedRulesValue struct {
	jsontypes.Normalized
}

// Type returns a NormalizedRulesType.
func (v NormalizedRulesValue) Type(_ context.Context) attr.Type {
	return NormalizedRulesType{}
}

// Equal returns true if the given value is equivalent.
func (v NormalizedRulesValue) Equal(o attr.Value) bool {
	other, ok := o.(NormalizedRulesValue)
	if !ok {
		return false
	}
	return v.Normalized.Equal(other.Normalized)
}

// StringSemanticEquals returns true when both JSON rule strings are logically
// equal after collapsing single-element arrays inside "field" objects.
func (v NormalizedRulesValue) StringSemanticEquals(ctx context.Context, other basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	otherRules, ok := other.(NormalizedRulesValue)
	if !ok {
		return v.Normalized.StringSemanticEquals(ctx, other)
	}
	if v.IsNull() {
		return otherRules.IsNull(), diags
	}
	if v.IsUnknown() {
		return otherRules.IsUnknown(), diags
	}
	if otherRules.IsNull() || otherRules.IsUnknown() {
		return false, diags
	}
	thisNorm, err1 := normalizeRulesJSONString(v.ValueString())
	thatNorm, err2 := normalizeRulesJSONString(otherRules.ValueString())
	if err1 != nil || err2 != nil {
		return v.Normalized.StringSemanticEquals(ctx, otherRules.Normalized)
	}
	return jsontypes.NewNormalizedValue(thisNorm).StringSemanticEquals(ctx, jsontypes.NewNormalizedValue(thatNorm))
}

// NewNormalizedRulesValue creates a NormalizedRulesValue with a known value.
func NewNormalizedRulesValue(v string) NormalizedRulesValue {
	return NormalizedRulesValue{Normalized: jsontypes.NewNormalizedValue(v)}
}

// NewNormalizedRulesNull creates a NormalizedRulesValue with a null value.
func NewNormalizedRulesNull() NormalizedRulesValue {
	return NormalizedRulesValue{Normalized: jsontypes.NewNormalizedNull()}
}

// NewNormalizedRulesUnknown creates a NormalizedRulesValue with an unknown value.
func NewNormalizedRulesUnknown() NormalizedRulesValue {
	return NormalizedRulesValue{Normalized: jsontypes.NewNormalizedUnknown()}
}

// normalizeRulesJSONString parses a JSON string and collapses single-element
// arrays inside "field" objects to plain string values.
func normalizeRulesJSONString(raw string) (string, error) {
	var tree map[string]any
	if err := json.Unmarshal([]byte(raw), &tree); err != nil {
		return "", err
	}
	normalizeRuleNode(tree)
	out, err := json.Marshal(tree)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// normalizeRuleNode walks a parsed JSON rule tree and collapses
// single-element arrays inside "field" objects to plain string values.
func normalizeRuleNode(node any) {
	switch v := node.(type) {
	case map[string]any:
		if field, ok := v["field"]; ok {
			if fieldMap, ok := field.(map[string]any); ok {
				for key, val := range fieldMap {
					if arr, ok := val.([]any); ok && len(arr) == 1 {
						fieldMap[key] = arr[0]
					}
				}
			}
		}
		for _, child := range v {
			normalizeRuleNode(child)
		}
	case []any:
		for _, child := range v {
			normalizeRuleNode(child)
		}
	}
}
