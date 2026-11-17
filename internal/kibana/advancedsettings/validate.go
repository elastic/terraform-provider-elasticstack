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

package advancedsettings

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	minSettingKeyLength = 1
	maxSettingKeyLength = 1024

	resetAdvice = "To reset the setting to its Kibana default, remove the key."
)

func settingsValidators() []validator.Map {
	return []validator.Map{
		mapvalidator.KeysAre(stringvalidator.UTF8LengthBetween(minSettingKeyLength, maxSettingKeyLength)),
		noNullEntriesValidator{},
	}
}

type noNullEntriesValidator struct{}

func (noNullEntriesValidator) Description(context.Context) string {
	return "no entry may be null; remove the key to reset the setting"
}

func (v noNullEntriesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (noNullEntriesValidator) ValidateMap(_ context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	for key, element := range req.ConfigValue.Elements() {
		if element.IsNull() {
			resp.Diagnostics.Append(nullEntryDiagnostic(req.Path, key))
		}
	}
}

func nullEntryDiagnostic(settingsPath path.Path, key string) diag.Diagnostic {
	return diag.NewAttributeErrorDiagnostic(
		settingsPath.AtMapKey(key),
		"Null setting value",
		fmt.Sprintf("The value of setting %q is null. %s", key, resetAdvice),
	)
}

func (r *Resource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var stringSettings, jsonSettings types.Map
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, stringSettingsPath, &stringSettings)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, jsonSettingsPath, &jsonSettings)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(keysDeclaredInBothMaps(stringSettings, jsonSettings)...)
	resp.Diagnostics.Append(invalidJSONSettings(jsonSettings)...)
}

func keysDeclaredInBothMaps(stringSettings, jsonSettings types.Map) diag.Diagnostics {
	var diags diag.Diagnostics
	for key := range stringSettings.Elements() {
		if _, declaredInBoth := jsonSettings.Elements()[key]; declaredInBoth {
			diags.AddAttributeError(
				jsonSettingsPath.AtMapKey(key),
				"Setting declared in both maps",
				fmt.Sprintf("The setting %q is declared in both string_settings and json_settings. Declare it in only one of them.", key),
			)
		}
	}
	return diags
}

func invalidJSONSettings(jsonSettings types.Map) diag.Diagnostics {
	var diags diag.Diagnostics
	for key, element := range jsonSettings.Elements() {
		if !typeutils.IsKnown(element) {
			continue
		}
		value, _ := element.(jsontypes.Normalized)
		_, jsonDiags := parseJSONSetting(key, value.ValueString())
		diags.Append(jsonDiags...)
	}
	return diags
}

func parseJSONSetting(key, text string) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	entryPath := jsonSettingsPath.AtMapKey(key)
	parsed, err := parseJSONKeepingNumbers(text)
	switch {
	case err != nil:
		diags.AddAttributeError(entryPath, "Invalid JSON value", fmt.Sprintf("The value of setting %q is not valid JSON: %s", key, err))
	case parsed == nil:
		diags.AddAttributeError(entryPath, "JSON null is not allowed", fmt.Sprintf("The value of setting %q is the JSON value null. %s", key, resetAdvice))
	}
	return parsed, diags
}
