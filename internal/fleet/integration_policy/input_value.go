package integration_policy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// InputValue is a custom value type for an individual input that implements semantic equality
// Semantic equality uses the defaults attribute to populate unspecified values before comparison
type InputValue struct {
	basetypes.ObjectValue
}

// Type returns an InputType.
func (v InputValue) Type(ctx context.Context) attr.Type {
	return NewInputType(v.AttributeTypes(ctx))
}

// Equal returns true if the given value is equivalent.
func (v InputValue) Equal(o attr.Value) bool {
	other, ok := o.(InputValue)
	if !ok {
		return false
	}
	return v.ObjectValue.Equal(other.ObjectValue)
}

func (v InputValue) MaybeEnabled(ctx context.Context) (bool, diag.Diagnostics) {
	if v.IsNull() || v.IsUnknown() {
		return false, nil
	}

	var input integrationPolicyInputsModel
	diags := v.As(ctx, &input, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return false, diags
	}

	input, defaultDiags := applyDefaultsToInput(ctx, input, input.Defaults)
	diags.Append(defaultDiags...)
	if diags.HasError() {
		return false, diags
	}

	if !utils.IsKnown(input.Enabled) {
		return true, diags
	}

	// The input will be treated as disabled unless at least one stream is enabled
	for _, stream := range input.Streams.Elements() {
		streamModel := integrationPolicyInputStreamModel{}
		d := stream.(types.Object).As(ctx, &streamModel, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return false, diags
		}

		if !utils.IsKnown(streamModel.Enabled) || streamModel.Enabled.ValueBool() {
			return true, diags
		}
	}

	return false, nil
}

