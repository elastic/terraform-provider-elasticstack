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

// AliasObjectValue is the value type for a template alias. It implements
// [basetypes.ObjectValuableWithSemanticEquals] so routing-only and API echo shapes match
// design.md §2 during plan, refresh, and post-apply checks.
type AliasObjectValue struct {
	basetypes.ObjectValue
}

// ObjectSemanticEquals compares this value to newValuable using the alias routing predicate
// (design.md §2). The framework calls priorValue.ObjectSemanticEquals(ctx, newValue) during
// post-create/update/read semantic normalization; see terraform-plugin-framework fwschemadata.
func (v AliasObjectValue) ObjectSemanticEquals(ctx context.Context, newValuable basetypes.ObjectValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(AliasObjectValue)
	if !ok {
		diags.AddError(
			"Semantic equality check error",
			"An unexpected value type was received while comparing template alias values. "+
				"Please report this to the provider developers.\n\n"+
				"Expected type: AliasObjectValue\n"+
				"Got type: "+fmt.Sprintf("%T", newValuable),
		)
		return false, diags
	}

	// For framework calls, v is the prior value and newValue is the new (proposed/refreshed) value.
	// The asymmetric helpers in aliasElementModelsSemanticallyEqual encode "prior == config side,
	// incoming == API/refreshed side", so we run forward and reverse to cover both directions.
	if v.IsNull() {
		return newValue.IsNull(), diags
	}

	if v.IsUnknown() {
		return newValue.IsUnknown(), diags
	}

	if newValue.IsNull() || newValue.IsUnknown() {
		return false, diags
	}

	var a AliasElementModel
	d := v.As(ctx, &a, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty: true,
	})
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	var b AliasElementModel
	d = newValue.As(ctx, &b, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty: true,
	})
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	aFilled := fillUnknownAliasModelFieldsFromOther(a, b)
	okForward, d := aliasElementModelsSemanticallyEqual(ctx, aFilled, b)
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}
	if okForward {
		return true, diags
	}

	bFilled := fillUnknownAliasModelFieldsFromOther(b, a)
	okReverse, d := aliasElementModelsSemanticallyEqual(ctx, bFilled, a)
	diags.Append(d...)
	return okReverse, diags
}

// Type returns an AliasObjectType.
func (v AliasObjectValue) Type(_ context.Context) attr.Type {
	return NewAliasObjectType()
}

// Equal returns true if the given value is equivalent (strict object equality).
//
// Do not delegate to ObjectSemanticEquals here: Terraform correlates planned vs actual set
// elements using Equal during apply consistency checks; semantic equality would hide real
// mismatches and trigger "planned set element not correlate with any element in actual".
func (v AliasObjectValue) Equal(o attr.Value) bool {
	other, ok := o.(AliasObjectValue)
	if !ok {
		return false
	}
	return v.ObjectValue.Equal(other.ObjectValue)
}

func fillUnknownAliasModelFieldsFromOther(m, other AliasElementModel) AliasElementModel {
	out := m
	if m.IndexRouting.IsUnknown() {
		out.IndexRouting = other.IndexRouting
	}
	if m.SearchRouting.IsUnknown() {
		out.SearchRouting = other.SearchRouting
	}
	if m.Routing.IsUnknown() {
		out.Routing = other.Routing
	}
	if m.IsHidden.IsUnknown() {
		out.IsHidden = other.IsHidden
	}
	if m.IsWriteIndex.IsUnknown() {
		out.IsWriteIndex = other.IsWriteIndex
	}
	if m.Filter.IsUnknown() {
		out.Filter = other.Filter
	}
	return out
}

