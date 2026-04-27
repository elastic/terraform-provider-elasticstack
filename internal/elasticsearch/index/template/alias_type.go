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

package template

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.ObjectTypable                    = (*AliasObjectType)(nil)
	_ basetypes.ObjectValuable                   = (*AliasObjectValue)(nil)
	_ basetypes.ObjectValuableWithSemanticEquals = (*AliasObjectValue)(nil)
)

// AliasAttributeTypes returns attribute types for a single template alias block element.
func AliasAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":           types.StringType,
		"index_routing":  types.StringType,
		"routing":        types.StringType,
		"search_routing": types.StringType,
		"filter":         jsontypes.NormalizedType{},
		"is_hidden":      types.BoolType,
		"is_write_index": types.BoolType,
	}
}

// AliasObjectType is the Terraform type for a template alias nested block element.
type AliasObjectType struct {
	basetypes.ObjectType
}

// NewAliasObjectType constructs an AliasObjectType with the standard alias schema.
func NewAliasObjectType() AliasObjectType {
	return AliasObjectType{
		ObjectType: basetypes.ObjectType{
			AttrTypes: AliasAttributeTypes(),
		},
	}
}

// String returns a human readable string of the type name.
func (t AliasObjectType) String() string {
	return "template.AliasObjectType"
}

// ValueType returns the Value type.
func (t AliasObjectType) ValueType(_ context.Context) attr.Value {
	return AliasObjectValue{
		ObjectValue: basetypes.NewObjectUnknown(t.AttributeTypes()),
	}
}

// Equal returns true if the given type is equivalent.
func (t AliasObjectType) Equal(o attr.Type) bool {
	other, ok := o.(AliasObjectType)
	if !ok {
		return false
	}
	return t.ObjectType.Equal(other.ObjectType)
}

// ValueFromObject returns an ObjectValuable type given a basetypes.ObjectValue.
func (t AliasObjectType) ValueFromObject(_ context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	return AliasObjectValue{ObjectValue: in}, nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.
func (t AliasObjectType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.ObjectType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	objectValue, ok := attrValue.(basetypes.ObjectValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	return AliasObjectValue{ObjectValue: objectValue}, nil
}

// AliasObjectValue is the value type for a template alias with routing-aware semantic equality.
type AliasObjectValue struct {
	basetypes.ObjectValue
}

// Type returns an AliasObjectType.
func (v AliasObjectValue) Type(ctx context.Context) attr.Type {
	return NewAliasObjectType()
}

// Equal returns true if the given value is equivalent (strict object equality).
func (v AliasObjectValue) Equal(o attr.Value) bool {
	other, ok := o.(AliasObjectValue)
	if !ok {
		return false
	}
	return v.ObjectValue.Equal(other.ObjectValue)
}

// ObjectSemanticEquals applies the strict alias routing predicate from design.md §2.
// Receiver v is the prior/state value; newValuable is the new value (plan or refreshed).
func (v AliasObjectValue) ObjectSemanticEquals(ctx context.Context, newValuable basetypes.ObjectValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(AliasObjectValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)
		return false, diags
	}

	if v.IsNull() {
		return newValue.IsNull(), diags
	}

	if v.IsUnknown() {
		return newValue.IsUnknown(), diags
	}

	if newValue.IsNull() || newValue.IsUnknown() {
		return false, diags
	}

	var prior aliasObjectModel
	d := v.As(ctx, &prior, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	var incoming aliasObjectModel
	d = newValue.As(ctx, &incoming, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	if !prior.Name.Equal(incoming.Name) ||
		!prior.Routing.Equal(incoming.Routing) ||
		!prior.IsHidden.Equal(incoming.IsHidden) ||
		!prior.IsWriteIndex.Equal(incoming.IsWriteIndex) {
		return false, diags
	}

	filterEqual, d := aliasFiltersSemanticallyEqual(ctx, prior.Filter, incoming.Filter)
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}
	if !filterEqual {
		return false, diags
	}

	if !routingFieldSemanticallyEqual(prior.IndexRouting, incoming.IndexRouting, incoming.Routing) {
		return false, diags
	}

	if !routingFieldSemanticallyEqual(prior.SearchRouting, incoming.SearchRouting, incoming.Routing) {
		return false, diags
	}

	return true, diags
}

type aliasObjectModel struct {
	Name          types.String         `tfsdk:"name"`
	IndexRouting  types.String         `tfsdk:"index_routing"`
	Routing       types.String         `tfsdk:"routing"`
	SearchRouting types.String         `tfsdk:"search_routing"`
	Filter        jsontypes.Normalized `tfsdk:"filter"`
	IsHidden      types.Bool           `tfsdk:"is_hidden"`
	IsWriteIndex  types.Bool           `tfsdk:"is_write_index"`
}

func aliasFiltersSemanticallyEqual(ctx context.Context, a, b jsontypes.Normalized) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	if a.IsNull() && b.IsNull() {
		return true, diags
	}
	if a.IsUnknown() && b.IsUnknown() {
		return true, diags
	}
	if a.IsNull() != b.IsNull() || a.IsUnknown() || b.IsUnknown() {
		return false, diags
	}

	eq, d := a.StringSemanticEquals(ctx, b)
	diags.Append(d...)
	return eq, diags
}

// routingFieldSemanticallyEqual encodes:
//
//	v.field ≡ new.field  ⇔  v.field == new.field
//	  OR (v.field is null/empty AND new.field == new.routing AND new.routing != "")
func routingFieldSemanticallyEqual(priorField, newField, newRouting types.String) bool {
	if priorField.IsUnknown() || newField.IsUnknown() || newRouting.IsUnknown() {
		return false
	}

	if priorField.Equal(newField) {
		return true
	}

	priorUnset := priorField.IsNull() || priorField.ValueString() == ""
	if !priorUnset {
		return false
	}

	if newRouting.IsNull() {
		return false
	}

	newR := newRouting.ValueString()
	if newR == "" {
		return false
	}

	if newField.IsNull() {
		return false
	}

	return newField.ValueString() == newR
}

// NewAliasObjectNull creates a null alias object value.
func NewAliasObjectNull() AliasObjectValue {
	return AliasObjectValue{
		ObjectValue: basetypes.NewObjectNull(AliasAttributeTypes()),
	}
}

// NewAliasObjectUnknown creates an unknown alias object value.
func NewAliasObjectUnknown() AliasObjectValue {
	return AliasObjectValue{
		ObjectValue: basetypes.NewObjectUnknown(AliasAttributeTypes()),
	}
}

// NewAliasObjectValue constructs a known alias object from attribute values.
func NewAliasObjectValue(attrs map[string]attr.Value) (AliasObjectValue, diag.Diagnostics) {
	obj, diags := basetypes.NewObjectValue(AliasAttributeTypes(), attrs)
	return AliasObjectValue{ObjectValue: obj}, diags
}
