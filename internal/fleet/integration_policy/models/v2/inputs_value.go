package v2

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// InputsValue is a custom value type for the inputs map that implements semantic equality
// Disabled inputs (enabled=false) are ignored during equality checks
type InputsValue struct {
	basetypes.MapValue
}

// Type returns an InputsType.
func (v InputsValue) Type(ctx context.Context) attr.Type {
	return NewInputsType(v.ElementType(ctx))
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
func (v InputsValue) MapSemanticEquals(ctx context.Context, newValuable basetypes.MapValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(InputsValue)
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

	// Handle null/unknown cases
	if v.IsNull() {
		return newValue.IsNull(), diags
	}

	if v.IsUnknown() {
		return newValue.IsUnknown(), diags
	}

	// Convert both maps to integrationPolicyInputsModel
	oldInputsMap := utils.MapTypeAs[integrationPolicyInputsModel](ctx, v.MapValue, path.Root("inputs"), &diags)
	if diags.HasError() {
		return false, diags
	}

	newInputsMap := utils.MapTypeAs[integrationPolicyInputsModel](ctx, newValue.MapValue, path.Root("inputs"), &diags)
	if diags.HasError() {
		return false, diags
	}

	// Filter out disabled inputs from both maps
	enabledOldInputs := filterEnabledInputs(ctx, oldInputsMap)
	enabledNewInputs := filterEnabledInputs(ctx, newInputsMap)

	// Check if the number of enabled inputs is the same
	if len(enabledOldInputs) != len(enabledNewInputs) {
		return false, diags
	}

	// Compare each enabled input
	for inputID, oldInput := range enabledOldInputs {
		newInput, exists := enabledNewInputs[inputID]
		if !exists {
			return false, diags
		}

		// Compare enabled flags
		if !oldInput.Enabled.Equal(newInput.Enabled) {
			return false, diags
		}

		// Compare vars using semantic equality if both are known
		if utils.IsKnown(oldInput.Vars) && utils.IsKnown(newInput.Vars) {
			varsEqual, d := oldInput.Vars.StringSemanticEquals(ctx, newInput.Vars)
			diags.Append(d...)
			if diags.HasError() {
				return false, diags
			}
			if !varsEqual {
				return false, diags
			}
		} else if !oldInput.Vars.Equal(newInput.Vars) {
			// If one is null/unknown, use regular equality
			return false, diags
		}

		// Compare streams
		streamsEqual, d := compareStreams(ctx, oldInput.Streams, newInput.Streams)
		diags.Append(d...)
		if diags.HasError() {
			return false, diags
		}
		if !streamsEqual {
			return false, diags
		}
	}

	return true, diags
}

// filterEnabledInputs returns a map of only the enabled inputs
func filterEnabledInputs(ctx context.Context, inputs map[string]integrationPolicyInputsModel) map[string]integrationPolicyInputsModel {
	if inputs == nil {
		return nil
	}

	enabled := make(map[string]integrationPolicyInputsModel)
	for inputID, input := range inputs {
		// Only include inputs that are explicitly enabled or unknown
		// Disabled inputs (enabled=false) are excluded
		if input.Enabled.IsNull() || input.Enabled.IsUnknown() || input.Enabled.ValueBool() {
			enabled[inputID] = input
		}
	}
	return enabled
}

// compareStreams compares two stream maps, ignoring disabled streams
func compareStreams(ctx context.Context, oldStreams, newStreams basetypes.MapValue) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Handle null/unknown cases
	if oldStreams.IsNull() && newStreams.IsNull() {
		return true, diags
	}
	if oldStreams.IsUnknown() && newStreams.IsUnknown() {
		return true, diags
	}
	if oldStreams.IsNull() != newStreams.IsNull() || oldStreams.IsUnknown() != newStreams.IsUnknown() {
		return false, diags
	}

	// Convert both maps to integrationPolicyInputStreamModel
	oldStreamsMap := utils.MapTypeAs[integrationPolicyInputStreamModel](ctx, oldStreams, path.Root("streams"), &diags)
	if diags.HasError() {
		return false, diags
	}

	newStreamsMap := utils.MapTypeAs[integrationPolicyInputStreamModel](ctx, newStreams, path.Root("streams"), &diags)
	if diags.HasError() {
		return false, diags
	}

	// Filter out disabled streams from both maps
	enabledOldStreams := filterEnabledStreams(oldStreamsMap)
	enabledNewStreams := filterEnabledStreams(newStreamsMap)

	// Check if the number of enabled streams is the same
	if len(enabledOldStreams) != len(enabledNewStreams) {
		return false, diags
	}

	// Compare each enabled stream
	for streamID, oldStream := range enabledOldStreams {
		newStream, exists := enabledNewStreams[streamID]
		if !exists {
			return false, diags
		}

		// Compare enabled flags
		if !oldStream.Enabled.Equal(newStream.Enabled) {
			return false, diags
		}

		// Compare vars using semantic equality if both are known
		if utils.IsKnown(oldStream.Vars) && utils.IsKnown(newStream.Vars) {
			varsEqual, d := oldStream.Vars.StringSemanticEquals(ctx, newStream.Vars)
			diags.Append(d...)
			if diags.HasError() {
				return false, diags
			}
			if !varsEqual {
				return false, diags
			}
		} else if !oldStream.Vars.Equal(newStream.Vars) {
			// If one is null/unknown, use regular equality
			return false, diags
		}
	}

	return true, diags
}

// filterEnabledStreams returns a map of only the enabled streams
func filterEnabledStreams(streams map[string]integrationPolicyInputStreamModel) map[string]integrationPolicyInputStreamModel {
	if streams == nil {
		return nil
	}

	enabled := make(map[string]integrationPolicyInputStreamModel)
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
func NewInputsNull(elemType attr.Type) InputsValue {
	return InputsValue{
		MapValue: basetypes.NewMapNull(elemType),
	}
}

// NewInputsUnknown creates an InputsValue with an unknown value.
func NewInputsUnknown(elemType attr.Type) InputsValue {
	return InputsValue{
		MapValue: basetypes.NewMapUnknown(elemType),
	}
}

// NewInputsValue creates an InputsValue with a known value.
func NewInputsValue(elemType attr.Type, elements map[string]attr.Value) (InputsValue, diag.Diagnostics) {
	mapValue, diags := basetypes.NewMapValue(elemType, elements)
	return InputsValue{
		MapValue: mapValue,
	}, diags
}

// NewInputsValueFrom creates an InputsValue from a map of Go values.
func NewInputsValueFrom(ctx context.Context, elemType attr.Type, elements any) (InputsValue, diag.Diagnostics) {
	mapValue, diags := basetypes.NewMapValueFrom(ctx, elemType, elements)
	return InputsValue{
		MapValue: mapValue,
	}, diags
}
