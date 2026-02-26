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

package slo

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
	_ basetypes.ListTypable                    = (*GroupByType)(nil)
	_ basetypes.ListValuableWithSemanticEquals = (*GroupByValue)(nil)
)

// GroupByType is a custom list type for group_by that supports semantic equality.
// Kibana treats an empty group_by list as equivalent to ["*"].
type GroupByType struct {
	basetypes.ListType
}

func (t GroupByType) String() string {
	return "slo.GroupByType"
}

func (t GroupByType) ValueType(_ context.Context) attr.Value {
	return GroupByValue{
		ListValue: basetypes.NewListUnknown(t.ElementType()),
	}
}

func (t GroupByType) Equal(o attr.Type) bool {
	other, ok := o.(GroupByType)
	if !ok {
		return false
	}
	return t.ListType.Equal(other.ListType)
}

func (t GroupByType) ValueFromList(_ context.Context, in basetypes.ListValue) (basetypes.ListValuable, diag.Diagnostics) {
	return GroupByValue{ListValue: in}, nil
}

func (t GroupByType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.ListType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	listValue, ok := attrValue.(basetypes.ListValue)
	if !ok {
		return nil, fmt.Errorf("unexpected type %T, expected basetypes.ListValue", attrValue)
	}

	return GroupByValue{ListValue: listValue}, nil
}

func NewGroupByType() GroupByType {
	return GroupByType{
		ListType: basetypes.ListType{
			ElemType: types.StringType,
		},
	}
}

// GroupByValue is a custom list value for group_by that implements semantic equality.
// Kibana considers [] semantically equal to ["*"].
type GroupByValue struct {
	basetypes.ListValue
}

func (v GroupByValue) Type(_ context.Context) attr.Type {
	return NewGroupByType()
}

func (v GroupByValue) Equal(o attr.Value) bool {
	other, ok := o.(GroupByValue)
	if !ok {
		return false
	}
	return v.ListValue.Equal(other.ListValue)
}

func (v GroupByValue) ListSemanticEquals(_ context.Context, priorValuable basetypes.ListValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	priorValue, ok := priorValuable.(GroupByValue)
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

	if v.IsNull() {
		return priorValue.IsNull(), diags
	}
	if v.IsUnknown() {
		return priorValue.IsUnknown(), diags
	}
	if priorValue.IsNull() || priorValue.IsUnknown() {
		return false, diags
	}

	// Semantic rule: [] <=> ["*"]
	thisEmpty := len(v.Elements()) == 0
	priorEmpty := len(priorValue.Elements()) == 0
	if thisEmpty && priorEmpty {
		return true, diags
	}

	thisStar := isSingleKnownStar(v)
	priorStar := isSingleKnownStar(priorValue)
	if (thisEmpty && priorStar) || (priorEmpty && thisStar) {
		return true, diags
	}

	return v.ListValue.Equal(priorValue.ListValue), diags
}

func isSingleKnownStar(v GroupByValue) bool {
	elems := v.Elements()
	if len(elems) != 1 {
		return false
	}

	s, ok := elems[0].(types.String)
	if !ok || s.IsNull() || s.IsUnknown() {
		return false
	}

	return s.ValueString() == "*"
}

func NewGroupByNull() GroupByValue {
	return GroupByValue{
		ListValue: basetypes.NewListNull(types.StringType),
	}
}

func NewGroupByUnknown() GroupByValue {
	return GroupByValue{
		ListValue: basetypes.NewListUnknown(types.StringType),
	}
}

func NewGroupByValue(elements []attr.Value) (GroupByValue, diag.Diagnostics) {
	listValue, diags := basetypes.NewListValue(types.StringType, elements)
	return GroupByValue{
		ListValue: listValue,
	}, diags
}
