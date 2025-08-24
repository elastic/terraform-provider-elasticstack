package api_key

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var (
	_ validator.String = requiresTypeValidator{}
	_ validator.Object = requiresTypeValidator{}
)

// requiresTypeValidator validates that a string attribute is only provided
// when the resource has a specific value for the "type" attribute.
type requiresTypeValidator struct {
	expectedType string
}

// RequiresType returns a validator which ensures that the configured attribute
// is only provided when the "type" attribute matches the expected value.
func RequiresType(expectedType string) requiresTypeValidator {
	return requiresTypeValidator{
		expectedType: expectedType,
	}
}

func (validator requiresTypeValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Ensures that the attribute is only provided when type=%s", validator.expectedType)
}

func (validator requiresTypeValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

// validateType contains the common validation logic for both string and object validators
func (validator requiresTypeValidator) validateType(ctx context.Context, config tfsdk.Config, attrPath path.Path, diagnostics *diag.Diagnostics) bool {
	// Get the type attribute value from the same configuration object
	var typeAttr *string
	diags := config.GetAttribute(ctx, path.Root("type"), &typeAttr)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return false
	}

	// If type is unknown or empty, we can't validate
	if typeAttr == nil {
		return true
	}

	// Check if the current type matches the expected type
	if *typeAttr != validator.expectedType {
		diagnostics.AddAttributeError(
			attrPath,
			fmt.Sprintf("Attribute not valid for API key type '%s'", *typeAttr),
			fmt.Sprintf("The %s attribute can only be used when type='%s', but type='%s' was specified.",
				attrPath.String(), validator.expectedType, *typeAttr),
		)
		return false
	}

	return true
}

func (validator requiresTypeValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	// If the attribute is null or unknown, there's nothing to validate
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	validator.validateType(ctx, req.Config, req.Path, &resp.Diagnostics)
}

func (validator requiresTypeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If the attribute is null or unknown, there's nothing to validate
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	validator.validateType(ctx, req.Config, req.Path, &resp.Diagnostics)
}
