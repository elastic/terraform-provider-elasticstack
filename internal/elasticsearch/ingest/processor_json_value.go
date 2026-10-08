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

package ingest

import (
	"context"
	"reflect"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable                    = ProcessorJSONType{}
	_ basetypes.StringValuable                   = (*ProcessorJSONValue)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*ProcessorJSONValue)(nil)
)

// ProcessorJSONType is a custom string type for ingest processor JSON.
type ProcessorJSONType struct {
	customtypes.NormalizedJSONType
}

func (t ProcessorJSONType) String() string {
	return "ingest.ProcessorJSONType"
}

func (t ProcessorJSONType) ValueType(ctx context.Context) attr.Value {
	return ProcessorJSONValue{
		NormalizedJSONValue: t.NormalizedJSONType.ValueType(ctx).(customtypes.NormalizedJSONValue),
	}
}

func (t ProcessorJSONType) Equal(o attr.Type) bool {
	other, ok := o.(ProcessorJSONType)
	if !ok {
		return false
	}
	return t.NormalizedJSONType.Equal(other.NormalizedJSONType)
}

func (t ProcessorJSONType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	if in.IsNull() {
		return NewProcessorJSONNull(), nil
	}
	if in.IsUnknown() {
		return NewProcessorJSONUnknown(), nil
	}
	return NewProcessorJSONValue(in.ValueString()), nil
}

func (t ProcessorJSONType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	return typeutils.StringTypableValueFromTerraform(ctx, t.StringType, t.ValueFromString, in)
}

// ProcessorJSONValue is a custom string value for ingest processor JSON.
//
// The typed go-elasticsearch client represents fields like remove.field and
// append.value as slices ([]string, []json.RawMessage). When users write a
// single string value the client round-trips it as a single-element array,
// causing "Provider produced inconsistent result after apply". The semantic
// equality implementation normalizes single-element primitive arrays to scalars
// before comparison so both forms are treated as equivalent.
type ProcessorJSONValue struct {
	customtypes.NormalizedJSONValue
}

func (v ProcessorJSONValue) Type(ctx context.Context) attr.Type {
	return ProcessorJSONType{
		NormalizedJSONType: v.NormalizedJSONValue.Type(ctx).(customtypes.NormalizedJSONType),
	}
}

func (v ProcessorJSONValue) Equal(o attr.Value) bool {
	other, ok := o.(ProcessorJSONValue)
	if !ok {
		return false
	}
	return v.NormalizedJSONValue.Equal(other.NormalizedJSONValue)
}

// StringSemanticEquals returns true when two processor JSON values are
// equivalent after normalizing single-element arrays to scalars.
func (v ProcessorJSONValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	newValue, ok := newValuable.(ProcessorJSONValue)
	if !ok {
		return v.Normalized.StringSemanticEquals(ctx, newValuable)
	}

	var diags diag.Diagnostics

	if v.IsNull() || v.IsUnknown() || newValue.IsNull() || newValue.IsUnknown() {
		return v.Normalized.Equal(newValue.Normalized), diags
	}

	vMap, diags := typeutils.UnmarshalJSONForSemanticEquals[any](v.ValueString())
	if diags.HasError() {
		return false, diags
	}

	newMap, diags := typeutils.UnmarshalJSONForSemanticEquals[any](newValue.ValueString())
	if diags.HasError() {
		return false, diags
	}

	return reflect.DeepEqual(
		normalizeProcessorJSON(vMap),
		normalizeProcessorJSON(newMap),
	), diags
}

// normalizeProcessorJSON recursively collapses single-element arrays containing
// a primitive value (string, number, bool) into the scalar value itself. This
// compensates for the typed go-elasticsearch client converting fields like
// remove.field from a string to a []string on deserialization.
func normalizeProcessorJSON(v any) any {
	switch val := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, vv := range val {
			out[k] = normalizeProcessorJSON(vv)
		}
		return out
	case []any:
		if len(val) == 1 {
			switch val[0].(type) {
			case string, float64, bool:
				return val[0]
			}
		}
		out := make([]any, len(val))
		for i, vv := range val {
			out[i] = normalizeProcessorJSON(vv)
		}
		return out
	default:
		return v
	}
}

func NewProcessorJSONNull() ProcessorJSONValue {
	return ProcessorJSONValue{NormalizedJSONValue: customtypes.NewNormalizedJSONNull()}
}

func NewProcessorJSONUnknown() ProcessorJSONValue {
	return ProcessorJSONValue{NormalizedJSONValue: customtypes.NewNormalizedJSONUnknown()}
}

func NewProcessorJSONValue(value string) ProcessorJSONValue {
	return ProcessorJSONValue{NormalizedJSONValue: customtypes.NewNormalizedJSONValue(value)}
}
