package apikey

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

// requiresType returns a validator which ensures that the configured attribute
// is only provided when the "type" attribute matches the expected value.
func requiresType(expectedType string) requiresTypeValidator {
	return requiresTypeValidator{
		expectedType: expectedType,
	}
}

func (v requiresTypeValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Ensures that the attribute is only provided when type=%s", v.expectedType)
}

func (v requiresTypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// validateType contains the common validation logic for both string and object validators
func (v requiresTypeValidator) validateType(ctx context.Context, config tfsdk.Config, attrPath path.Path, diagnostics *diag.Diagnostics) {
	// Get the type attribute value from the same configuration object
	var typeAttr *string
	diags := config.GetAttribute(ctx, path.Root("type"), &typeAttr)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return
	}

	// If type is unknown or empty, we can't validate
	if typeAttr == nil {
		return
	}

	// Check if the current type matches the expected type
	if *typeAttr != v.expectedType {
		diagnostics.AddAttributeError(
			attrPath,
			fmt.Sprintf("Attribute not valid for API key type '%s'", *typeAttr),
			fmt.Sprintf("The %s attribute can only be used when type='%s', but type='%s' was specified.",
				attrPath.String(), v.expectedType, *typeAttr),
		)
		return
	}
}

func (v requiresTypeValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	// If the attribute is null or unknown, there's nothing to validate
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	v.validateType(ctx, req.Config, req.Path, &resp.Diagnostics)
}

func (v requiresTypeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If the attribute is null or unknown, there's nothing to validate
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	v.validateType(ctx, req.Config, req.Path, &resp.Diagnostics)
}
