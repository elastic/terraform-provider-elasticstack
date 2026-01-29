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
// based on the value of a dependent field. It contains either a static path or
// a path expression to the field that this condition depends on, and a list of
// allowed values for that field.
// When the dependent field matches one of the allowed values, additional
// validation logic can be applied to the current field.
// Use dependentPath for absolute paths, or dependentPathExpression for relative paths.
type condition struct {
	description             func() string
	dependentPath           *path.Path
	dependentPathExpression *path.Expression
	allowedValues           []string
	validateValue           valueValidator
}

// Description describes the validation in plain text formatting.
func (v condition) Description(ctx context.Context) string {
	return v.description()
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v condition) MarkdownDescription(ctx context.Context) string {
	return v.description()
}

// dependentFieldHasAllowedValue checks if the dependent field specified by the condition's
// dependentPath or dependentPathExpression has a value that matches one of the allowed values
// defined in the condition. It retrieves the dependent field's value from the provided
// configuration context and compares it against the condition's allowedValues slice.
//
// The method returns three values:
//   - bool: true if the dependent field has a non-null, non-unknown value that matches
//     one of the allowed values; false otherwise
//   - string: the string representation of the dependent field's current value
//   - diag.Diagnostics: any diagnostics encountered while retrieving the field value
//
// If the dependent field is null, unknown, or its value doesn't match any of the
// allowed values, the condition is considered not met and the method returns false.
func (v condition) dependentFieldHasAllowedValue(ctx context.Context, config tfsdk.Config, currentPath path.Path) (bool, string, diag.Diagnostics) {
	var dependentValue types.String
	var diags diag.Diagnostics

	// Determine which path to use
	if v.dependentPathExpression != nil {
		// Merge the path expression with the current path to resolve relative references
		merged := currentPath.Expression().Merge(*v.dependentPathExpression)
		matchedPaths, matchDiags := config.PathMatches(ctx, merged)
		diags.Append(matchDiags...)
		if diags.HasError() {
			return false, "", diags
		}

		// For validation purposes, we expect exactly one match
		if len(matchedPaths) == 0 {
			// No match found, condition not met
			return false, "", nil
		}

		// Use the first matched path
		diags.Append(config.GetAttribute(ctx, matchedPaths[0], &dependentValue)...)
	} else {
		// Use static path
		diags.Append(config.GetAttribute(ctx, *v.dependentPath, &dependentValue)...)
	}

	if diags.HasError() {
		return false, "", diags
	}

	dependentValueStr := dependentValue.ValueString()
	dependentFieldHasAllowedValue := false

	if !dependentValue.IsNull() && !dependentValue.IsUnknown() {
		for _, allowedValue := range v.allowedValues {
			if dependentValueStr == allowedValue {
				dependentFieldHasAllowedValue = true
				break
			}
		}
	}

	return dependentFieldHasAllowedValue, dependentValueStr, nil
}

