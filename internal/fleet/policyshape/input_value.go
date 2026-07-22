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
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
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

// inputModelSansDefaults mirrors InputModel but omits the Defaults field, for
// InputType configurations that don't declare a `defaults` attribute at all
// (e.g. the elasticstack_fleet_managed_integration resource's InputType --
// see internal/fleet/managedintegration/schema.go's
// managedIntegrationInputAttributeTypes, which deliberately excludes `defaults`
// because managed integrations don't surface package-defaults introspection).
//
// terraform-plugin-framework's ObjectValue.As() requires an exact
// field/attribute match: a target struct field with no corresponding object
// attribute produces a "Value Conversion Error: mismatch between struct and
// object: Struct defines fields not found in object: defaults" hard error.
// InputModel unconditionally declares `defaults`, so it cannot be used to
// decode an InputValue whose object type has no such attribute -- this
// struct is the fallback for that case. See decodeInputModel, which picks
// between the two based on whether `defaults` is actually present.
type inputModelSansDefaults struct {
	Enabled   types.Bool           `tfsdk:"enabled"`
	Condition types.String         `tfsdk:"condition"`
	Vars      jsontypes.Normalized `tfsdk:"vars"`
	Streams   types.Map            `tfsdk:"streams"`
}

// toInputModel widens m into an InputModel with a null Defaults, so all of
// this file's existing defaults-aware logic (applyDefaultsToInput,
// EnabledByDefault's typeutils.IsKnown(input.Defaults) check, etc.) treats a
// schema with no `defaults` attribute exactly like one where `defaults` is
// present but was left unset -- both already correctly no-op the
// defaults-merging step.
func (m inputModelSansDefaults) toInputModel() InputModel {
	return InputModel{
		Enabled:   m.Enabled,
		Condition: m.Condition,
		Vars:      m.Vars,
		Defaults:  types.ObjectNull(InputDefaultsAttributeTypes()),
		Streams:   m.Streams,
	}
}

