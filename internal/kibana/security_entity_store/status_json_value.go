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
	"fmt"
	"sort"

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

// StatusJSONType is a Terraform Plugin Framework string type for the raw JSON body
// of the Kibana entity store status endpoint (GET /api/security/entity_store/status), whose
// top-level "engines" array may be returned in any order by the API.
type StatusJSONType struct {
	jsontypes.NormalizedType
}

// String returns a human readable string of the type name.
func (t StatusJSONType) String() string {
	return "security_entity_store.StatusJSONType"
}

// ValueType returns the Value type.
func (t StatusJSONType) ValueType(_ context.Context) attr.Value {
	return StatusJSONValue{}
}

// Equal returns true if the given type is equivalent.
func (t StatusJSONType) Equal(o attr.Type) bool {
	other, ok := o.(StatusJSONType)
	if !ok {
		return false
	}
	return t.NormalizedType.Equal(other.NormalizedType)
}

// ValueFromString returns a StringValuable type given a StringValue.
func (t StatusJSONType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return StatusJSONValue{StringValue: in}, nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.
func (t StatusJSONType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.NormalizedType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	norm, ok := attrValue.(jsontypes.Normalized)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	return StatusJSONValue{Normalized: norm}, nil
}

// StatusJSONValue holds the raw JSON body of the Kibana entity store status
// endpoint. The stored string is byte-identical to the API response; engine array ordering
// is ignored only during StringSemanticEquals.
type StatusJSONValue struct {
	jsontypes.Normalized
}

// Type returns an StatusJSONType.
func (v StatusJSONValue) Type(_ context.Context) attr.Type {
	return StatusJSONType{}
}

// Equal returns true if the given value is equivalent.
func (v StatusJSONValue) Equal(o attr.Value) bool {
	other, ok := o.(StatusJSONValue)
	if !ok {
		return false
	}
	return v.Normalized.Equal(other.Normalized)
}

// StringSemanticEquals returns true if the given value is semantically equal to the receiver.
// It shadows jsontypes.Normalized.StringSemanticEquals on the embedded field so the "engines"
// array order is ignored: both raw JSON strings are copied and canonically engine-sorted, then
// compared using the embedded jsontypes.Normalized.StringSemanticEquals, which normalizes
// whitespace, key order, and string escape representation. JSON number literal representation
// is significant (the library decodes with encoding/json's UseNumber, so 1e2 and 100 do NOT
// compare equal) and array order other than the top-level "engines" key stays significant.
func (v StatusJSONValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	newValue, ok, diags := typeutils.AssertSameType(v, newValuable)
	if !ok {
		return false, diags
	}

	return v.SemanticallyEqual(ctx, newValue)
}

// SemanticallyEqual is the same comparison as StringSemanticEquals for explicit
// StatusJSONValue pairs (e.g. import-time semantic checks in acceptance tests).
func (v StatusJSONValue) SemanticallyEqual(ctx context.Context, other StatusJSONValue) (bool, diag.Diagnostics) {
	if v.IsNull() {
		return other.IsNull(), nil
	}
	if v.IsUnknown() {
		return other.IsUnknown(), nil
	}
	if other.IsNull() || other.IsUnknown() {
		return false, nil
	}

	vCopy := StatusJSONValue{Normalized: jsontypes.NewNormalizedValue(canonicalizeStatusJSONEngines(v.ValueString()))}
	otherCopy := StatusJSONValue{Normalized: jsontypes.NewNormalizedValue(canonicalizeStatusJSONEngines(other.ValueString()))}
	return vCopy.Normalized.StringSemanticEquals(ctx, otherCopy.Normalized)
}

// canonicalizeStatusJSONEngines returns rawJSON with the top-level "engines" array stable-sorted
// by each engine's "type" field so two responses differing only in engine order compare equal.
// Bodies it cannot decode (malformed JSON, non-object engines, engines elements that are not
// JSON objects or whose "type" is missing or not a JSON string) are returned unchanged, so
// comparison falls back to plain jsontypes.Normalized JSON semantics. A JSON null or missing
// "engines" key is preserved as-is (it never becomes an empty array).
func canonicalizeStatusJSONEngines(rawJSON string) string {
	rawBody := []byte(rawJSON)
	var body map[string]json.RawMessage
	if err := json.Unmarshal(rawBody, &body); err != nil {
		return rawJSON
	}
	var engines []json.RawMessage
	if err := json.Unmarshal(body["engines"], &engines); err != nil {
		// A missing "engines" key unmarshals as empty input, which errors.
		return rawJSON
	}
	// A JSON null "engines" decodes without error into a nil slice; re-marshaling
	// a nil slice would rewrite it as an empty array, so it is preserved as-is.
	if engines == nil {
		return rawJSON
	}
	pairs := make([]statusJSONKeyedEngine, len(engines))
	for i, e := range engines {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(e, &fields); err != nil {
			return rawJSON
		}
		// An explicit JSON null "type" decodes without error into the zero string
		// and sorts first. A missing "type" key (which unmarshals as empty input
		// and errors) or a non-string "type" triggers the raw fallback above.
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

// statusJSONKeyedEngine pairs a raw engine JSON payload with its decoded "type" field so the
// stable sort permutes keys and payloads together, preserving the relative order of duplicate
// type keys.
type statusJSONKeyedEngine struct {
	typeKey string
	raw     json.RawMessage
}

// NewStatusJSONNull creates an StatusJSONValue with a null value.
func NewStatusJSONNull() StatusJSONValue {
	return StatusJSONValue{Normalized: jsontypes.NewNormalizedNull()}
}

// NewStatusJSONUnknown creates an StatusJSONValue with an unknown value.
func NewStatusJSONUnknown() StatusJSONValue {
	return StatusJSONValue{Normalized: jsontypes.NewNormalizedUnknown()}
}

// NewStatusJSONValue creates an StatusJSONValue with a known value,
// stored byte-identically to the given raw JSON body (no normalization or re-marshaling).
func NewStatusJSONValue(value string) StatusJSONValue {
	return StatusJSONValue{Normalized: jsontypes.NewNormalizedValue(value)}
}
