package user

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ validator.String = atMostOneOfValidator{}
	_ validator.String = requiresAttributeValidator{}
	_ validator.String = preferWriteOnlyAttributeValidator{}
)

// atMostOneOfValidator validates that at most one of the specified attributes is set.
type atMostOneOfValidator struct {
	pathExpressions path.Expressions
}

// AtMostOneOf returns a validator which ensures that at most one of the specified attributes is configured.
func AtMostOneOf(pathExpressions ...path.Expression) validator.String {
	return atMostOneOfValidator{
		pathExpressions: pathExpressions,
	}
}

func (v atMostOneOfValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Ensure that at most one of these attributes is configured: %v", v.pathExpressions)
}

func (v atMostOneOfValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v atMostOneOfValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If the current attribute is null or unknown, no validation required
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// Count how many of the specified attributes are set
	expressions := req.PathExpression.MergeExpressions(v.pathExpressions...)

	for _, expression := range v.pathExpressions {
		matchedPaths, diags := req.Config.PathMatches(ctx, expression)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, matchedPath := range matchedPaths {
			// Don't compare with self
			if matchedPath.Equal(req.Path) {
				continue
			}

			var matchedValue attr.Value
			diags := req.Config.GetAttribute(ctx, matchedPath, &matchedValue)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			// If another attribute is also set, that's a conflict
			if !matchedValue.IsNull() && !matchedValue.IsUnknown() {
				resp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
					req.Path,
					fmt.Sprintf("Only one of %s can be configured at a time", expressions),
				))
				return
			}
		}
	}
}

// requiresAttributeValidator validates that if the current attribute is set, the required attribute must also be set.
type requiresAttributeValidator struct {
	requiredPath path.Expression
}

// RequiresAttribute returns a validator which ensures that if the current attribute is configured, the required attribute must also be configured.
func RequiresAttribute(requiredPath path.Expression) validator.String {
	return requiresAttributeValidator{
		requiredPath: requiredPath,
	}
}

func (v requiresAttributeValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Ensure that if configured, %s must also be configured", v.requiredPath)
}

func (v requiresAttributeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v requiresAttributeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If the current attribute is null or unknown, no validation required
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// Check if the required attribute is set
	matchedPaths, diags := req.Config.PathMatches(ctx, v.requiredPath)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(matchedPaths) == 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Missing Required Attribute",
			fmt.Sprintf("Attribute %s requires %s to also be set", req.Path, v.requiredPath),
		)
		return
	}

	for _, matchedPath := range matchedPaths {
		var requiredValue attr.Value
		diags := req.Config.GetAttribute(ctx, matchedPath, &requiredValue)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if requiredValue.IsNull() || requiredValue.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Missing Required Attribute",
				fmt.Sprintf("Attribute %s requires %s to also be set", req.Path, v.requiredPath),
			)
			return
		}
	}
}

// preferWriteOnlyAttributeValidator is a validator that suggests using the write-only version of an attribute
type preferWriteOnlyAttributeValidator struct {
	writeOnlyAttrName string
}

// PreferWriteOnlyAttribute returns a validator that warns when a non-write-only attribute is used instead of its write-only counterpart
func PreferWriteOnlyAttribute(writeOnlyAttrName string) validator.String {
	return preferWriteOnlyAttributeValidator{
		writeOnlyAttrName: writeOnlyAttrName,
	}
}

func (v preferWriteOnlyAttributeValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Suggest using %s for better security with ephemeral resources", v.writeOnlyAttrName)
}

func (v preferWriteOnlyAttributeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v preferWriteOnlyAttributeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// This is just a suggestion/warning, not an error
	// If the attribute is set, we add an informational diagnostic
	if !req.ConfigValue.IsNull() && !req.ConfigValue.IsUnknown() {
		resp.Diagnostics.AddAttributeWarning(
			req.Path,
			"Consider Using Write-Only Attribute",
			fmt.Sprintf("Consider using the '%s' attribute instead when working with ephemeral resources like Vault secrets. "+
				"This prevents the sensitive value from being stored in the state file. "+
				"See https://developer.hashicorp.com/terraform/language/manage-sensitive-data/ephemeral for more information.",
				v.writeOnlyAttrName),
		)
	}
}
