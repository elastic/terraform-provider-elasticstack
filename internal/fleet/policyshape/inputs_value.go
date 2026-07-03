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

package policyshape

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// InputsValue is a custom value type for the inputs map that implements semantic equality
// Disabled inputs (enabled=false) are ignored during equality checks
type InputsValue struct {
	basetypes.MapValue
}

// InputsConfigured reports whether inputs has at least one known, non-null
// element (i.e. the user configured the `inputs` attribute at all).
func InputsConfigured(inputs InputsValue) bool {
	return typeutils.IsKnown(inputs) && !inputs.IsNull() && len(inputs.Elements()) > 0
}

// Type returns an InputsType.
func (v InputsValue) Type(ctx context.Context) attr.Type {
	elemType := v.ElementType(ctx)
	inputType, ok := elemType.(InputType)
	if !ok {
		// Fallback for when ElementType is not InputType (shouldn't happen in practice)
		return NewInputsType(NewInputType(InputAttributeTypes()))
	}
	return NewInputsType(inputType)
}

// Equal returns true if the given value is equivalent.
func (v InputsValue) Equal(o attr.Value) bool {
	other, ok := o.(InputsValue)
	if !ok {
		return false
	}
	return v.MapValue.Equal(other.MapValue)
}

// MapSemanticEquals returns true if the given map value is semantically equal to the current map value.
// Disabled inputs (enabled=false) are ignored during the comparison.
func (v InputsValue) MapSemanticEquals(ctx context.Context, priorValuable basetypes.MapValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	priorValue, ok := priorValuable.(InputsValue)
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

	// Handle null/unknown cases
	if v.IsNull() {
		return priorValue.IsNull(), diags
	}

	if v.IsUnknown() {
		return priorValue.IsUnknown(), diags
	}

	remainingPriorInputs := priorValue.Elements()
	for inputID, oldInputValue := range v.Elements() {
		newInput := oldInputValue.(InputValue)
		priorInput, exists := remainingPriorInputs[inputID]
		if !exists {
			// If the new input is disabled, we can ignore its absence in the prior inputs
			enabled, d := newInput.MaybeEnabled(ctx)
			diags.Append(d...)
			if diags.HasError() {
				return false, diags
			}

			if !enabled {
				continue
			}

			enabledByDefault, d := newInput.EnabledByDefault(ctx)
			diags.Append(d...)
			if diags.HasError() {
				return false, diags
			}

			if enabledByDefault {
				continue
			}

			return false, diags
		}

		priorInputValue := priorInput.(InputValue)
		equals, d := newInput.ObjectSemanticEquals(ctx, priorInputValue)
		diags.Append(d...)
		if diags.HasError() {
			return false, diags
		}
		if !equals {
			return false, diags
		}

		// Remove the processed input from remainingPriorInputs
		delete(remainingPriorInputs, inputID)
	}

	// After processing all new inputs, check if there are any remaining prior inputs
	for _, priorInputValue := range remainingPriorInputs {
		priorInput := priorInputValue.(InputValue)
		// If the prior input is enabled, it's a difference
		enabled, d := priorInput.MaybeEnabled(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return false, diags
		}

		if enabled {
			return false, diags
		}
	}

	return true, diags
}

// filterEnabledStreams returns a map of only the enabled streams
func filterEnabledStreams(streams map[string]InputStreamModel) map[string]InputStreamModel {
	if streams == nil {
		return nil
	}

	enabled := make(map[string]InputStreamModel)
	for streamID, stream := range streams {
		// Only include streams that are explicitly enabled or unknown
		// Disabled streams (enabled=false) are excluded
		if stream.Enabled.IsNull() || stream.Enabled.IsUnknown() || stream.Enabled.ValueBool() {
			enabled[streamID] = stream
		}
	}
	return enabled
}

// NewInputsNull creates an InputsValue with a null value.
func NewInputsNull(elemType InputType) InputsValue {
	return InputsValue{
		MapValue: basetypes.NewMapNull(elemType),
	}
}

// NewInputsValueFrom creates an InputsValue from a map of Go values.
func NewInputsValueFrom(ctx context.Context, elemType InputType, elements any) (InputsValue, diag.Diagnostics) {
	mapValue, diags := basetypes.NewMapValueFrom(ctx, elemType, elements)
	return InputsValue{
		MapValue: mapValue,
	}, diags
}
