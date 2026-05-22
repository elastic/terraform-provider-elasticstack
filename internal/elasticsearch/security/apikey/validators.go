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
	_ validator.String = RequiresTypeValidator{}
	_ validator.Object = RequiresTypeValidator{}
)

// RequiresTypeValidator validates that a string attribute is only provided
// when the resource has a specific value for the "type" attribute.
type RequiresTypeValidator struct {
	expectedType string
}

// RequiresType returns a validator which ensures that the configured attribute
// is only provided when the "type" attribute matches the expected value.
func RequiresType(expectedType string) RequiresTypeValidator {
	return RequiresTypeValidator{
		expectedType: expectedType,
	}
}

// EffectiveAPIKeyTypeFromOptionalString returns the effective API key type for
// an optional string config attribute, defaulting to DefaultAPIKeyType when
// the attribute is unset.
func EffectiveAPIKeyTypeFromOptionalString(typeAttr *string) string {
	if typeAttr == nil || *typeAttr == "" {
		return DefaultAPIKeyType
	}
	return *typeAttr
}

func (v RequiresTypeValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Ensures that the attribute is only provided when type=%s", v.expectedType)
}

func (v RequiresTypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v RequiresTypeValidator) validateType(ctx context.Context, config tfsdk.Config, attrPath path.Path, diagnostics *diag.Diagnostics) {
	var typeAttr *string
	diagnostics.Append(config.GetAttribute(ctx, path.Root("type"), &typeAttr)...)
	if diagnostics.HasError() {
		return
	}

	// Treat unset type the same as Open(): default to rest.
	apiKeyType := EffectiveAPIKeyTypeFromOptionalString(typeAttr)
	if apiKeyType == v.expectedType {
		return
	}

	diagnostics.AddAttributeError(
		attrPath,
		fmt.Sprintf("Attribute not valid for API key type '%s'", apiKeyType),
		fmt.Sprintf("The %s attribute can only be used when type='%s', but type='%s' was specified.",
			attrPath.String(), v.expectedType, apiKeyType),
	)
}

func (v RequiresTypeValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	v.validateType(ctx, req.Config, req.Path, &resp.Diagnostics)
}

func (v RequiresTypeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	v.validateType(ctx, req.Config, req.Path, &resp.Diagnostics)
}