// aliasElementModelsSemanticallyEqual is the directed comparison: prior is the configuration or older
// state side; incoming is the API/refreshed side for the rules in design.md §2.
func aliasElementModelsSemanticallyEqual(ctx context.Context, prior, incoming AliasElementModel) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !prior.Name.Equal(incoming.Name) ||
		!aliasMainRoutingSemanticallyEqual(
			prior.Routing, prior.IndexRouting, prior.SearchRouting,
			incoming.Routing, incoming.IndexRouting, incoming.SearchRouting,
		) ||
		!aliasOptionalBoolSemanticEqual(prior.IsHidden, incoming.IsHidden) ||
		!aliasOptionalBoolSemanticEqual(prior.IsWriteIndex, incoming.IsWriteIndex) {
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

	if !routingFieldSemanticallyEqual(prior.IndexRouting, incoming.IndexRouting, incoming.Routing) &&
		!aliasEsIndexRoutingEchoesPriorMainRouting(prior, incoming) {
		return false, diags
	}

	if !routingFieldSemanticallyEqual(prior.SearchRouting, incoming.SearchRouting, incoming.Routing) {
		return false, diags
	}

	return true, diags
}

// aliasOptionalBoolSemanticEqual mirrors SDKv2 Optional+Default(false): null, unknown, and explicit false
// are equivalent for template.alias booleans; both must agree when either side is explicitly true.
func aliasOptionalBoolSemanticEqual(a, b types.Bool) bool {
	if a.Equal(b) {
		return true
	}
	aUnset := a.IsNull() || a.IsUnknown()
	bUnset := b.IsNull() || b.IsUnknown()
	if aUnset && bUnset {
		return true
	}
	aFalse := aUnset || (!a.IsNull() && !a.IsUnknown() && !a.ValueBool())
	bFalse := bUnset || (!b.IsNull() && !b.IsUnknown() && !b.ValueBool())
	return aFalse && bFalse
}

// aliasRoutingFieldStringsSemanticallyEqual treats null and "" as equivalent (matching how routing
// fields are compared elsewhere in this file). Unknown values are not semantically equal here because
// callers handle unknown precedence at the model level.
func aliasRoutingFieldStringsSemanticallyEqual(a, b types.String) bool {
	if a.IsUnknown() || b.IsUnknown() {
		return a.Equal(b)
	}
	aEmpty := a.IsNull() || a.ValueString() == ""
	bEmpty := b.IsNull() || b.ValueString() == ""
	if aEmpty != bEmpty {
		return false
	}
	if aEmpty {
		return true
	}
	return a.ValueString() == b.ValueString()
}

// aliasEsIndexRoutingEchoesPriorMainRouting handles GET responses where Elasticsearch omits routing and
// sets index_routing to the configured generic routing value even when index_routing was distinct in the template.
func aliasEsIndexRoutingEchoesPriorMainRouting(prior, incoming AliasElementModel) bool {
	if !incoming.Routing.IsNull() && incoming.Routing.ValueString() != "" {
		return false
	}
	pr := ""
	if !prior.Routing.IsNull() && !prior.Routing.IsUnknown() {
		pr = prior.Routing.ValueString()
	}
	if pr == "" {
		return false
	}
	incIdx := ""
	if !incoming.IndexRouting.IsNull() && !incoming.IndexRouting.IsUnknown() {
		incIdx = incoming.IndexRouting.ValueString()
	}
	if incIdx != pr {
		return false
	}
	if !aliasRoutingFieldStringsSemanticallyEqual(prior.SearchRouting, incoming.SearchRouting) {
		return false
	}
	pi := ""
	if !prior.IndexRouting.IsNull() && !prior.IndexRouting.IsUnknown() {
		pi = prior.IndexRouting.ValueString()
	}
	if pi == incIdx {
		return false
	}
	return true
}