func (v condition) validate(ctx context.Context, config tfsdk.Config, val attr.Value, p path.Path) diag.Diagnostics {
	dependentFieldHasAllowedValue, dependentValueStr, diags := v.dependentFieldHasAllowedValue(ctx, config, p)
	if diags.HasError() {
		return diags
	}

	return v.validateValue(dependentFieldHasAllowedValue, dependentValueStr, val, p)
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

func (v condition) ValidateObject(ctx context.Context, request validator.ObjectRequest, response *validator.ObjectResponse) {
	response.Diagnostics.Append(v.validate(ctx, request.Config, request.ConfigValue, request.Path)...)
}

// DependantPathOneOf creates a condition that validates a dependent path's value is one of the allowed values.
// It returns a condition that checks if the value at dependentPath matches any of the provided allowedValues.
// If the dependent field does not have an allowed value, it generates a diagnostic error indicating
// which values are permitted and what the current value is.
//
// Parameters:
//   - dependentPath: The path to the attribute that must have one of the allowed values
//   - allowedValues: A slice of strings representing the valid values for the dependent path
//
// Returns:
//   - condition: A condition struct that can be used for validation
func DependantPathOneOf(dependentPath path.Path, allowedValues []string) condition {
	return condition{
		dependentPath: &dependentPath,
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

// AllowedIfDependentPathOneOf creates a validation condition that allows the current attribute
// to be set only when a dependent attribute at the specified path has one of the allowed values.
//
// Parameters:
//   - dependentPath: The path to the attribute that this validation depends on
//   - allowedValues: A slice of string values that the dependent attribute must match
//
// Returns:
//   - condition: A validation condition that can be used with conditional validators
//
// Example:
//
//	// Only allow "ssl_cert" to be set when "protocol" is "https"
//	AllowedIfDependentPathOneOf(path.Root("protocol"), []string{"https"})
func AllowedIfDependentPathOneOf(dependentPath path.Path, allowedValues []string) condition {
	return condition{
		dependentPath: &dependentPath,
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

// AllowedIfDependentPathEquals returns a condition that allows a field to be set
// only if the value at the specified dependent path equals the required value.
// This is a convenience function that wraps AllowedIfDependentPathOneOf with a
// single value slice.
//
// Parameters:
//   - dependentPath: The path to the field whose value determines if this field is allowed
//   - requiredValue: The exact string value that the dependent field must equal
//
// Returns:
//   - condition: A validation condition that enforces the dependency rule
func AllowedIfDependentPathEquals(dependentPath path.Path, requiredValue string) condition {
	return AllowedIfDependentPathOneOf(dependentPath, []string{requiredValue})
}

// RequiredIfDependentPathEquals returns a condition that makes a field required
// when the value at the specified dependent path equals the given required value.
// This is a convenience function that wraps RequiredIfDependentPathOneOf with
// a single value slice.
//
// Parameters:
//   - dependentPath: The path to the field whose value will be checked
//   - requiredValue: The value that, when present at dependentPath, makes this field required
//
// Returns:
//   - condition: A validation condition function
func RequiredIfDependentPathEquals(dependentPath path.Path, requiredValue string) condition {
	return RequiredIfDependentPathOneOf(dependentPath, []string{requiredValue})
}

// RequiredIfDependentPathOneOf returns a condition that validates an attribute is required
// when a dependent attribute's value matches one of the specified allowed values.
//
// The condition checks if the dependent attribute (specified by dependentPath) has a value
// that is present in the allowedValues slice. If the dependent attribute matches any of
// the allowed values, then the attribute being validated must not be null or unknown.
//
// Parameters:
//   - dependentPath: The path to the attribute whose value determines the requirement
//   - allowedValues: A slice of string values that trigger the requirement when matched
//
// Returns:
//   - condition: A validation condition that enforces the requirement rule
//
// Example usage:
//
//	validator := RequiredIfDependentPathOneOf(
//	  path.Root("type"),
//	  []string{"custom", "advanced"},
//	)
//	// This would require the current attribute when "type" equals "custom" or "advanced"
func RequiredIfDependentPathOneOf(dependentPath path.Path, allowedValues []string) condition {
	return condition{
		dependentPath: &dependentPath,
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

// ForbiddenIfDependentPathOneOf creates a validation condition that forbids setting a value
// when a dependent field matches one of the specified allowed values.
//
// This validator is useful for creating mutually exclusive configuration scenarios where
// certain attributes should not be set when another attribute has specific values.
//
// Parameters:
//   - dependentPath: The path to the field whose value determines the validation behavior
//   - allowedValues: A slice of string values that, when matched by the dependent field,
//     will trigger the forbidden condition
//
// Returns:
//   - condition: A validation condition that will generate an error if the current field
//     is set while the dependent field matches any of the allowed values
//
// Example usage:
//
//	validator := ForbiddenIfDependentPathOneOf(
//	  path.Root("type"),
//	  []string{"basic", "simple"},
//	)
//	// This will prevent setting the current attribute when "type" equals "basic" or "simple"
func ForbiddenIfDependentPathOneOf(dependentPath path.Path, allowedValues []string) condition {
	return condition{
		dependentPath: &dependentPath,
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

// AllowedIfDependentPathExpressionOneOf creates a validation condition that allows the current attribute
// to be set only when a dependent attribute matched by the path expression has one of the allowed values.
// This uses a relative path expression that is resolved relative to the field being validated.
//
// Parameters:
//   - dependentPathExpression: The path expression to match the dependent attribute relative to the current field
//   - allowedValues: A slice of string values that the dependent attribute must match
//
// Returns:
//   - condition: A validation condition that can be used with conditional validators
//
// Example:
//
//	// Only allow "ssl_cert" to be set when a sibling "protocol" field is "https"
//	AllowedIfDependentPathExpressionOneOf(path.MatchRelative().AtParent().AtName("protocol"), []string{"https"})
func AllowedIfDependentPathExpressionOneOf(dependentPathExpression path.Expression, allowedValues []string) condition {
	descStr := "dependent field"
	return condition{
		dependentPathExpression: &dependentPathExpression,
		allowedValues:           allowedValues,
		description: func() string {
			if len(allowedValues) == 1 {
				return fmt.Sprintf("value can only be set when %s equals %q", descStr, allowedValues[0])
			}
			return fmt.Sprintf("value can only be set when %s is one of %v", descStr, allowedValues)
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
							descStr,
							allowedValues[0],
							descStr,
							dependentValueStr,
						),
					)
				} else {
					diags.AddAttributeError(p, "Invalid Configuration",
						fmt.Sprintf("Attribute %s can only be set when %s is one of %v, but %s is %q",
							p,
							descStr,
							allowedValues,
							descStr,
							dependentValueStr,
						),
					)
				}
			}

			return diags
		},
	}
}

// RequiredIfDependentPathExpressionOneOf returns a condition that validates an attribute is required
// when a dependent attribute matched by the path expression has a value matching one of the specified
// allowed values. This uses a relative path expression that is resolved relative to the field being validated.
//
// Parameters:
//   - dependentPathExpression: The path expression to match the dependent attribute relative to the current field
//   - allowedValues: A slice of string values that trigger the requirement when matched
//
// Returns:
//   - condition: A validation condition that enforces the requirement rule
//
// Example usage:
//
//	validator := RequiredIfDependentPathExpressionOneOf(
//	  path.MatchRelative().AtParent().AtName("type"),
//	  []string{"custom", "advanced"},
//	)
//	// This would require the current attribute when sibling "type" equals "custom" or "advanced"
func RequiredIfDependentPathExpressionOneOf(dependentPathExpression path.Expression, allowedValues []string) condition {
	descStr := "dependent field"
	return condition{
		dependentPathExpression: &dependentPathExpression,
		allowedValues:           allowedValues,
		description: func() string {
			if len(allowedValues) == 1 {
				return fmt.Sprintf("value required when %s equals %q", descStr, allowedValues[0])
			}
			return fmt.Sprintf("value required when %s is one of %v", descStr, allowedValues)
		},
		validateValue: func(dependentFieldHasAllowedValue bool, dependentValueStr string, val attr.Value, p path.Path) diag.Diagnostics {
			var diags diag.Diagnostics
			isEmpty := val.IsNull() || val.IsUnknown()

			if !dependentFieldHasAllowedValue {
				return diags
			}

			if isEmpty {
				var msg string
				if len(allowedValues) == 1 {
					msg = fmt.Sprintf(
						"Attribute %s must be set when %s equals %q",
						p,
						descStr,
						allowedValues[0],
					)
				} else {
					msg = fmt.Sprintf(
						"Attribute %s must be set when %s is one of %v",
						p,
						descStr,
						allowedValues,
					)
				}

				diags.AddAttributeError(
					p,
					"Invalid Configuration",
					msg,
				)
			}
			return diags
		},
	}
}

// ForbiddenIfDependentPathExpressionOneOf creates a validation condition that forbids setting a value
// when a dependent field matched by the path expression has one of the specified allowed values.
// This uses a relative path expression that is resolved relative to the field being validated.
//
// Parameters:
//   - dependentPathExpression: The path expression to match the dependent attribute relative to the current field
//   - allowedValues: A slice of string values that, when matched by the dependent field, will trigger the forbidden condition
//
// Returns:
//   - condition: A validation condition that will generate an error if the current field is set
//     while the dependent field matches any of the allowed values
//
// Example usage:
//
//	validator := ForbiddenIfDependentPathExpressionOneOf(
//	  path.MatchRelative().AtParent().AtName("type"),
//	  []string{"basic", "simple"},
//	)
//	// This will prevent setting the current attribute when sibling "type" equals "basic" or "simple"
func ForbiddenIfDependentPathExpressionOneOf(dependentPathExpression path.Expression, allowedValues []string) condition {
	descStr := "dependent field"
	return condition{
		dependentPathExpression: &dependentPathExpression,
		allowedValues:           allowedValues,
		description: func() string {
			if len(allowedValues) == 1 {
				return fmt.Sprintf("value cannot be set when %s equals %q", descStr, allowedValues[0])
			}
			return fmt.Sprintf("value cannot be set when %s is one of %v", descStr, allowedValues)
		},
		validateValue: func(dependentFieldHasAllowedValue bool, dependentValueStr string, val attr.Value, p path.Path) diag.Diagnostics {
			var diags diag.Diagnostics

			if !dependentFieldHasAllowedValue {
				return diags
			}

			isEmpty := val.IsNull() || val.IsUnknown()
			isSet := !isEmpty
			if isSet {
				var msg string
				if len(allowedValues) == 1 {
					msg = fmt.Sprintf(
						"Attribute %s cannot be set when %s equals %q",
						p,
						descStr,
						allowedValues[0],
					)
				} else {
					msg = fmt.Sprintf(
						"Attribute %s cannot be set when %s is one of %v",
						p,
						descStr,
						allowedValues,
					)
				}

				diags.AddAttributeError(
					p,
					"Invalid Configuration",
					msg,
				)
			}
			return diags
		},
	}
}
