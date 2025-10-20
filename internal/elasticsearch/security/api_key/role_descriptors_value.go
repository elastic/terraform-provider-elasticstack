package api_key

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.StringValuable                   = (*RoleDescriptorsValue)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*RoleDescriptorsValue)(nil)
	_ xattr.ValidateableAttribute                = (*RoleDescriptorsValue)(nil)
)

type RoleDescriptorsValue struct {
	jsontypes.Normalized
}

// Type returns a RoleDescriptorsType.
func (v RoleDescriptorsValue) Type(_ context.Context) attr.Type {
	return RoleDescriptorsType{}
}

func (v RoleDescriptorsValue) WithDefaults() (RoleDescriptorsValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	if v.IsNull() {
		return v, diags
	}

	if v.IsUnknown() {
		return v, diags
	}

	var parsedValue map[string]models.ApiKeyRoleDescriptor
	err := json.Unmarshal([]byte(v.ValueString()), &parsedValue)
	if err != nil {
		diags.AddError("Failed to unmarshal role descriptors value", err.Error())
		return RoleDescriptorsValue{}, diags
	}

	for role, descriptor := range parsedValue {
		for i, index := range descriptor.Indices {
			if index.AllowRestrictedIndices == nil {
				descriptor.Indices[i].AllowRestrictedIndices = new(bool)
				*descriptor.Indices[i].AllowRestrictedIndices = false
			}
		}
		parsedValue[role] = descriptor
	}

	valueWithDefaults, err := json.Marshal(parsedValue)
	if err != nil {
		diags.AddError("Failed to marshal sanitized config value", err.Error())
		return RoleDescriptorsValue{}, diags
	}

	return NewRoleDescriptorsValue(string(valueWithDefaults)), diags
}

// StringSemanticEquals returns true if the given config object value is semantically equal to the current config object value.
// The comparison will ignore any default values present in one value, but unset in the other.
func (v RoleDescriptorsValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(RoleDescriptorsValue)
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

	thisWithDefaults, d := v.WithDefaults()
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	thatWithDefaults, d := newValue.WithDefaults()
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	return thisWithDefaults.Normalized.StringSemanticEquals(ctx, thatWithDefaults.Normalized)
}

// NewRoleDescriptorsNull creates a RoleDescriptorsValue with a null value. Determine whether the value is null via IsNull method.
func NewRoleDescriptorsNull() RoleDescriptorsValue {
	return RoleDescriptorsValue{
		Normalized: jsontypes.NewNormalizedNull(),
	}
}

// NewRoleDescriptorsUnknown creates a RoleDescriptorsValue with an unknown value. Determine whether the value is unknown via IsUnknown method.
func NewRoleDescriptorsUnknown() RoleDescriptorsValue {
	return RoleDescriptorsValue{
		Normalized: jsontypes.NewNormalizedUnknown(),
	}
}

// NewRoleDescriptorsValue creates a RoleDescriptorsValue with a known value. Access the value via ValueString method.
func NewRoleDescriptorsValue(value string) RoleDescriptorsValue {
	if value == "" {
		return NewRoleDescriptorsNull()
	}

	return RoleDescriptorsValue{
		Normalized: jsontypes.NewNormalizedValue(value),
	}
}
