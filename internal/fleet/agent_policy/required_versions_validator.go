package agent_policy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Set = (*uniqueVersionValidator)(nil)

// uniqueVersionValidator validates that all required_versions have unique version strings
type uniqueVersionValidator struct{}

func (v uniqueVersionValidator) Description(ctx context.Context) string {
	return "Ensures that all required_versions entries have unique version strings"
}

func (v uniqueVersionValidator) MarkdownDescription(ctx context.Context) string {
	return "Ensures that all required_versions entries have unique version strings"
}

func (v uniqueVersionValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	// If the value is unknown or null, validation should not be performed
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	elements := req.ConfigValue.Elements()
	versions := make(map[string]bool)

	for _, elem := range elements {
		obj, ok := elem.(types.Object)
		if !ok {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid Type",
				fmt.Sprintf("Expected object type in set, got %T", elem),
			)
			continue
		}

		attrs := obj.Attributes()
		versionAttr, ok := attrs["version"]
		if !ok {
			continue
		}

		versionStr, ok := versionAttr.(types.String)
		if !ok {
			continue
		}

		if versionStr.IsNull() || versionStr.IsUnknown() {
			continue
		}

		version := versionStr.ValueString()
		if versions[version] {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Duplicate Version",
				fmt.Sprintf("The version '%s' appears multiple times in required_versions. Each version must be unique.", version),
			)
			return
		}
		versions[version] = true
	}
}

// UniqueVersions returns a validator that ensures all required_versions have unique version strings
func UniqueVersions() validator.Set {
	return uniqueVersionValidator{}
}