// decodeInputModel decodes v into an InputModel, tolerating InputType
// configurations that don't declare a `defaults` attribute (see
// inputModelSansDefaults). This is the shared entry point EnabledByDefault,
// MaybeEnabled, and ObjectSemanticEquals use instead of calling v.As(ctx,
// &InputModel{}, ...) directly, so none of them hard-fail against the
// managed integration resource's defaults-less InputType.
func decodeInputModel(ctx context.Context, v InputValue, opts basetypes.ObjectAsOptions) (InputModel, diag.Diagnostics) {
	if _, ok := v.AttributeTypes(ctx)[AttrDefaults]; !ok {
		var m inputModelSansDefaults
		diags := v.As(ctx, &m, opts)
		return m.toInputModel(), diags
	}
	var m InputModel
	diags := v.As(ctx, &m, opts)
	return m, diags
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

func (v InputValue) EnabledByDefault(ctx context.Context) (bool, diag.Diagnostics) {
	if v.IsNull() || v.IsUnknown() {
		return false, nil
	}

	input, diags := decodeInputModel(ctx, v, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return false, diags
	}

	if !typeutils.IsKnown(input.Defaults) {
		return false, diags
	}

	// Extract defaults model
	var defaults InputDefaultsModel
	d := input.Defaults.As(ctx, &defaults, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	for _, stream := range defaults.Streams {
		if typeutils.IsKnown(stream.Enabled) && stream.Enabled.ValueBool() {
			return true, nil
		}
	}

	return false, nil
}

func (v InputValue) MaybeEnabled(ctx context.Context) (bool, diag.Diagnostics) {
	if !typeutils.IsKnown(v) {
		return false, nil
	}

	input, diags := decodeInputModel(ctx, v, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return false, diags
	}

	input, defaultDiags := applyDefaultsToInput(ctx, input, input.Defaults)
	diags.Append(defaultDiags...)
	if diags.HasError() {
		return false, diags
	}

	if !typeutils.IsKnown(input.Enabled) {
		return true, diags
	}

	// The input will be treated as disabled unless at least one stream is enabled
	for _, stream := range input.Streams.Elements() {
		streamModel := InputStreamModel{}
		d := stream.(types.Object).As(ctx, &streamModel, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return false, diags
		}

		if !typeutils.IsKnown(streamModel.Enabled) || streamModel.Enabled.ValueBool() {
			return true, diags
		}
	}

	return false, diags
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
	oldInput, d := decodeInputModel(ctx, v, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	newInput, d := decodeInputModel(ctx, newValue, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	defaults := oldInput.Defaults
	if !typeutils.IsKnown(defaults) {
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

	// Condition has no package-supplied default, so a plain comparison is
	// sufficient (unlike Vars, it is never implicitly filled in).
	if !oldInputWithDefaults.Condition.Equal(newInputWithDefaults.Condition) {
		return false, diags
	}

	// Compare vars using semantic equality if both are known
	if typeutils.IsKnown(oldInputWithDefaults.Vars) && typeutils.IsKnown(newInputWithDefaults.Vars) {
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
func applyDefaultsToInput(ctx context.Context, input InputModel, defaultsObj types.Object) (InputModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	// If defaults is null or unknown, return input as-is
	if !typeutils.IsKnown(defaultsObj) {
		return input, diags
	}

	// Extract defaults model
	var defaults InputDefaultsModel
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
	if !typeutils.IsKnown(defaults) {
		return vars, nil
	}

	if !typeutils.IsKnown(vars) {
		return defaults, nil
	}

	var varsMap map[string]any
	var defaultsMap map[string]any

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
func applyDefaultsToStreams(ctx context.Context, streams basetypes.MapValue, defaultStreams map[string]InputDefaultsStreamModel) (basetypes.MapValue, diag.Diagnostics) {
	if len(defaultStreams) == 0 {
		return streams, nil
	}

	// If streams is not known, create new streams from defaults
	if !typeutils.IsKnown(streams) {
		streamsMap := make(map[string]InputStreamModel)
		for streamID, streamDefaults := range defaultStreams {
			// InputDefaultsStreamModel has no Condition field (the package
			// registry does not publish default condition expressions), so
			// build the InputStreamModel explicitly rather than converting.
			streamsMap[streamID] = InputStreamModel{
				Enabled: streamDefaults.Enabled,
				Vars:    streamDefaults.Vars,
			}
		}
		return types.MapValueFrom(ctx, StreamType(), streamsMap)
	}

	// Convert streams to model
	var diags diag.Diagnostics
	streamsMap := typeutils.MapTypeAs[InputStreamModel](ctx, streams, path.Root("streams"), &diags)
	if diags.HasError() {
		return streams, diags
	}

	// Apply defaults to each stream
	for streamID, streamDefaults := range defaultStreams {
		stream, exists := streamsMap[streamID]
		if !exists {
			// Stream not configured, use defaults
			streamsMap[streamID] = InputStreamModel{
				Enabled: streamDefaults.Enabled,
				Vars:    streamDefaults.Vars,
			}
			continue
		}

		// Apply defaults to existing stream
		if !typeutils.IsKnown(stream.Enabled) && typeutils.IsKnown(streamDefaults.Enabled) {
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

	return types.MapValueFrom(ctx, StreamType(), streamsMap)
}

// compareStreams compares two inputs' streams after defaults have been applied
func compareStreams(ctx context.Context, oldInput, newInput InputModel) (bool, diag.Diagnostics) {
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
	oldStreamsMap := typeutils.MapTypeAs[InputStreamModel](ctx, oldInput.Streams, path.Root("streams"), &diags)
	if diags.HasError() {
		return false, diags
	}

	newStreamsMap := typeutils.MapTypeAs[InputStreamModel](ctx, newInput.Streams, path.Root("streams"), &diags)
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

		if !oldStream.Condition.Equal(newStream.Condition) {
			return false, diags
		}

		// Strip server-managed keys (data_stream.*) injected by Fleet 9.5+
		// from both sides before comparing. These keys are synthesised by Fleet
		// and are not user-configurable; their presence in the API response
		// must not trigger a diff.
		oldVarsStripped, stripDiags := stripServerManagedVarsKeys(oldStream.Vars)
		diags.Append(stripDiags...)
		if diags.HasError() {
			return false, diags
		}
		newVarsStripped, stripDiags := stripServerManagedVarsKeys(newStream.Vars)
		diags.Append(stripDiags...)
		if diags.HasError() {
			return false, diags
		}

		// Compare vars using semantic equality if both are known
		if typeutils.IsKnown(oldVarsStripped) && typeutils.IsKnown(newVarsStripped) {
			varsEqual, d := oldVarsStripped.StringSemanticEquals(ctx, newVarsStripped)
			diags.Append(d...)
			if diags.HasError() {
				return false, diags
			}
			if !varsEqual {
				return false, diags
			}
		} else if !oldVarsStripped.Equal(newVarsStripped) {
			// If one is null/unknown, use regular equality
			return false, diags
		}
	}

	return true, diags
}

var serverManagedVarsKeys = []string{"data_stream.dataset", "data_stream.type"}

// stripServerManagedVarsKeys removes Fleet-injected keys from a vars map.
func stripServerManagedVarsKeys(input jsontypes.Normalized) (jsontypes.Normalized, diag.Diagnostics) {
	var diags diag.Diagnostics

	if input.IsNull() || input.IsUnknown() {
		return input, diags
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(input.ValueString()), &raw); err != nil {
		diags.AddError("Failed to parse vars JSON", err.Error())
		return input, diags
	}

	for _, key := range serverManagedVarsKeys {
		delete(raw, key)
	}

	if len(raw) == 0 {
		return jsontypes.NewNormalizedValue("{}"), diags
	}

	out, err := json.Marshal(raw)
	if err != nil {
		diags.AddError("Failed to re-marshal vars JSON", err.Error())
		return input, diags
	}

	return jsontypes.NewNormalizedValue(string(out)), diags
}
