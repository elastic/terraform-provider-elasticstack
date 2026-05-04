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

		if !reflect.DeepEqual(userVal, apiVal) {
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

	userType, userHasType := userField["type"]
	apiType, apiHasType := apiField["type"]

	if userHasType && apiHasType {
		if !reflect.DeepEqual(userType, apiType) {
			return false
		}
	} else if userHasType || apiHasType {
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

		if !reflect.DeepEqual(userVal, apiVal) {
			return false
		}
	}

	return true
}
