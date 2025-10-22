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

type valueValidator func(dependentFieldHasAllowedValue bool, dependentValueStr string, val attr.Value, p path.Path) diag.Diagnostics

// condition represents a validation rule that enforces conditional requirements
// based on the value of a dependent field. It contains the path to the field
// that this condition depends on and a list of allowed values for that field.
// When the dependent field matches one of the allowed values, additional
// validation logic can be applied to the current field.
type condition struct {
	description   func() string
	dependentPath path.Path
	allowedValues []string
	validateValue valueValidator
}

// Description describes the validation in plain text formatting.
func (v condition) Description(ctx context.Context) string {
	return v.description()
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v condition) MarkdownDescription(ctx context.Context) string {
	return v.description()
}

// isConditionMet checks if the condition is satisfied by evaluating the dependent field's value
func (v condition) isConditionMet(ctx context.Context, config tfsdk.Config) (bool, string, diag.Diagnostics) {
	var dependentValue types.String
	diags := config.GetAttribute(ctx, v.dependentPath, &dependentValue)

	if diags.HasError() {
		return false, "", diags
	}

	// If dependent value is null, unknown, or doesn't match any allowed values,
	// then the condition is not met
	dependentValueStr := dependentValue.ValueString()
	conditionMet := false

	if !dependentValue.IsNull() && !dependentValue.IsUnknown() {
		for _, allowedValue := range v.allowedValues {
			if dependentValueStr == allowedValue {
				conditionMet = true
				break
			}
		}
	}

	return conditionMet, dependentValueStr, nil
}

func (v condition) validate(ctx context.Context, config tfsdk.Config, val attr.Value, p path.Path) diag.Diagnostics {
	conditionMet, dependentValueStr, diags := v.isConditionMet(ctx, config)
	if diags.HasError() {
		return diags
	}

	return v.validateValue(conditionMet, dependentValueStr, val, p)
}

func (v condition) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	response.Diagnostics.Append(v.validate(ctx, request.Config, request.ConfigValue, request.Path)...)
}

func (v condition) ValidateList(ctx context.Context, request validator.ListRequest, response *validator.ListResponse) {
	response.Diagnostics.Append(v.validate(ctx, request.Config, request.ConfigValue, request.Path)...)
}

func (v condition) ValidateInt64(ctx context.Context, request validator.Int64Request, response *validator.Int64Response) {
	response.Diagnostics.Append(v.validate(ctx, request.Config, request.ConfigValue, request.Path)...)
}

// Assertion validations validate that the path matches one of the provided values and fail otherwise
// eg using Any to skip validation for some cases
//   stringvalidator.Any(
// 	   stringvalidator.ExactlyOneOf(path.MatchRoot("index"), path.MatchRoot("data_view_id")),
// 	   validators.StringAssert(
// 		 path.Root("type"),
// 		 []string{"machine_learning", "esql"},
// 	   ),
//   )
// ------------------------------------------------------------------------------

// StringAssert returns a validator that ensures a string attribute's value is within
// the allowedValues slice when the attribute at dependentPath meets certain conditions.
// This conditional validator is typically used to enforce in combination with other validations
// eg composing with validator.Any to skip validation for certain cases
//
// Parameters:
//   - dependentPath: The path to the attribute that this validation depends on
//   - allowedValues: The slice of string values that are considered valid for assertion
//
// Returns a validator.String that can be used in schema attribute validation.
func DependantPathOneOf(dependentPath path.Path, allowedValues []string) condition {
	return condition{
		dependentPath: dependentPath,
		allowedValues: allowedValues,
		description: func() string {
			return fmt.Sprintf("Attribute '%v' is not one of %s",
				dependentPath,
				allowedValues,
			)
		},
		validateValue: func(dependentFieldHasAllowedValue bool, dependentValueStr string, val attr.Value, p path.Path) diag.Diagnostics {
			if !dependentFieldHasAllowedValue {
				var diags diag.Diagnostics
				diags.AddAttributeError(p, "Invalid Configuration", fmt.Sprintf("Attribute '%s' is not one of %v, %s is %q",
					dependentPath,
					allowedValues,
					dependentPath,
					dependentValueStr,
				))

				return diags
			}

			return nil
		},
	}
}

// Allowance validations validate that the value is only set when the condition is met
// ----------------------------------------------------------------------------------

