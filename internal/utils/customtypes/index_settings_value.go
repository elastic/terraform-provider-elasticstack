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

package customtypes

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable                    = IndexSettingsType{}
	_ basetypes.StringValuable                   = (*IndexSettingsValue)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*IndexSettingsValue)(nil)
	_ xattr.ValidateableAttribute                = (*IndexSettingsValue)(nil)
)

// IndexSettingsType is a Terraform Plugin Framework string type for Elasticsearch index settings JSON objects.
type IndexSettingsType struct {
	jsontypes.NormalizedType
}

// String returns a human readable string of the type name.
func (t IndexSettingsType) String() string {
	return "customtypes.IndexSettingsType"
}

// ValueType returns the Value type.
func (t IndexSettingsType) ValueType(_ context.Context) attr.Value {
	return IndexSettingsValue{}
}

// Equal returns true if the given type is equivalent.
func (t IndexSettingsType) Equal(o attr.Type) bool {
	other, ok := o.(IndexSettingsType)
	if !ok {
		return false
	}
	return t.NormalizedType.Equal(other.NormalizedType)
}

// ValueFromString returns a StringValuable type given a StringValue.
func (t IndexSettingsType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return IndexSettingsValue{Normalized: jsontypes.Normalized{StringValue: in}}, nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.
func (t IndexSettingsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.NormalizedType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	norm, ok := attrValue.(jsontypes.Normalized)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	return IndexSettingsValue{Normalized: norm}, nil
}

// IndexSettingsValue holds a JSON object string for index template settings with semantic equality matching DiffIndexSettingSuppress.
type IndexSettingsValue struct {
	jsontypes.Normalized
}

// Type returns an IndexSettingsType.
func (v IndexSettingsValue) Type(_ context.Context) attr.Type {
	return IndexSettingsType{}
}

// Equal returns true if the given value is equivalent.
func (v IndexSettingsValue) Equal(o attr.Value) bool {
	other, ok := o.(IndexSettingsValue)
	if !ok {
		return false
	}
	return v.Normalized.Equal(other.Normalized)
}

// ValidateAttribute ensures the value is valid JSON and unmarshals to a JSON object (map), consistent with stringIsJSONObject.
func (v IndexSettingsValue) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if v.IsNull() || v.IsUnknown() {
		return
	}

	v.Normalized.ValidateAttribute(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(v.ValueString()), &m); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"expected value to be a JSON object",
			fmt.Sprintf("This value must be an object, not a simple type or array. Check the documentation for the expected format. %s", err),
		)
	}
}

// StringSemanticEquals compares normalized flattened index settings (dotted keys, index. prefix, stringified values).
// It shadows jsontypes.Normalized.StringSemanticEquals on the embedded field so index-setting semantics apply for
// Terraform drift and apply consistency. The index template resource also applies a plan modifier that rewrites
// planned settings to CanonicalIndexSettingsJSON so state matches Elasticsearch's nested {"index":{...}} shape.
func (v IndexSettingsValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(IndexSettingsValue)
	if !ok {
		diags.AddError(
			"Semantic equality check error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)
		return false, diags
	}

	return v.SemanticallyEqual(ctx, newValue)
}

// SemanticallyEqual is the same comparison as StringSemanticEquals for explicit IndexSettingsValue pairs
// (plan reconciliation helpers).
func (v IndexSettingsValue) SemanticallyEqual(_ context.Context, other IndexSettingsValue) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	if v.IsNull() {
		return other.IsNull(), diags
	}

	if v.IsUnknown() {
		return other.IsUnknown(), diags
	}

	if other.IsNull() || other.IsUnknown() {
		return false, diags
	}

	var o, n map[string]any
	if err := json.Unmarshal([]byte(v.ValueString()), &o); err != nil {
		return false, diags
	}
	if err := json.Unmarshal([]byte(other.ValueString()), &n); err != nil {
		return false, diags
	}

	return reflect.DeepEqual(
		normalizeIndexSettings(flattenMap(o)),
		normalizeIndexSettings(flattenMap(n)),
	), diags
}

func normalizeIndexSettings(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, val := range m {
		if strings.HasPrefix(k, "index.") {
			out[k] = fmt.Sprintf("%v", val)
			continue
		}
		out[fmt.Sprintf("index.%s", k)] = fmt.Sprintf("%v", val)
	}
	return out
}

// CanonicalIndexSettingsJSON returns compact JSON for the same effective index settings as raw,
// in the nested shape Elasticsearch uses (e.g. {"index":{"number_of_shards":"3"}}).
// It applies the same flattening, index.-prefix normalization, and stringification as semantic equality.
func CanonicalIndexSettingsJSON(raw string) (string, error) {
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return "", fmt.Errorf("unmarshal settings JSON: %w", err)
	}
	flatNorm := normalizeIndexSettings(flattenMap(m))
	nested := unflattenDottedMap(flatNorm)
	b, err := json.Marshal(nested)
	if err != nil {
		return "", fmt.Errorf("marshal canonical settings: %w", err)
	}
	return string(b), nil
}

// unflattenDottedMap turns dotted keys (after normalizeIndexSettings, e.g. index.number_of_shards)
// into a nested map for JSON matching Elasticsearch's index settings object shape.
func unflattenDottedMap(flat map[string]any) map[string]any {
	root := make(map[string]any)
	for k, v := range flat {
		parts := strings.Split(k, ".")
		cur := root
		for i := range parts {
			p := parts[i]
			if i == len(parts)-1 {
				cur[p] = v
				break
			}
			existing, ok := cur[p]
			if !ok {
				nm := make(map[string]any)
				cur[p] = nm
				cur = nm
				continue
			}
			nm, ok := existing.(map[string]any)
			if !ok {
				nm = make(map[string]any)
				cur[p] = nm
			}
			cur = nm
		}
	}
	return root
}

// flattenMap flattens nested maps into dotted keys (port of tfsdkutils.flattenMap).
func flattenMap(m map[string]any) map[string]any {
	out := make(map[string]any)

	var flattener func(string, map[string]any, map[string]any)
	flattener = func(k string, src, dst map[string]any) {
		if len(k) > 0 {
			k += "."
		}
		for key, val := range src {
			switch inner := val.(type) {
			case map[string]any:
				flattener(k+key, inner, dst)
			default:
				dst[k+key] = val
			}
		}
	}
	flattener("", m, out)
	return out
}

// NewIndexSettingsNull creates an IndexSettingsValue with a null value.
func NewIndexSettingsNull() IndexSettingsValue {
	return IndexSettingsValue{Normalized: jsontypes.NewNormalizedNull()}
}

// NewIndexSettingsUnknown creates an IndexSettingsValue with an unknown value.
func NewIndexSettingsUnknown() IndexSettingsValue {
	return IndexSettingsValue{Normalized: jsontypes.NewNormalizedUnknown()}
}

// NewIndexSettingsValue creates an IndexSettingsValue with a known value.
func NewIndexSettingsValue(value string) IndexSettingsValue {
	return IndexSettingsValue{Normalized: jsontypes.NewNormalizedValue(value)}
}
