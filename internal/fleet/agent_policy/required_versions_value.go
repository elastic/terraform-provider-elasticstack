package agent_policy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.SetValuable                   = (*RequiredVersionsValue)(nil)
	_ basetypes.SetValuableWithSemanticEquals = (*RequiredVersionsValue)(nil)
)

// RequiredVersionsValue is a custom set value that implements semantic equality based on version only.
type RequiredVersionsValue struct {
	basetypes.SetValue
}

// Type returns a RequiredVersionsType.
func (v RequiredVersionsValue) Type(ctx context.Context) attr.Type {
	return RequiredVersionsType{
		SetType: basetypes.SetType{
			ElemType: v.ElementType(ctx),
		},
	}
}

// Equal returns true if the given value is equivalent.
// This uses the standard SetValue equality which compares all fields.
func (v RequiredVersionsValue) Equal(o attr.Value) bool {
	other, ok := o.(RequiredVersionsValue)
	if !ok {
		return false
	}
	return v.SetValue.Equal(other.SetValue)
}

// SetSemanticEquals implements custom semantic equality that only compares version fields.
// This ensures that changes to percentage alone don't trigger recreation of resources.
func (v RequiredVersionsValue) SetSemanticEquals(ctx context.Context, newValuable basetypes.SetValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(RequiredVersionsValue)
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

	// Handle null and unknown cases
	if v.IsNull() {
		return newValue.IsNull(), diags
	}
	if v.IsUnknown() {
		return newValue.IsUnknown(), diags
	}

	// Extract versions from both sets
	oldVersions, d := extractVersions(ctx, v.SetValue)
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	newVersions, d := extractVersions(ctx, newValue.SetValue)
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	// Compare only the sets of versions
	if len(oldVersions) != len(newVersions) {
		return false, diags
	}

	// Check if all versions in old set exist in new set
	for version := range oldVersions {
		if !newVersions[version] {
			return false, diags
		}
	}

	return true, diags
}

// extractVersions extracts version strings from a set of required version objects
func extractVersions(ctx context.Context, setValue basetypes.SetValue) (map[string]bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	versions := make(map[string]bool)

	elements := setValue.Elements()
	for _, elem := range elements {
		obj, ok := elem.(basetypes.ObjectValue)
		if !ok {
			diags.AddError(
				"Version Extraction Error",
				fmt.Sprintf("Expected ObjectValue, got %T", elem),
			)
			continue
		}

		attrs := obj.Attributes()
		versionAttr, ok := attrs["version"]
		if !ok {
			diags.AddError(
				"Version Extraction Error",
				"Required version object missing 'version' attribute",
			)
			continue
		}

		versionStr, ok := versionAttr.(types.String)
		if !ok {
			diags.AddError(
				"Version Extraction Error",
				fmt.Sprintf("Expected String for version, got %T", versionAttr),
			)
			continue
		}

		if !versionStr.IsNull() && !versionStr.IsUnknown() {
			versions[versionStr.ValueString()] = true
		}
	}

	return versions, diags
}

// NewRequiredVersionsValueNull creates a RequiredVersionsValue with a null value.
func NewRequiredVersionsValueNull(elemType attr.Type) RequiredVersionsValue {
	return RequiredVersionsValue{
		SetValue: basetypes.NewSetNull(elemType),
	}
}

// NewRequiredVersionsValueUnknown creates a RequiredVersionsValue with an unknown value.
func NewRequiredVersionsValueUnknown(elemType attr.Type) RequiredVersionsValue {
	return RequiredVersionsValue{
		SetValue: basetypes.NewSetUnknown(elemType),
	}
}

// NewRequiredVersionsValue creates a RequiredVersionsValue with a known value.
func NewRequiredVersionsValue(elemType attr.Type, elements []attr.Value) (RequiredVersionsValue, diag.Diagnostics) {
	setValue, diags := basetypes.NewSetValue(elemType, elements)
	return RequiredVersionsValue{
		SetValue: setValue,
	}, diags
}

// NewRequiredVersionsValueMust creates a RequiredVersionsValue with a known value and panics on error.
func NewRequiredVersionsValueMust(elemType attr.Type, elements []attr.Value) RequiredVersionsValue {
	return RequiredVersionsValue{
		SetValue: basetypes.NewSetValueMust(elemType, elements),
	}
}
