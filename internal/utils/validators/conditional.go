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

// condition represents a validation rule that enforces conditional requirements
// based on the value of a dependent field. It contains the path to the field
// that this condition depends on and a list of allowed values for that field.
// When the dependent field matches one of the allowed values, additional
// validation logic can be applied to the current field.
type condition struct {
	dependentPath path.Path
	allowedValues []string
}

// conditionalValidation represents a validation rule that applies different constraints
// based on a specified condition. It controls whether a value is required, allowed,
// or forbidden when the condition is met. This struct is used to implement
// conditional validation logic where the validation behavior depends on the
// evaluation of a particular condition.
type conditionalValidation struct {
	condition      condition
	valueRequired  bool
	valueAllowed   bool
	valueForbidden bool
}

// Description describes the validation in plain text formatting.
func (v conditionalValidation) Description(_ context.Context) string {
	if v.valueForbidden {
		return fmt.Sprintf("Value cannot be set when %s equals %q",
			v.condition.dependentPath,
			v.condition.allowedValues[0],
		)
	} else if v.valueRequired {
		return fmt.Sprintf("Value must be set when %s equals %q",
			v.condition.dependentPath,
			v.condition.allowedValues[0],
		)
	} else if v.valueAllowed {
		if len(v.condition.allowedValues) == 1 {
			return fmt.Sprintf("value can only be set when %s equals %q", v.condition.dependentPath, v.condition.allowedValues[0])
		}
		return fmt.Sprintf("value can only be set when %s is one of %v", v.condition.dependentPath, v.condition.allowedValues)
	} else {
		return fmt.Sprintf("Unknown validation for %s equals %q",
			v.condition.dependentPath,
			v.condition.allowedValues[0],
		)
	}
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v conditionalValidation) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Description describes the validation in plain text formatting.
func (v condition) Description(_ context.Context) string {
	return fmt.Sprintf("Attribute '%v' is not one of %s",
		v.dependentPath,
		v.allowedValues,
	)
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v condition) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
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

	if !conditionMet {
		diags.AddAttributeError(p, "Invalid Configuration", fmt.Sprintf("Attribute '%s' is not one of %v, %s is %q",
			v.dependentPath,
			v.allowedValues,
			v.dependentPath,
			dependentValueStr,
		))

		return diags

	}
	return nil
}

func (v conditionalValidation) validateValueForbidden(conditionMet bool, dependentValueStr string, val attr.Value, p path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if !conditionMet {
		return diags
	}

	isEmpty := val.IsNull() || val.IsUnknown()
	isSet := !isEmpty
	if isSet {
		diags.AddAttributeError(p, "Invalid Configuration",
			fmt.Sprintf("Attribute %s cannot be set when %s equals %q",
				p,
				v.condition.dependentPath,
				v.condition.allowedValues[0],
			),
		)
	}
	return diags
}

func (v conditionalValidation) validateValueRequired(conditionMet bool, dependentValueStr string, val attr.Value, p path.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	isEmpty := val.IsNull() || val.IsUnknown()

	if !conditionMet {
		return diags
	}

	if isEmpty {
		diags.AddAttributeError(p, "Invalid Configuration",
			fmt.Sprintf("Attribute %s must be set when %s equals %q",
				p,
				v.condition.dependentPath,
				v.condition.allowedValues[0],
			),
		)
	}
	return diags
}

func (v conditionalValidation) validateValueAllowed(conditionMet bool, dependentValueStr string, val attr.Value, p path.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	isEmpty := val.IsNull() || val.IsUnknown()
	isSet := !isEmpty

	if conditionMet {
		return diags
	}

	if isSet {
		if len(v.condition.allowedValues) == 1 {
			diags.AddAttributeError(p, "Invalid Configuration",
				fmt.Sprintf("Attribute %s can only be set when %s equals %q, but %s is %q",
					p,
					v.condition.dependentPath,
					v.condition.allowedValues[0],
					v.condition.dependentPath,
					dependentValueStr,
				),
			)
		} else {
			diags.AddAttributeError(p, "Invalid Configuration",
				fmt.Sprintf("Attribute %s can only be set when %s is one of %v, but %s is %q",
					p,
					v.condition.dependentPath,
					v.condition.allowedValues,
					v.condition.dependentPath,
					dependentValueStr,
				),
			)
		}
	}

	return diags
}

