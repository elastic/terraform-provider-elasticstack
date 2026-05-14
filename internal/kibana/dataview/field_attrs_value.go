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

package dataview

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// FieldAttrsValue is a custom value type for field_attrs that implements semantic equality,
// suppressing server-generated count-only entries and count drift when count is unset in config.
type FieldAttrsValue struct {
	basetypes.MapValue
}

// Type returns a FieldAttrsType.
func (v FieldAttrsValue) Type(ctx context.Context) attr.Type {
	elemType := v.ElementType(ctx)
	if elemType == nil {
		return NewFieldAttrsType(getFieldAttrElemType())
	}
	return NewFieldAttrsType(elemType)
}

// Equal returns true if the given value is equivalent.
func (v FieldAttrsValue) Equal(o attr.Value) bool {
	other, ok := o.(FieldAttrsValue)
	if !ok {
		return false
	}
	return v.MapValue.Equal(other.MapValue)
}

// MapSemanticEquals compares the config-derived map to prior state, ignoring server-only count
// noise per REQ-015 (fix-dataview-field-attrs-drift).
func (v FieldAttrsValue) MapSemanticEquals(ctx context.Context, priorValuable basetypes.MapValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	priorValue, ok := priorValuable.(FieldAttrsValue)
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
		// Only null↔null is equal here. An unknown prior value (e.g. early plan rounds with
		// computed dependencies) is intentionally treated as a real change so the framework
		// continues planning for it; mirrors InputsValue.MapSemanticEquals.
		if priorValue.IsNull() {
			return true, diags
		}
		if priorValue.IsUnknown() {
			return false, diags
		}
		for _, priorAttr := range priorValue.Elements() {
			priorOV, ok := priorAttr.(basetypes.ObjectValue)
			if !ok {
				diags.AddError(
					"Semantic Equality Check Error",
					"Expected basetypes.ObjectValue for field_attrs entry, got "+fmt.Sprintf("%T", priorAttr),
				)
				return false, diags
			}
			var priorModel fieldAttrModel
			diags.Append(priorOV.As(ctx, &priorModel, basetypes.ObjectAsOptions{})...)
			if diags.HasError() {
				return false, diags
			}
			if !priorModel.CustomLabel.IsNull() {
				return false, diags
			}
		}
		return true, diags
	}

	if v.IsUnknown() {
		return priorValue.IsUnknown(), diags
	}

	newElems := v.Elements()
	for fieldName, newAttrValue := range newElems {
		newOV, ok := newAttrValue.(basetypes.ObjectValue)
		if !ok {
			diags.AddError(
				"Semantic Equality Check Error",
				"Expected basetypes.ObjectValue for field_attrs entry "+fieldName+", got "+fmt.Sprintf("%T", newAttrValue),
			)
			return false, diags
		}

		var newModel fieldAttrModel
		diags.Append(newOV.As(ctx, &newModel, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return false, diags
		}

		priorAttrValue, exists := priorValue.Elements()[fieldName]
		if !exists {
			return false, diags
		}

		priorOV, ok := priorAttrValue.(basetypes.ObjectValue)
		if !ok {
			diags.AddError(
				"Semantic Equality Check Error",
				"Expected basetypes.ObjectValue for field_attrs entry "+fieldName+", got "+fmt.Sprintf("%T", priorAttrValue),
			)
			return false, diags
		}

		var priorModel fieldAttrModel
		diags.Append(priorOV.As(ctx, &priorModel, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return false, diags
		}

		if !newModel.CustomLabel.Equal(priorModel.CustomLabel) {
			return false, diags
		}

		if !newModel.Count.IsNull() {
			if priorModel.Count.IsNull() || !newModel.Count.Equal(priorModel.Count) {
				return false, diags
			}
		}
	}

	for fieldName, priorAttrValue := range priorValue.Elements() {
		if _, inNew := newElems[fieldName]; inNew {
			continue
		}

		priorOV, ok := priorAttrValue.(basetypes.ObjectValue)
		if !ok {
			diags.AddError(
				"Semantic Equality Check Error",
				"Expected basetypes.ObjectValue for field_attrs entry "+fieldName+", got "+fmt.Sprintf("%T", priorAttrValue),
			)
			return false, diags
		}

		var priorModel fieldAttrModel
		diags.Append(priorOV.As(ctx, &priorModel, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return false, diags
		}

		if !priorModel.CustomLabel.IsNull() {
			return false, diags
		}
	}

	return true, diags
}

// NewFieldAttrsNull creates a FieldAttrsValue with a null value.
func NewFieldAttrsNull(elemType attr.Type) FieldAttrsValue {
	return FieldAttrsValue{
		MapValue: basetypes.NewMapNull(elemType),
	}
}

// NewFieldAttrsUnknown creates a FieldAttrsValue with an unknown value.
func NewFieldAttrsUnknown(elemType attr.Type) FieldAttrsValue {
	return FieldAttrsValue{
		MapValue: basetypes.NewMapUnknown(elemType),
	}
}

// NewFieldAttrsValue creates a FieldAttrsValue with a known value.
func NewFieldAttrsValue(elemType attr.Type, elements map[string]attr.Value) (FieldAttrsValue, diag.Diagnostics) {
	mapValue, diags := basetypes.NewMapValue(elemType, elements)
	return FieldAttrsValue{
		MapValue: mapValue,
	}, diags
}

// NewFieldAttrsValueFrom creates a FieldAttrsValue from a map of Go values.
func NewFieldAttrsValueFrom(ctx context.Context, elemType attr.Type, elements any) (FieldAttrsValue, diag.Diagnostics) {
	mapValue, diags := basetypes.NewMapValueFrom(ctx, elemType, elements)
	return FieldAttrsValue{
		MapValue: mapValue,
	}, diags
}
