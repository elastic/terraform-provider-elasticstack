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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
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
		return
	}
	if m == nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"expected value to be a JSON object",
			"This value must be an object, not the JSON null literal. Check the documentation for the expected format.",
		)
	}
}

// StringSemanticEquals compares normalized flattened index settings (dotted keys, index. prefix, stringified values).
// It shadows jsontypes.Normalized.StringSemanticEquals on the embedded field so index-setting semantics apply for
// Terraform drift and apply consistency between the user's input form and the canonical
// {"index":{...}} shape Elasticsearch returns.
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
		diags.AddError(
			"Invalid index settings JSON",
			fmt.Sprintf("Failed to parse prior index settings as JSON during semantic comparison: %s", err.Error()),
		)
		return false, diags
	}
	if err := json.Unmarshal([]byte(other.ValueString()), &n); err != nil {
		diags.AddError(
			"Invalid index settings JSON",
			fmt.Sprintf("Failed to parse new index settings as JSON during semantic comparison: %s", err.Error()),
		)
		return false, diags
	}

	return reflect.DeepEqual(
		normalizeIndexSettings(typeutils.FlattenMap(o)),
		normalizeIndexSettings(typeutils.FlattenMap(n)),
	), diags
}

// normalizeIndexSettings stringifies values and ensures every key uses an index.* prefix for comparison.
//
// Merge rule when flattening produces the same canonical key from a top-level setting and a nested
// "index": {...} path (e.g. number_of_shards vs index.number_of_shards): the nested form wins. We
// apply flat keys first, then keys whose pre-prefix path contains a dot (nested/dotted sources), so
// the latter overwrite deterministically. Remaining ties use sorted (canonical key, flat key) order.
func normalizeIndexSettings(m map[string]any) map[string]any {
	type entry struct {
		flatKey string
		nk      string
		val     string
	}
	flat := make([]entry, 0)
	dotted := make([]entry, 0)
	for k, val := range m {
		nk := k
		if !strings.HasPrefix(k, "index.") {
			nk = "index." + k
		}
		e := entry{k, nk, fmt.Sprintf("%v", val)}
		if strings.Contains(k, ".") {
			dotted = append(dotted, e)
		} else {
			flat = append(flat, e)
		}
	}
	sort.Slice(flat, func(i, j int) bool {
		if flat[i].nk != flat[j].nk {
			return flat[i].nk < flat[j].nk
		}
		return flat[i].flatKey < flat[j].flatKey
	})
	sort.Slice(dotted, func(i, j int) bool {
		if dotted[i].nk != dotted[j].nk {
			return dotted[i].nk < dotted[j].nk
		}
		return dotted[i].flatKey < dotted[j].flatKey
	})
	out := make(map[string]any, len(m))
	for _, e := range flat {
		out[e.nk] = e.val
	}
	for _, e := range dotted {
		out[e.nk] = e.val
	}
	return out
}

// CanonicalIndexSettingsJSON returns compact JSON for the same effective index settings as raw,
// in the nested shape Elasticsearch uses (e.g. {"index":{"number_of_shards":"3"}}).
// It applies the same flattening, index.-prefix normalization, and stringification as semantic equality.
//
// When the same logical setting appears as both a top-level key and under a nested path (e.g.
// number_of_shards alongside index.number_of_shards), normalizeIndexSettings applies a deterministic
// merge: values whose original flat key contained a dot (nested/dotted source) overwrite values
// from single-segment keys sharing the same canonical index.* key. Object keys in the output are
// sorted recursively so repeated calls return an identical byte string.
func CanonicalIndexSettingsJSON(raw string) (string, error) {
	var top any
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &top); err != nil {
		return "", fmt.Errorf("unmarshal settings JSON: %w", err)
	}
	if top == nil {
		return "", fmt.Errorf("settings must be a JSON object, not null")
	}
	m, ok := top.(map[string]any)
	if !ok {
		return "", fmt.Errorf("settings must be a JSON object")
	}
	flatNorm := normalizeIndexSettings(typeutils.FlattenMap(m))
	nested := unflattenDottedMap(flatNorm)
	b, err := marshalSettingsJSONSorted(nested)
	if err != nil {
		return "", fmt.Errorf("marshal canonical settings: %w", err)
	}
	return string(b), nil
}

// marshalSettingsJSONSorted marshals maps with sorted keys at every object level so canonical
// settings strings are stable across Go releases and map iteration order.
func marshalSettingsJSONSorted(v any) ([]byte, error) {
	switch t := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var buf bytes.Buffer
		buf.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			keyBytes, err := json.Marshal(k)
			if err != nil {
				return nil, err
			}
			buf.Write(keyBytes)
			buf.WriteByte(':')
			valBytes, err := marshalSettingsJSONSorted(t[k])
			if err != nil {
				return nil, err
			}
			buf.Write(valBytes)
		}
		buf.WriteByte('}')
		return buf.Bytes(), nil
	default:
		return json.Marshal(t)
	}
}

// unflattenDottedMap turns dotted keys (after normalizeIndexSettings, e.g. index.number_of_shards)
// into a nested map for JSON matching Elasticsearch's index settings object shape.
//
// Conflict handling: keys are processed in lexicographic order so output is deterministic
// regardless of Go's randomized map iteration. When the same prefix appears as both a leaf
// (e.g. "index.foo") and a nested path (e.g. "index.foo.bar"), the deeper/nested path wins
// because it is sorted after the shorter key and overwrites the leaf. Such conflicts are not
// expected for valid Elasticsearch index settings; this rule simply guarantees stability.
func unflattenDottedMap(flat map[string]any) map[string]any {
	root := make(map[string]any)
	keys := make([]string, 0, len(flat))
	for k := range flat {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := flat[k]
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
