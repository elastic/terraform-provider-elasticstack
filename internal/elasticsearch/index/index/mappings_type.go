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

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable = (*mappingsType)(nil)
)

// mappingsType is a custom string type for index mappings that implements
// semantic equality to treat template-injected mapping content as non-drift.
type mappingsType struct {
	jsontypes.NormalizedType
}

func (t mappingsType) String() string {
	return "index.MappingsType"
}

func (t mappingsType) Equal(o attr.Type) bool {
	other, ok := o.(mappingsType)
	if !ok {
		return false
	}
	return t.NormalizedType.Equal(other.NormalizedType)
}

func (t mappingsType) ValueType(_ context.Context) attr.Value {
	return mappingsValue{}
}

func (t mappingsType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	if in.IsNull() {
		return NewMappingsNull(), nil
	}
	if in.IsUnknown() {
		return NewMappingsUnknown(), nil
	}
	return NewMappingsValue(in.ValueString()), nil
}

func (t mappingsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.NormalizedType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	normalized, ok := attrValue.(jsontypes.Normalized)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	return mappingsValue{
		Normalized: normalized,
	}, nil
}

var (
	_ basetypes.StringValuable                   = (*mappingsValue)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*mappingsValue)(nil)
)

// mappingsValue is the value type for index mappings that implements semantic
// equality to treat template-injected mapping content as non-drift.
type mappingsValue struct {
	jsontypes.Normalized
}

func (v mappingsValue) Type(_ context.Context) attr.Type {
	return mappingsType{}
}

func (v mappingsValue) Equal(o attr.Value) bool {
	other, ok := o.(mappingsValue)
	if !ok {
		return false
	}
	return v.Normalized.Equal(other.Normalized)
}

// StringSemanticEquals returns true if the refreshed/API mappings are a
// non-drifting superset of the prior user-intent mappings.
func (v mappingsValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(mappingsValue)
	if !ok {
		// Fall back to standard normalized comparison for unexpected types
		return v.Normalized.StringSemanticEquals(context.Background(), newValuable)
	}

	if v.IsNull() || v.IsUnknown() {
		return v.Normalized.Equal(newValue.Normalized), diags
	}

	if newValue.IsNull() || newValue.IsUnknown() {
		return v.Normalized.Equal(newValue.Normalized), diags
	}

	var priorMap, newMap map[string]any
	if err := json.Unmarshal([]byte(v.ValueString()), &priorMap); err != nil {
		diags.AddError("Semantic Equality Check Error", err.Error())
		return false, diags
	}
	if err := json.Unmarshal([]byte(newValue.ValueString()), &newMap); err != nil {
		diags.AddError("Semantic Equality Check Error", err.Error())
		return false, diags
	}

	// priorMap = API/state value (may contain template extras)
	// newMap   = planned/config value (user intent)
	// We check if priorMap is a non-drifting superset of newMap.
	return mappingsSemanticallyEqual(newMap, priorMap), diags
}

func NewMappingsNull() mappingsValue {
	return mappingsValue{
		Normalized: jsontypes.NewNormalizedNull(),
	}
}

func NewMappingsUnknown() mappingsValue {
	return mappingsValue{
		Normalized: jsontypes.NewNormalizedUnknown(),
	}
}

func NewMappingsValue(value string) mappingsValue {
	return mappingsValue{
		Normalized: jsontypes.NewNormalizedValue(value),
	}
}
