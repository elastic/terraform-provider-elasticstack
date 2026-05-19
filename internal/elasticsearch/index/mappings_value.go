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

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable                    = MappingsType{}
	_ basetypes.StringValuable                   = (*MappingsValue)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*MappingsValue)(nil)
)

// MappingsType is a custom string type for Elasticsearch index/template mappings JSON.
type MappingsType struct {
	jsontypes.NormalizedType
}

// String returns a human readable string of the type name.
func (t MappingsType) String() string {
	return "index.MappingsType"
}

// ValueType returns the Value type.
func (t MappingsType) ValueType(_ context.Context) attr.Value {
	return MappingsValue{}
}

// Equal returns true if the given type is equivalent.
func (t MappingsType) Equal(o attr.Type) bool {
	other, ok := o.(MappingsType)
	if !ok {
		return false
	}
	return t.NormalizedType.Equal(other.NormalizedType)
}

// ValueFromString returns a StringValuable type given a StringValue.
func (t MappingsType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	if in.IsNull() {
		return NewMappingsNull(), nil
	}
	if in.IsUnknown() {
		return NewMappingsUnknown(), nil
	}
	return NewMappingsValue(in.ValueString()), nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.
func (t MappingsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.NormalizedType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	normalized, ok := attrValue.(jsontypes.Normalized)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	return MappingsValue{
		Normalized: normalized,
	}, nil
}

// MappingsValue is a custom string value type for Elasticsearch index/template mappings JSON.
//
// Normalization is baked into construction: NewMappingsValue strips implicit
// "type":"object" entries that the typed go-elasticsearch client injects when
// serialising ObjectProperty values. Without this, ES round-trips add spurious
// keys that cause "Provider produced inconsistent result after apply" errors.
//
// StringSemanticEquals treats the API value as a non-drifting superset of user
// intent, mirroring the behaviour previously embedded in the index resource's
// private mappingsValue type. This allows template-injected extras (extra
// properties, dynamic_templates, _meta, etc.) to not trigger plan changes.
type MappingsValue struct {
	jsontypes.Normalized
}

// Type returns an MappingsType.
func (v MappingsValue) Type(_ context.Context) attr.Type {
	return MappingsType{}
}

// Equal returns true if the given value is equivalent.
func (v MappingsValue) Equal(o attr.Value) bool {
	other, ok := o.(MappingsValue)
	if !ok {
		return false
	}
	return v.Normalized.Equal(other.Normalized)
}

// StringSemanticEquals returns true if the refreshed/API mappings are a
// non-drifting superset of the prior user-intent mappings.
func (v MappingsValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(MappingsValue)
	if !ok {
		// Fall back to standard normalized comparison for unexpected types.
		return v.Normalized.StringSemanticEquals(ctx, newValuable)
	}

	if v.IsNull() || v.IsUnknown() {
		return v.Normalized.Equal(newValue.Normalized), diags
	}

	if newValue.IsNull() || newValue.IsUnknown() {
		return v.Normalized.Equal(newValue.Normalized), diags
	}

	var vMap, newMap map[string]any
	if err := json.Unmarshal([]byte(v.ValueString()), &vMap); err != nil {
		diags.AddError("Semantic Equality Check Error", err.Error())
		return false, diags
	}
	if err := json.Unmarshal([]byte(newValue.ValueString()), &newMap); err != nil {
		diags.AddError("Semantic Equality Check Error", err.Error())
		return false, diags
	}

	// Semantic equality for mappings is bidirectional: two mapping values are
	// semantically equal when one is a non-drifting superset of the other.
	// This handles both planning (plan vs prior state) and apply (state vs plan).
	return MappingsSemanticallyEqual(vMap, newMap) || MappingsSemanticallyEqual(newMap, vMap), diags
}

// normalizeMappings recursively normalises a decoded mapping tree:
//
//   - Strips implicit "type":"object" from any node that also contains a
//     "properties" key. The typed go-elasticsearch client always injects this
//     via ObjectProperty.MarshalJSON even when the original JSON omitted it,
//     causing spurious drift between plan and state on first create.
//   - Collapses single-element string arrays for dynamic-template keys that
//     Elasticsearch accepts as either a string or an array
//     (match, match_mapping_type, path_match, path_unmatch, unmatch,
//     unmatch_mapping_type).
//   - Converts string-encoded JSON booleans and null back to their native types
//     so that ImportState round-trips produce the same stored value as the
//     original apply (Elasticsearch echoes some boolean fields such as
//     dynamic as the JSON string "false" instead of the boolean false).
func normalizeMappings(v any) any {
	switch val := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, vv := range val {
			normalized := normalizeMappings(vv)
			switch k {
			case "match", "match_mapping_type", "path_match", "path_unmatch",
				"unmatch", "unmatch_mapping_type":
				if arr, ok := normalized.([]any); ok && len(arr) == 1 {
					if s, ok := arr[0].(string); ok {
						normalized = s
					}
				}
			}
			out[k] = normalized
		}

		// Strip "type":"object" when "properties" is also present: it is the
		// implicit default type and the typed ES client always injects it even
		// when absent from the original JSON.
		if typeVal, hasType := out["type"]; hasType && typeVal == "object" {
			if _, hasProps := out["properties"]; hasProps {
				delete(out, "type")
			}
		}

		return out
	case []any:
		out := make([]any, len(val))
		for i, vv := range val {
			out[i] = normalizeMappings(vv)
		}
		return out
	case string:
		// Convert string-encoded JSON booleans and null back to their native
		// types. Elasticsearch echoes some mapping fields (e.g. dynamic) as
		// JSON strings instead of booleans. Normalizing here ensures the stored
		// value after import matches the value stored after the initial apply,
		// so ImportStateVerify does not fail due to "false" vs false.
		return typeutils.NormalizeJSONScalar(val)
	default:
		return v
	}
}

