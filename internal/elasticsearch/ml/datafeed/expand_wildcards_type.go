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

package datafeed

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.SetTypable                    = (*ExpandWildcardsType)(nil)
	_ basetypes.SetValuableWithSemanticEquals = (*ExpandWildcardsValue)(nil)
)

// expandWildcardsAllTokens are the constituent values that Elasticsearch
// normalizes "all" into when returning a datafeed configuration.
var expandWildcardsAllTokens = map[string]struct{}{
	"open":   {},
	"closed": {},
	"hidden": {},
}

// ExpandWildcardsType is a custom set type for indices_options.expand_wildcards
// that supports semantic equality. Elasticsearch normalizes the shorthand token
// "all" into ["open", "closed", "hidden"], so these two values are considered
// semantically equal.
type ExpandWildcardsType struct {
	basetypes.SetType
}

func (t ExpandWildcardsType) String() string {
	return "datafeed.ExpandWildcardsType"
}

func (t ExpandWildcardsType) ValueType(_ context.Context) attr.Value {
	return ExpandWildcardsValue{
		SetValue: basetypes.NewSetUnknown(t.ElementType()),
	}
}

func (t ExpandWildcardsType) Equal(o attr.Type) bool {
	other, ok := o.(ExpandWildcardsType)
	if !ok {
		return false
	}
	return t.SetType.Equal(other.SetType)
}

func (t ExpandWildcardsType) ValueFromSet(_ context.Context, in basetypes.SetValue) (basetypes.SetValuable, diag.Diagnostics) {
	return ExpandWildcardsValue{SetValue: in}, nil
}

func (t ExpandWildcardsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.SetType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	setValue, ok := attrValue.(basetypes.SetValue)
	if !ok {
		return nil, fmt.Errorf("unexpected type %T, expected basetypes.SetValue", attrValue)
	}

	return ExpandWildcardsValue{SetValue: setValue}, nil
}

// ExpandWildcardsValue is a custom set value for indices_options.expand_wildcards
// that implements semantic equality. Elasticsearch normalizes "all" →
// ["open", "closed", "hidden"], so these representations are treated as equal.
type ExpandWildcardsValue struct {
	basetypes.SetValue
}

func (v ExpandWildcardsValue) Type(_ context.Context) attr.Type {
	return ExpandWildcardsType{SetType: basetypes.SetType{ElemType: types.StringType}}
}

func (v ExpandWildcardsValue) Equal(o attr.Value) bool {
	other, ok := o.(ExpandWildcardsValue)
	if !ok {
		return false
	}
	return v.SetValue.Equal(other.SetValue)
}

func (v ExpandWildcardsValue) ToSetValue(_ context.Context) (basetypes.SetValue, diag.Diagnostics) {
	return v.SetValue, nil
}

// SetSemanticEquals returns true if the two ExpandWildcardsValue instances are
// semantically equal. The key rule: "all" expands to {"open","closed","hidden"},
// so ["all"] and ["open","closed","hidden"] (in any order) are equal.
func (v ExpandWildcardsValue) SetSemanticEquals(_ context.Context, priorValuable basetypes.SetValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	priorValue, ok := priorValuable.(ExpandWildcardsValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", priorValuable),
		)
		return false, diags
	}

	// Handle null/unknown conservatively.
	if v.IsNull() {
		return priorValue.IsNull(), diags
	}
	if v.IsUnknown() {
		return priorValue.IsUnknown(), diags
	}
	if priorValue.IsNull() || priorValue.IsUnknown() {
		return false, diags
	}

	thisNormalized := normalizeExpandWildcards(v)
	priorNormalized := normalizeExpandWildcards(priorValue)

	if len(thisNormalized) != len(priorNormalized) {
		return false, diags
	}

	for token := range thisNormalized {
		if _, found := priorNormalized[token]; !found {
			return false, diags
		}
	}

	return true, diags
}

// normalizeExpandWildcards returns a set of string tokens after expanding
// the shorthand "all" into its constituent values {"open","closed","hidden"}.
// All other tokens are kept as-is.
func normalizeExpandWildcards(v ExpandWildcardsValue) map[string]struct{} {
	result := make(map[string]struct{})
	for _, elem := range v.Elements() {
		s, ok := elem.(types.String)
		if !ok || s.IsNull() || s.IsUnknown() {
			// Treat non-string or null/unknown elements as opaque literals.
			result[fmt.Sprintf("%v", elem)] = struct{}{}
			continue
		}
		if s.ValueString() == "all" {
			for token := range expandWildcardsAllTokens {
				result[token] = struct{}{}
			}
		} else {
			result[s.ValueString()] = struct{}{}
		}
	}
	return result
}

// NewExpandWildcardsNull returns an ExpandWildcardsValue with a null value.
func NewExpandWildcardsNull() ExpandWildcardsValue {
	return ExpandWildcardsValue{
		SetValue: basetypes.NewSetNull(types.StringType),
	}
}

// NewExpandWildcardsUnknown returns an ExpandWildcardsValue with an unknown value.
func NewExpandWildcardsUnknown() ExpandWildcardsValue {
	return ExpandWildcardsValue{
		SetValue: basetypes.NewSetUnknown(types.StringType),
	}
}

// NewExpandWildcardsValue returns an ExpandWildcardsValue with a known value
// constructed from the given string elements.
func NewExpandWildcardsValue(elements []attr.Value) (ExpandWildcardsValue, diag.Diagnostics) {
	setValue, diags := basetypes.NewSetValue(types.StringType, elements)
	return ExpandWildcardsValue{
		SetValue: setValue,
	}, diags
}