// ObjectSemanticEquals returns true if the given object value is semantically equal to the current object value.
// Semantic equality applies defaults from the defaults attribute before comparing values.
func (v InputValue) ObjectSemanticEquals(ctx context.Context, newValuable basetypes.ObjectValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(InputValue)
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

	// Convert both values to the model
	var oldInput integrationPolicyInputsModel
	d := v.As(ctx, &oldInput, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	var newInput integrationPolicyInputsModel
	d = newValue.As(ctx, &newInput, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	defaults := oldInput.Defaults
	if !utils.IsKnown(defaults) {
		defaults = newInput.Defaults
	}

	// Apply defaults to both inputs
	oldInputWithDefaults, d := applyDefaultsToInput(ctx, oldInput, defaults)
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	newInputWithDefaults, d := applyDefaultsToInput(ctx, newInput, defaults)
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	// Ignore the disabled attribute in equality checks.
	// Disabled inputs are handled at the InputsValue level.

	// Compare vars using semantic equality if both are known
	if utils.IsKnown(oldInputWithDefaults.Vars) && utils.IsKnown(newInputWithDefaults.Vars) {
		varsEqual, d := oldInputWithDefaults.Vars.StringSemanticEquals(ctx, newInputWithDefaults.Vars)
		diags.Append(d...)
		if diags.HasError() {
			return false, diags
		}
		if !varsEqual {
			return false, diags
		}
	} else if !oldInputWithDefaults.Vars.Equal(newInputWithDefaults.Vars) {
		// If one is null/unknown, use regular equality
		return false, diags
	}

	// Compare streams
	streamsEqual, d := compareStreams(ctx, oldInputWithDefaults, newInputWithDefaults)
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}
	if !streamsEqual {
		return false, diags
	}

	return true, diags
}

// applyDefaultsToInput applies defaults from the defaults attribute to the input
func applyDefaultsToInput(ctx context.Context, input integrationPolicyInputsModel, defaultsObj types.Object) (integrationPolicyInputsModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	// If defaults is null or unknown, return input as-is
	if !utils.IsKnown(defaultsObj) {
		return input, diags
	}

	// Extract defaults model
	var defaults inputDefaultsModel
	d := defaultsObj.As(ctx, &defaults, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return input, diags
	}

	result := input

	// Apply var defaults
	varsWithDefaults, d := applyDefaultsToVars(input.Vars, defaults.Vars)
	diags.Append(d...)
	if diags.HasError() {
		return input, diags
	}
	result.Vars = varsWithDefaults

	// Apply stream defaults
	streamsWithDefaults, d := applyDefaultsToStreams(ctx, input.Streams, defaults.Streams)
	diags.Append(d...)
	if diags.HasError() {
		return input, diags
	}
	result.Streams = streamsWithDefaults

	return result, diags
}

func applyDefaultsToVars(vars jsontypes.Normalized, defaults jsontypes.Normalized) (jsontypes.Normalized, diag.Diagnostics) {
	if !utils.IsKnown(defaults) {
		return vars, nil
	}

	if !utils.IsKnown(vars) {
		return defaults, nil
	}

	var varsMap map[string]interface{}
	var defaultsMap map[string]interface{}

	diags := vars.Unmarshal(&varsMap)
	d := defaults.Unmarshal(&defaultsMap)
	diags.Append(d...)
	if diags.HasError() {
		return vars, diags
	}

	for key, defaultValue := range defaultsMap {
		if _, exists := varsMap[key]; !exists {
			varsMap[key] = defaultValue
		}
	}

	varsBytes, err := json.Marshal(varsMap)
	if err != nil {
		diags.AddError("Failed to marshal vars with defaults", err.Error())
		return vars, diags
	}

	varsWithDefaults := jsontypes.NewNormalizedValue(string(varsBytes))
	return varsWithDefaults, diags
}

// applyDefaultsToStreams applies defaults to streams
func applyDefaultsToStreams(ctx context.Context, streams basetypes.MapValue, defaultStreams map[string]inputDefaultsStreamModel) (basetypes.MapValue, diag.Diagnostics) {
	if len(defaultStreams) == 0 {
		return streams, nil
	}

	// If streams is not known, create new streams from defaults
	if !utils.IsKnown(streams) {
		streamsMap := make(map[string]integrationPolicyInputStreamModel)
		for streamID, streamDefaults := range defaultStreams {
			streamsMap[streamID] = integrationPolicyInputStreamModel(streamDefaults)
		}
		return types.MapValueFrom(ctx, getInputStreamType(), streamsMap)
	}

	// Convert streams to model
	var diags diag.Diagnostics
	streamsMap := utils.MapTypeAs[integrationPolicyInputStreamModel](ctx, streams, path.Root("streams"), &diags)
	if diags.HasError() {
		return streams, diags
	}

	// Apply defaults to each stream
	for streamID, streamDefaults := range defaultStreams {
		stream, exists := streamsMap[streamID]
		if !exists {
			// Stream not configured, use defaults
			streamsMap[streamID] = integrationPolicyInputStreamModel(streamDefaults)
			continue
		}

		// Apply defaults to existing stream
		if !utils.IsKnown(stream.Enabled) && utils.IsKnown(streamDefaults.Enabled) {
			stream.Enabled = streamDefaults.Enabled
		}
		varsWithDefaults, d := applyDefaultsToVars(stream.Vars, streamDefaults.Vars)
		diags.Append(d...)
		if diags.HasError() {
			return streams, diags
		}
		stream.Vars = varsWithDefaults
		streamsMap[streamID] = stream
	}

	return types.MapValueFrom(ctx, getInputStreamType(), streamsMap)
}

// compareStreams compares two inputs' streams after defaults have been applied
func compareStreams(ctx context.Context, oldInput, newInput integrationPolicyInputsModel) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Handle null/unknown cases
	if oldInput.Streams.IsNull() && newInput.Streams.IsNull() {
		return true, diags
	}
	if oldInput.Streams.IsUnknown() && newInput.Streams.IsUnknown() {
		return true, diags
	}
	if oldInput.Streams.IsNull() != newInput.Streams.IsNull() || oldInput.Streams.IsUnknown() != newInput.Streams.IsUnknown() {
		return false, diags
	}

	// Convert both maps to model
	oldStreamsMap := utils.MapTypeAs[integrationPolicyInputStreamModel](ctx, oldInput.Streams, path.Root("streams"), &diags)
	if diags.HasError() {
		return false, diags
	}

	newStreamsMap := utils.MapTypeAs[integrationPolicyInputStreamModel](ctx, newInput.Streams, path.Root("streams"), &diags)
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

// NewInputNull creates an InputValue with a null value.
func NewInputNull(attrTypes map[string]attr.Type) InputValue {
	return InputValue{
		ObjectValue: basetypes.NewObjectNull(attrTypes),
	}
}

// NewInputUnknown creates an InputValue with an unknown value.
func NewInputUnknown(attrTypes map[string]attr.Type) InputValue {
	return InputValue{
		ObjectValue: basetypes.NewObjectUnknown(attrTypes),
	}
}

// NewInputValue creates an InputValue with a known value.
func NewInputValue(attrTypes map[string]attr.Type, attributes map[string]attr.Value) (InputValue, diag.Diagnostics) {
	objectValue, diags := basetypes.NewObjectValue(attrTypes, attributes)
	return InputValue{
		ObjectValue: objectValue,
	}, diags
}

// NewInputValueFrom creates an InputValue from a Go value.
func NewInputValueFrom(ctx context.Context, attrTypes map[string]attr.Type, val any) (InputValue, diag.Diagnostics) {
	objectValue, diags := basetypes.NewObjectValueFrom(ctx, attrTypes, val)
	return InputValue{
		ObjectValue: objectValue,
	}, diags
}