// NewMappingsNull creates an MappingsValue with a null value.
func NewMappingsNull() MappingsValue {
	return MappingsValue{Normalized: jsontypes.NewNormalizedNull()}
}

// NewMappingsUnknown creates an MappingsValue with an unknown value.
func NewMappingsUnknown() MappingsValue {
	return MappingsValue{Normalized: jsontypes.NewNormalizedUnknown()}
}

// NewMappingsValue creates an MappingsValue with the given JSON string,
// applying normalization (implicit "type":"object" stripping, single-element
// array collapsing) before storing the value.
func NewMappingsValue(value string) MappingsValue {
	var m any
	if err := json.Unmarshal([]byte(value), &m); err == nil {
		m = normalizeMappings(m)
		if nb, err := json.Marshal(m); err == nil {
			value = string(nb)
		}
	}
	return MappingsValue{Normalized: jsontypes.NewNormalizedValue(value)}
}

// ---- semantic equality helpers -----------------------------------------------

// scalarSemanticEqual returns true when two scalar leaf values are semantically
// equal, accounting for the case where Elasticsearch echoes a non-string scalar
// (bool, number) back as its stringified form. For example, the user-authored
// JSON value true (decoded as bool) must compare equal to the API-echoed "true"
// (decoded as string), and vice-versa.
//
// Only plain scalar types (bool, float64, string) are handled here. For
// identical structured values (map, slice) the reflect.DeepEqual fast path at
// the top of this function returns true; non-identical structured values fall
// through to the default case and return false, so structural comparison is
// not affected.
func scalarSemanticEqual(a, b any) bool {
	if reflect.DeepEqual(a, b) {
		return true
	}

	// Require at least one side to be a non-string scalar so that two
	// structurally different strings are never falsely equated.
	_, aIsString := a.(string)
	_, bIsString := b.(string)
	if aIsString && bIsString {
		return false
	}

	// One side is a non-string scalar. Determine whether a non-string scalar
	// on one side matches a stringified equivalent on the other.
	switch x := a.(type) {
	case bool:
		var apiStr string
		if s, ok := b.(string); ok {
			apiStr = s
		} else {
			return false
		}
		if x {
			return apiStr == "true"
		}
		return apiStr == "false"
	case float64:
		// json.Unmarshal decodes numbers as float64 when UseNumber is not set.
		// Compare numerically to avoid scientific-notation mismatches from %g.
		apiStr, ok := b.(string)
		if !ok {
			return false
		}
		if sv, err := strconv.ParseFloat(apiStr, 64); err == nil {
			return x == sv
		}
		return false
	case string:
		// a is a string, so b must be a non-string scalar (ensured by the
		// guard above). Recurse with sides swapped so b's case is handled.
		return scalarSemanticEqual(b, a)
	default:
		return false
	}
}