// StringConditionalAllowance returns a validator which ensures that a string attribute
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
//			validators.StringConditionalAllowance(
//				path.Root("auth_type"),
//				[]string{"none"},
//			),
//		},
//	},
func AllowedIfDependentPathOneOf(dependentPath path.Path, allowedValues []string) condition {
	return condition{
		dependentPath: dependentPath,
		allowedValues: allowedValues,
		description: func() string {
			if len(allowedValues) == 1 {
				return fmt.Sprintf("value can only be set when %s equals %q", dependentPath, allowedValues[0])
			}
			return fmt.Sprintf("value can only be set when %s is one of %v", dependentPath, allowedValues)
		},
		validateValue: func(dependentFieldHasAllowedValue bool, dependentValueStr string, val attr.Value, p path.Path) diag.Diagnostics {
			var diags diag.Diagnostics
			isEmpty := val.IsNull() || val.IsUnknown()
			isSet := !isEmpty

			if dependentFieldHasAllowedValue {
				return diags
			}

			if isSet {
				if len(allowedValues) == 1 {
					diags.AddAttributeError(p, "Invalid Configuration",
						fmt.Sprintf("Attribute %s can only be set when %s equals %q, but %s is %q",
							p,
							dependentPath,
							allowedValues[0],
							dependentPath,
							dependentValueStr,
						),
					)
				} else {
					diags.AddAttributeError(p, "Invalid Configuration",
						fmt.Sprintf("Attribute %s can only be set when %s is one of %v, but %s is %q",
							p,
							dependentPath,
							allowedValues,
							dependentPath,
							dependentValueStr,
						),
					)
				}
			}

			return diags
		},
	}
}

// StringConditionalAllowanceSingle is a convenience function for when there's only one allowed value.
func AllowedIfDependentPathEquals(dependentPath path.Path, requiredValue string) condition {
	return AllowedIfDependentPathOneOf(dependentPath, []string{requiredValue})
}

// Requirement validations validate that the value is set when the condition is met
// ----------------------------------------------------------------------------------

// Int64ConditionalRequirementSingle is a convenience function for when there's only one required value.
func RequiredIfDependentPathEquals(dependentPath path.Path, requiredValue string) condition {
	return RequiredIfDependentPathOneOf(dependentPath, []string{requiredValue})
}

// Int64ConditionalRequirement returns a validator that requires an int64 value to be present
// when the field at the specified dependentPath contains one of the allowedValues.
//
// Parameters:
//   - dependentPath: The path to the field whose value determines if this field is required
//   - allowedValues: A slice of string values that trigger the requirement when found in the dependent field
//
// Returns:
//   - validator.Int64: A validator that enforces the conditional requirement rule
func RequiredIfDependentPathOneOf(dependentPath path.Path, allowedValues []string) condition {
	return condition{
		dependentPath: dependentPath,
		allowedValues: allowedValues,
		description: func() string {
			if len(allowedValues) == 1 {
				return fmt.Sprintf("value required when %s equals %q", dependentPath, allowedValues[0])
			}
			return fmt.Sprintf("value required when %s is one of %v", dependentPath, allowedValues)
		},
		validateValue: func(dependentFieldHasAllowedValue bool, dependentValueStr string, val attr.Value, p path.Path) diag.Diagnostics {
			var diags diag.Diagnostics
			isEmpty := val.IsNull() || val.IsUnknown()

			if !dependentFieldHasAllowedValue {
				return diags
			}

			if isEmpty {
				diags.AddAttributeError(p, "Invalid Configuration",
					fmt.Sprintf("Attribute %s must be set when %s equals %q",
						p,
						dependentPath,
						allowedValues[0],
					),
				)
			}
			return diags
		},
	}
}

// Forbidden validate that the value is NOT set when the condition is met
// -------------------------------------------------------------------------------

// ListConditionalForbidden returns a validator that restricts a list attribute from having any values
// when a dependent attribute at the specified path contains one of the allowed values.
// When the dependent path's value matches any of the allowedValues, this validator will
// produce an error if the list attribute being validated is not empty.
//
// Parameters:
//   - dependentPath: The path to the attribute whose value determines the validation behavior
//   - allowedValues: The values that, when present in the dependent attribute, trigger the restriction
//
// Returns a List validator that enforces the conditional restriction rule.
func ForbiddenIfDependentPathOneOf(dependentPath path.Path, allowedValues []string) condition {
	return condition{
		dependentPath: dependentPath,
		allowedValues: allowedValues,
		description: func() string {
			if len(allowedValues) == 1 {
				return fmt.Sprintf("value cannot be set when %s equals %q", dependentPath, allowedValues[0])
			}
			return fmt.Sprintf("value cannot be set when %s is one of %v", dependentPath, allowedValues)
		},
		validateValue: func(dependentFieldHasAllowedValue bool, dependentValueStr string, val attr.Value, p path.Path) diag.Diagnostics {
			var diags diag.Diagnostics

			if !dependentFieldHasAllowedValue {
				return diags
			}

			isEmpty := val.IsNull() || val.IsUnknown()
			isSet := !isEmpty
			if isSet {
				diags.AddAttributeError(p, "Invalid Configuration",
					fmt.Sprintf("Attribute %s cannot be set when %s equals %q",
						p,
						dependentPath,
						allowedValues[0],
					),
				)
			}
			return diags
		},
	}
}