// AliasElementModel is the Terraform struct for one template alias (expand/flatten and semantic equality).
type AliasElementModel struct {
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

// aliasMainRoutingSemanticallyEqual compares the top-level routing field. Elasticsearch often
// omits `routing` on GET when index_routing/search_routing are set; it may drop `routing` entirely
// when those two differ (only index_routing and search_routing are returned).
func aliasMainRoutingSemanticallyEqual(
	priorRouting, priorIndex, priorSearch,
	incomingRouting, incomingIndex, incomingSearch types.String,
) bool {
	if priorRouting.Equal(incomingRouting) {
		return true
	}
	incomingREmpty := incomingRouting.IsNull() || incomingRouting.ValueString() == ""
	if !incomingREmpty {
		return false
	}
	// Elasticsearch may echo the generic routing value into index_routing on GET and omit routing
	// when all three routing fields were set in the template (observed on 8.x).
	prStr := ""
	if !priorRouting.IsNull() && !priorRouting.IsUnknown() {
		prStr = priorRouting.ValueString()
	}
	if prStr != "" && incomingSearch.Equal(priorSearch) {
		incIdx := ""
		if !incomingIndex.IsNull() && !incomingIndex.IsUnknown() {
			incIdx = incomingIndex.ValueString()
		}
		if incIdx == prStr {
			piStr := ""
			if !priorIndex.IsNull() && !priorIndex.IsUnknown() {
				piStr = priorIndex.ValueString()
			}
			if piStr != incIdx {
				return true
			}
		}
	}
	// API omitted routing: index/search unchanged from prior.
	if incomingIndex.Equal(priorIndex) && incomingSearch.Equal(priorSearch) {
		if !priorIndex.Equal(priorSearch) {
			return true
		}
		// index_routing == search_routing on prior (including both empty): echo of main routing
		if priorRouting.IsNull() || priorRouting.ValueString() == "" {
			return true
		}
		pr := priorRouting.ValueString()
		if incomingIndex.IsNull() || incomingSearch.IsNull() {
			return false
		}
		return incomingIndex.ValueString() == pr && incomingSearch.ValueString() == pr
	}
	// Routing-only config: prior left index/search at default ""; API echoes into index/search.
	priorIRUnset := priorIndex.IsNull() || priorIndex.ValueString() == ""
	priorSRUnset := priorSearch.IsNull() || priorSearch.ValueString() == ""
	if !priorIRUnset || !priorSRUnset {
		return false
	}
	if priorRouting.IsNull() || priorRouting.ValueString() == "" {
		return true
	}
	pr := priorRouting.ValueString()
	incIdx := ""
	if !incomingIndex.IsNull() && !incomingIndex.IsUnknown() {
		incIdx = incomingIndex.ValueString()
	}
	incSrch := ""
	if !incomingSearch.IsNull() && !incomingSearch.IsUnknown() {
		incSrch = incomingSearch.ValueString()
	}
	if incIdx == "" && incSrch == "" {
		return false
	}
	idxEq := incIdx == pr
	srchEq := incSrch == pr
	if idxEq && srchEq {
		return true
	}
	// Echo into index_routing only or search_routing only; the other field still matches prior (typically "").
	if idxEq && (incSrch == "" || incomingSearch.Equal(priorSearch)) {
		return true
	}
	if srchEq && (incIdx == "" || incomingIndex.Equal(priorIndex)) {
		return true
	}
	return false
}

// routingFieldSemanticallyEqual encodes:
//
//	v.field ≡ new.field  ⇔  v.field == new.field
//	  OR (v.field is null/empty AND new.field == effectiveNewRouting AND effectiveNewRouting != "")
//
// effectiveNewRouting prefers new.routing; when the API leaves routing empty, it falls back to
// newField so index_routing/search_routing echoes still match routing-only configuration.
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

	newR := ""
	if !newRouting.IsNull() {
		newR = newRouting.ValueString()
	}
	if newR == "" && !newField.IsNull() {
		newR = newField.ValueString()
	}

	newEmpty := newField.IsNull() || newField.ValueString() == ""
	if newR == "" {
		return newEmpty
	}
	if newEmpty {
		// Configuration omitted index_routing/search_routing (null in plan); state may use "" or echo
		// the generic routing value only on routing — treat as equivalent to the echo case.
		return true
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