// SemanticTextMappingType is the Elasticsearch mapping type string for semantic_text fields.
const SemanticTextMappingType = "semantic_text"

// MappingsSemanticallyEqual compares user-owned mappings against API mappings.
// It returns true when the API value is a non-drifting superset of user intent:
//   - All user-owned properties exist in the API with matching types.
//   - Template-injected extras (extra properties, dynamic_templates, _meta, …) are allowed.
//   - semantic_text model_settings auto-populated by ES are allowed.
func MappingsSemanticallyEqual(userMappings, apiMappings map[string]any) bool {
	if len(userMappings) == 0 && len(apiMappings) == 0 {
		return true
	}

	for key, userVal := range userMappings {
		apiVal, ok := apiMappings[key]
		if !ok {
			return false
		}

		if key == "properties" {
			userProps, ok := userVal.(map[string]any)
			if !ok {
				return false
			}
			apiProps, ok := apiVal.(map[string]any)
			if !ok {
				return false
			}
			if !propertiesSemanticallyEqual(userProps, apiProps) {
				return false
			}
			continue
		}

		if !scalarSemanticEqual(userVal, apiVal) {
			return false
		}
	}

	return true
}

// propertiesSemanticallyEqual recursively checks that all user-owned properties
// exist in the API with semantically equal definitions.
func propertiesSemanticallyEqual(userProps, apiProps map[string]any) bool {
	for fieldName, userFieldRaw := range userProps {
		apiFieldRaw, ok := apiProps[fieldName]
		if !ok {
			return false
		}
		if !fieldSemanticallyEqual(userFieldRaw, apiFieldRaw) {
			return false
		}
	}
	return true
}

// FieldSemanticallyEqual checks if two mapping object values are semantically equal,
// allowing for ES-auto-populated values such as semantic_text model_settings.
func FieldSemanticallyEqual(userFieldRaw, apiFieldRaw any) bool {
	return fieldSemanticallyEqual(userFieldRaw, apiFieldRaw)
}

// fieldSemanticallyEqual checks if two field definitions are semantically equal,
// allowing for ES-auto-populated values such as semantic_text model_settings.
func fieldSemanticallyEqual(userFieldRaw, apiFieldRaw any) bool {
	userField, ok := userFieldRaw.(map[string]any)
	if !ok {
		return false
	}
	apiField, ok := apiFieldRaw.(map[string]any)
	if !ok {
		return false
	}

	for key, userVal := range userField {
		apiVal, ok := apiField[key]
		if !ok {
			return false
		}

		if key == "properties" {
			userProps, ok := userVal.(map[string]any)
			if !ok {
				return false
			}
			apiProps, ok := apiVal.(map[string]any)
			if !ok {
				return false
			}
			if !propertiesSemanticallyEqual(userProps, apiProps) {
				return false
			}
			continue
		}

		if key == "script" {
			if scriptSemanticallyEqual(userVal, apiVal) {
				continue
			}
			return false
		}

		if userMap, userIsMap := userVal.(map[string]any); userIsMap {
			if apiMap, apiIsMap := apiVal.(map[string]any); apiIsMap {
				if fieldSemanticallyEqual(userMap, apiMap) {
					continue
				}
				return false
			}
		}

		if !scalarSemanticEqual(userVal, apiVal) {
			return false
		}
	}

	return true
}

// scriptSemanticallyEqual allows a user-authored script string to match Elasticsearch's
// expanded script object form returned by the API.
func scriptSemanticallyEqual(userVal, apiVal any) bool {
	userStr, userIsStr := userVal.(string)
	apiMap, apiIsMap := apiVal.(map[string]any)
	if userIsStr && apiIsMap {
		source, ok := apiMap["source"].(string)
		return ok && userStr == source
	}
	apiStr, apiIsStr := apiVal.(string)
	userMap, userIsMap := userVal.(map[string]any)
	if apiIsStr && userIsMap {
		source, ok := userMap["source"].(string)
		return ok && apiStr == source
	}
	return scalarSemanticEqual(userVal, apiVal)
}
