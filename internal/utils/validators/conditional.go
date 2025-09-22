package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// conditionalRequirement represents a validator which ensures that an attribute
// can only be set if another attribute at a specified path equals one of the specified values.
// This is a shared implementation that can be used for both string and float64 validators.
type conditionalRequirement struct {
	dependentPath path.Path
	allowedValues []string
}

// Description describes the validation in plain text formatting.
func (v conditionalRequirement) Description(_ context.Context) string {
	if len(v.allowedValues) == 1 {
		return fmt.Sprintf("value can only be set when %s equals %q", v.dependentPath, v.allowedValues[0])
	}
	return fmt.Sprintf("value can only be set when %s is one of %v", v.dependentPath, v.allowedValues)
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v conditionalRequirement) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v conditionalRequirement) validate(ctx context.Context, config tfsdk.Config, val attr.Value, p path.Path) diag.Diagnostics {
	if val.IsNull() || val.IsUnknown() {
		return nil
	}

	// Get the value at the dependent path
	var dependentValue types.String
	diags := config.GetAttribute(ctx, v.dependentPath, &dependentValue)
	if diags.HasError() {
		return diags
	}

	// If dependent value is null, unknown, or doesn't match any allowed values,
	// then the current attribute should not be set
	dependentValueStr := dependentValue.ValueString()
	isAllowed := false

	if !dependentValue.IsNull() && !dependentValue.IsUnknown() {
		for _, allowedValue := range v.allowedValues {
			if dependentValueStr == allowedValue {
				isAllowed = true
				break
			}
		}
	}

	if !isAllowed {
		if len(v.allowedValues) == 1 {
			diags.AddAttributeError(p, "Invalid Configuration",
				fmt.Sprintf("Attribute %s can only be set when %s equals %q, but %s is %q",
					p,
					v.dependentPath,
					v.allowedValues[0],
					v.dependentPath,
					dependentValueStr,
				),
			)
			return diags
		} else {
			diags.AddAttributeError(p, "Invalid Configuration",
				fmt.Sprintf("Attribute %s can only be set when %s is one of %v, but %s is %q",
					p,
					v.dependentPath,
					v.allowedValues,
					v.dependentPath,
					dependentValueStr,
				),
			)
			return diags
		}
	}

	return nil
}

// validateConditionalRequirement was an attempt at shared logic but is not used
// The validation logic is implemented directly in ValidateString and ValidateFloat64 methods

// ValidateString performs the validation for string attributes.
func (v conditionalRequirement) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	response.Diagnostics.Append(v.validate(ctx, request.Config, request.ConfigValue, request.Path)...)
}

// ValidateInt64 performs the validation for int64 attributes.
func (v conditionalRequirement) ValidateInt64(ctx context.Context, request validator.Int64Request, response *validator.Int64Response) {
	response.Diagnostics.Append(v.validate(ctx, request.Config, request.ConfigValue, request.Path)...)
}

// StringConditionalRequirement returns a validator which ensures that a string attribute
// can only be set if another attribute at the specified path equals one of the specified values.
//
// The dependentPath parameter should use path.Root() to specify the attribute path.
// For example: path.Root("auth_type")
//
// Example usage:
//
//	"connection_type": schema.StringAttribute{
//		Optional: true,
//		Validators: []validator.String{
//			validators.StringConditionalRequirement(
//				path.Root("auth_type"),
//				[]string{"none"},
//				"connection_type can only be set when auth_type is 'none'",
//			),
//		},
//	},
func StringConditionalRequirement(dependentPath path.Path, allowedValues []string) validator.String {
	return conditionalRequirement{
		dependentPath: dependentPath,
		allowedValues: allowedValues,
	}
}

// StringConditionalRequirementSingle is a convenience function for when there's only one allowed value.
func StringConditionalRequirementSingle(dependentPath path.Path, requiredValue string) validator.String {
	return StringConditionalRequirement(dependentPath, []string{requiredValue})
}

func Int64ConditionalRequirement(dependentPath path.Path, allowedValues []string) validator.Int64 {
	return conditionalRequirement{
		dependentPath: dependentPath,
		allowedValues: allowedValues,
	}
}

// Int64ConditionalRequirementSingle is a convenience function for when there's only one allowed value.
func Int64ConditionalRequirementSingle(dependentPath path.Path, requiredValue string) validator.Int64 {
	return Int64ConditionalRequirement(dependentPath, []string{requiredValue})
}
