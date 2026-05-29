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

package cloudconnector

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	varsElementValidatorDescription          = "each vars element must configure exactly one API union arm and must not set computed-only attributes in configuration"
	structuredFieldsRequireTypeDetailMessage = "`type` is required when any of `value`, `secret_value`, or `frozen` is configured. " +
		"Set `type` (e.g. `\"text\"` or `\"password\"`) or omit the structured fields."
)

// varsElementValidator enforces per-element exclusivity rules for the vars map union.
type varsElementValidator struct{}

func (varsElementValidator) Description(_ context.Context) string {
	return varsElementValidatorDescription
}

func (varsElementValidator) MarkdownDescription(_ context.Context) string {
	return varsElementValidatorDescription
}

func (varsElementValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	obj := req.ConfigValue
	if obj.IsNull() || obj.IsUnknown() {
		return
	}

	attrs := obj.Attributes()

	if configValueIsSet(attrs[attrVarsSecretRef]) {
		resp.Diagnostics.AddAttributeError(
			req.Path.AtName(attrVarsSecretRef),
			"Invalid vars element",
			"`secret_ref` is computed-only and must not be set in configuration.",
		)
		return
	}

	groupA := []struct {
		name string
		set  bool
	}{
		{attrVarsString, configValueIsSet(attrs[attrVarsString])},
		{attrVarsNumber, configValueIsSet(attrs[attrVarsNumber])},
		{attrVarsBool, configValueIsSet(attrs[attrVarsBool])},
	}

	groupB := []struct {
		name string
		set  bool
	}{
		{attrVarsType, configValueIsSet(attrs[attrVarsType])},
		{attrVarsFrozen, configValueIsSet(attrs[attrVarsFrozen])},
		{attrVarsValue, configValueIsSet(attrs[attrVarsValue])},
		{attrVarsSecretValue, configValueIsSet(attrs[attrVarsSecretValue])},
	}

	groupACount := 0
	var groupASet []string
	for _, field := range groupA {
		if field.set {
			groupACount++
			groupASet = append(groupASet, field.name)
		}
	}

	groupBNonFrozenCount := 0
	var groupBSet []string
	for _, field := range groupB {
		if field.set {
			groupBSet = append(groupBSet, field.name)
			if field.name != attrVarsFrozen {
				groupBNonFrozenCount++
			}
		}
	}

	if groupACount > 1 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid vars element",
			fmt.Sprintf("At most one bare var arm may be set; conflicting attributes: %s.", strings.Join(groupASet, ", ")),
		)
		return
	}

	if groupACount > 0 && len(groupBSet) > 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid vars element",
			fmt.Sprintf("Bare var arms (%s) cannot be combined with structured var attributes (%s).", strings.Join(groupASet, ", "), strings.Join(groupBSet, ", ")),
		)
		return
	}

	if !configValueIsSet(attrs[attrVarsType]) {
		if configValueIsSet(attrs[attrVarsValue]) || configValueIsSet(attrs[attrVarsSecretValue]) || configValueIsSet(attrs[attrVarsFrozen]) {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid vars element",
				structuredFieldsRequireTypeDetailMessage,
			)
			return
		}
	}

	if configValueIsSet(attrs[attrVarsType]) {
		valueArms := 0
		var valueArmNames []string
		if configValueIsSet(attrs[attrVarsValue]) {
			valueArms++
			valueArmNames = append(valueArmNames, attrVarsValue)
		}
		if configValueIsSet(attrs[attrVarsSecretValue]) {
			valueArms++
			valueArmNames = append(valueArmNames, attrVarsSecretValue)
		}

		if valueArms == 0 {
			if attrs[attrVarsValue].IsUnknown() || attrs[attrVarsSecretValue].IsUnknown() {
				return
			}
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid vars element",
				"When `type` is set, exactly one of `value` or `secret_value` must be set.",
			)
			return
		}
		if valueArms > 1 {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid vars element",
				fmt.Sprintf("When `type` is set, exactly one of `value` or `secret_value` may be set; conflicting attributes: %s.", strings.Join(valueArmNames, ", ")),
			)
			return
		}
		return
	}

	if groupACount == 0 && len(groupBSet) == 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid vars element",
			"At least one vars union arm must be configured (`string`, `number`, `bool`, or structured `type` with `value`/`secret_value`).",
		)
	}
}

type providerBlockMatchesCloudProvider struct{}

func (providerBlockMatchesCloudProvider) Description(_ context.Context) string {
	return "typed block must match cloud_provider when both are known"
}

func (providerBlockMatchesCloudProvider) MarkdownDescription(ctx context.Context) string {
	return providerBlockMatchesCloudProvider{}.Description(ctx)
}

func (providerBlockMatchesCloudProvider) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cloudProvider types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(attrCloudProvider), &cloudProvider)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if cloudProvider.IsNull() || cloudProvider.IsUnknown() {
		return
	}

	var aws, azure types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(attrAWSBlock), &aws)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(attrAzureBlock), &azure)...)
	if resp.Diagnostics.HasError() {
		return
	}

	provider := cloudProvider.ValueString()
	if configObjectIsSet(aws) && provider != cloudProviderAWS {
		resp.Diagnostics.AddAttributeError(
			path.Root(attrAWSBlock),
			"Invalid typed block for cloud_provider",
			fmt.Sprintf("The `aws` block requires `cloud_provider = %q`; got %q.", cloudProviderAWS, provider),
		)
	}
	if configObjectIsSet(azure) && provider != cloudProviderAzure {
		resp.Diagnostics.AddAttributeError(
			path.Root(attrAzureBlock),
			"Invalid typed block for cloud_provider",
			fmt.Sprintf("The `azure` block requires `cloud_provider = %q`; got %q.", cloudProviderAzure, provider),
		)
	}
}

func configValueIsSet(v attr.Value) bool {
	return !v.IsNull() && !v.IsUnknown()
}

func configObjectIsSet(obj types.Object) bool {
	return !obj.IsNull() && !obj.IsUnknown()
}