func (v conditionalValidation) validate(ctx context.Context, config tfsdk.Config, val attr.Value, p path.Path) diag.Diagnostics {
	conditionMet, dependentValueStr, diags := v.condition.isConditionMet(ctx, config)
	if diags.HasError() {
		return diags
	}

	if v.valueForbidden {
		return v.validateValueForbidden(conditionMet, dependentValueStr, val, p)
	} else if v.valueRequired {
		return v.validateValueRequired(conditionMet, dependentValueStr, val, p)
	} else if v.valueAllowed {
		return v.validateValueAllowed(conditionMet, dependentValueStr, val, p)
	}

	return nil
}

// validateConditionalRequirement was an attempt at shared logic but is not used
// The validation logic is implemented directly in ValidateString and ValidateFloat64 methods

// ValidateString performs the validation for string attributes.
func (v conditionalValidation) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	response.Diagnostics.Append(v.validate(ctx, request.Config, request.ConfigValue, request.Path)...)
}

// ValidateInt64 performs the validation for int64 attributes.
func (v conditionalValidation) ValidateInt64(ctx context.Context, request validator.Int64Request, response *validator.Int64Response) {
	response.Diagnostics.Append(v.validate(ctx, request.Config, request.ConfigValue, request.Path)...)
}

// ValidateInt64 performs the validation for List attributes.
func (v conditionalValidation) ValidateList(ctx context.Context, request validator.ListRequest, response *validator.ListResponse) {
	response.Diagnostics.Append(v.validate(ctx, request.Config, request.ConfigValue, request.Path)...)
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
func StringAssert(dependentPath path.Path, allowedValues []string) validator.String {
	return condition{
		dependentPath: dependentPath,
		allowedValues: allowedValues,
	}
}

// ListAssert creates a list validator that conditionally validates list values based on another attribute.
// It returns a validator that checks if the list contains only values from the allowedValues slice
// when the attribute at dependentPath meets certain conditions.
//
// Parameters:
//   - dependentPath: The path to the attribute that this validator depends on
//   - allowedValues: A slice of strings representing the valid values for the list
//
// Returns a validator.List that can be used to validate list attributes conditionally.
func ListAssert(dependentPath path.Path, allowedValues []string) validator.List {
	return condition{
		dependentPath: dependentPath,
		allowedValues: allowedValues,
	}
}

// Int64Assert creates a conditional validator for int64 values that checks if the current
// attribute is valid based on the value of another attribute specified by dependentPath.
// The validation passes when the dependent attribute's value is one of the allowedValues.
// This is useful for implementing conditional validation logic where an int64 field
// should only be validated or have certain constraints when another field has specific values.
//
// Parameters:
//   - dependentPath: The path to the attribute whose value determines if validation should occur
//   - allowedValues: A slice of string values that the dependent attribute must match for validation to pass
//
// Returns a validator.Int64 that can be used in Terraform schema validation.
func Int64Assert(dependentPath path.Path, allowedValues []string) validator.Int64 {
	return condition{
		dependentPath: dependentPath,
		allowedValues: allowedValues,
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
func StringConditionalAllowance(dependentPath path.Path, allowedValues []string) validator.String {
	return conditionalValidation{
		condition: condition{
			dependentPath: dependentPath,
			allowedValues: allowedValues,
		},
		valueAllowed: true,
	}
}

// StringConditionalAllowanceSingle is a convenience function for when there's only one allowed value.
func StringConditionalAllowanceSingle(dependentPath path.Path, requiredValue string) validator.String {
	return StringConditionalAllowance(dependentPath, []string{requiredValue})
}

// Int64ConditionalAllowanceSingle is a convenience function for when there's only one allowed value.
func Int64ConditionalAllowanceSingle(dependentPath path.Path, requiredValue string) validator.Int64 {
	return conditionalValidation{
		condition: condition{
			dependentPath: dependentPath,
			allowedValues: []string{requiredValue},
		},
		valueAllowed: true,
	}
}

// Int64ConditionalAllowance returns a validator which ensures that an int64 attribute
// can only be set if another attribute at the specified path equals one of the specified values.
//
// The dependentPath parameter should use path.Root() to specify the attribute path.
// For example: path.Root("auth_type")
//
// Example usage:
//
//	"connection_timeout": schema.Int64Attribute{
//		Optional: true,
//		Validators: []validator.Int64{
//			validators.Int64ConditionalAllowance(
//				path.Root("auth_type"),
//				[]string{"basic", "oauth"},
//			),
//		},
//	},
func Int64ConditionalAllowance(dependentPath path.Path, requiredValues []string) validator.Int64 {
	return conditionalValidation{
		condition: condition{
			dependentPath: dependentPath,
			allowedValues: requiredValues,
		},
		valueAllowed: true,
	}
}

// Requirement validations validate that the value is set when the condition is met
// ----------------------------------------------------------------------------------

// Int64ConditionalRequirementSingle is a convenience function for when there's only one required value.
func Int64ConditionalRequirementSingle(dependentPath path.Path, requiredValue string) validator.Int64 {
	return conditionalValidation{
		condition: condition{
			dependentPath: dependentPath,
			allowedValues: []string{requiredValue},
		},
		valueRequired: true,
	}
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
func Int64ConditionalRequirement(dependentPath path.Path, allowedValues []string) validator.Int64 {
	return conditionalValidation{
		condition: condition{
			dependentPath: dependentPath,
			allowedValues: allowedValues,
		},
		valueRequired: true,
	}
}

// StringConditionalRequirement returns a validator that requires a string value to be present
// when the field at the specified dependentPath contains one of the allowedValues.
//
// Parameters:
//   - dependentPath: The path to the field whose value determines if this field is required
//   - allowedValues: A slice of string values that trigger the requirement when found in the dependent field
//
// Returns:
//   - validator.String: A validator that enforces the conditional requirement rule
func StringConditionalRequirement(dependentPath path.Path, allowedValues []string) validator.String {
	return conditionalValidation{
		condition: condition{
			dependentPath: dependentPath,
			allowedValues: allowedValues,
		},
		valueRequired: true,
	}
}

// StringConditionalRequirementSingle creates a string validator that requires this field to be set
// when the dependent field at dependentPath equals the specified requiredValue.
// This is a convenience wrapper around StringConditionalRequirement for single value conditions.
func StringConditionalRequirementSingle(dependentPath path.Path, requiredValue string) validator.String {
	return StringConditionalRequirement(dependentPath, []string{requiredValue})
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
func ListConditionalForbidden(dependentPath path.Path, allowedValues []string) validator.List {
	return conditionalValidation{
		condition: condition{
			dependentPath: dependentPath,
			allowedValues: allowedValues,
		},
		valueForbidden: true,
	}
}

// StringConditionalForbidden returns a string validator that restricts the current attribute
// from being set when the dependent attribute at dependentPath contains one of the
// allowedValues. This validator enforces that certain string attributes are forbidden
// when specific conditions are met on related attributes.
//
// Parameters:
//   - dependentPath: The path to the attribute whose value determines the restriction
//   - allowedValues: List of values that, when present in the dependent attribute,
//     will forbid setting the current attribute
//
// Returns a validator.String that can be used in schema validation.
func StringConditionalForbidden(dependentPath path.Path, allowedValues []string) validator.String {
	return conditionalValidation{
		condition: condition{
			dependentPath: dependentPath,
			allowedValues: allowedValues,
		},
		valueForbidden: true,
	}
}

func Int64ConditionalForbidden(dependentPath path.Path, allowedValues []string) validator.Int64 {
	return conditionalValidation{
		condition: condition{
			dependentPath: dependentPath,
			allowedValues: allowedValues,
		},
		valueForbidden: true,
	}
}
