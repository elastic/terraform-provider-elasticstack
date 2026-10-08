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

package security_entity_store

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable                    = StatusJSONType{}
	_ basetypes.StringValuable                   = StatusJSONValue{}
	_ basetypes.StringValuableWithSemanticEquals = (*StatusJSONValue)(nil)
)

// StatusJSONType preserves raw status JSON while ignoring engine order during comparison.
type StatusJSONType struct {
	customtypes.NormalizedJSONType
}

func (t StatusJSONType) String() string {
	return "security_entity_store.StatusJSONType"
}

func (t StatusJSONType) ValueType(ctx context.Context) attr.Value {
	return StatusJSONValue{
		NormalizedJSONValue: t.NormalizedJSONType.ValueType(ctx).(customtypes.NormalizedJSONValue),
	}
}

func (t StatusJSONType) Equal(o attr.Type) bool {
	other, ok := o.(StatusJSONType)
	if !ok {
		return false
	}
	return t.NormalizedJSONType.Equal(other.NormalizedJSONType)
}

func (t StatusJSONType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	val, diags := t.NormalizedJSONType.ValueFromString(ctx, in)
	if diags.HasError() {
		return nil, diags
	}
	return StatusJSONValue{
		NormalizedJSONValue: val.(customtypes.NormalizedJSONValue),
	}, nil
}

func (t StatusJSONType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	return typeutils.StringTypableValueFromTerraform(ctx, t.StringType, t.ValueFromString, in)
}

// StatusJSONValue preserves the API response without canonicalizing its stored string.
type StatusJSONValue struct {
	customtypes.NormalizedJSONValue
}

func (v StatusJSONValue) Type(ctx context.Context) attr.Type {
	return StatusJSONType{
		NormalizedJSONType: v.NormalizedJSONValue.Type(ctx).(customtypes.NormalizedJSONType),
	}
}

func (v StatusJSONValue) Equal(o attr.Value) bool {
	other, ok := o.(StatusJSONValue)
	if !ok {
		return false
	}
	return v.NormalizedJSONValue.Equal(other.NormalizedJSONValue)
}

// StringSemanticEquals ignores top-level engine order while retaining Normalized's
// comparison semantics, including significant number literals and other array orders.
func (v StatusJSONValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	newValue, ok, diags := typeutils.AssertSameType(v, newValuable)
	if !ok {
		return false, diags
	}

	return v.SemanticallyEqual(ctx, newValue)
}

func (v StatusJSONValue) SemanticallyEqual(ctx context.Context, other StatusJSONValue) (bool, diag.Diagnostics) {
	if v.IsNull() {
		return other.IsNull(), nil
	}
	if v.IsUnknown() {
		return other.IsUnknown(), nil
	}
	if !typeutils.IsKnown(other) {
		return false, nil
	}

	// Identical invalid JSON must still reach the library's error diagnostics.
	if v.ValueString() == other.ValueString() && json.Valid([]byte(v.ValueString())) {
		return true, nil
	}

	canonical := jsontypes.NewNormalizedValue(canonicalizeStatusJSONEngines(v.ValueString()))
	otherCanonical := jsontypes.NewNormalizedValue(canonicalizeStatusJSONEngines(other.ValueString()))
	return canonical.StringSemanticEquals(ctx, otherCanonical)
}

// Failed engine decoding falls back to Normalized's ordered-array comparison.
// RawMessage preserves unmodeled fields and number literals in the comparison copy.
func canonicalizeStatusJSONEngines(rawJSON string) string {
	rawBody := []byte(rawJSON)
	var body map[string]json.RawMessage
	if err := json.Unmarshal(rawBody, &body); err != nil {
		return rawJSON
	}
	var engines []json.RawMessage
	if err := json.Unmarshal(body["engines"], &engines); err != nil {
		return rawJSON
	}
	// Keep null distinct from the empty array built by the sorted copy.
	if engines == nil {
		return rawJSON
	}
	pairs := make([]statusJSONKeyedEngine, len(engines))
	for i, e := range engines {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(e, &fields); err != nil {
			return rawJSON
		}
		// JSON null decodes to an empty key; missing or non-string types fail.
		var typeKey string
		if err := json.Unmarshal(fields["type"], &typeKey); err != nil {
			return rawJSON
		}
		pairs[i] = statusJSONKeyedEngine{typeKey: typeKey, raw: e}
	}
	sort.SliceStable(pairs, func(i, j int) bool { return pairs[i].typeKey < pairs[j].typeKey })
	sortedEngines := make([]json.RawMessage, len(pairs))
	for i, p := range pairs {
		sortedEngines[i] = p.raw
	}
	sortedBytes, err := json.Marshal(sortedEngines)
	if err != nil {
		return rawJSON
	}
	body["engines"] = sortedBytes
	normalized, err := json.Marshal(body)
	if err != nil {
		return rawJSON
	}
	return string(normalized)
}

type statusJSONKeyedEngine struct {
	typeKey string
	raw     json.RawMessage
}

func NewStatusJSONNull() StatusJSONValue {
	return StatusJSONValue{NormalizedJSONValue: customtypes.NewNormalizedJSONNull()}
}

func NewStatusJSONUnknown() StatusJSONValue {
	return StatusJSONValue{NormalizedJSONValue: customtypes.NewNormalizedJSONUnknown()}
}

func NewStatusJSONValue(value string) StatusJSONValue {
	return StatusJSONValue{NormalizedJSONValue: customtypes.NewNormalizedJSONValue(value)}
}
