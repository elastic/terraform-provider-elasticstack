// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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
