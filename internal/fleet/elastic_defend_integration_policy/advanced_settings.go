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

package elasticdefendintegrationpolicy

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const advancedSettingsKeyPrefix = ".advanced."

var advancedSettingKeyPattern = regexp.MustCompile(`^(linux|mac|windows)\.advanced\.(.+)$`)

// validateAdvancedSettingKeys returns diagnostics for map keys that do not match
// the required OS-prefixed advanced setting pattern.
func validateAdvancedSettingKeys(settings map[string]string) diag.Diagnostics {
	var diags diag.Diagnostics
	for key := range settings {
		if !advancedSettingKeyPattern.MatchString(key) {
			diags.AddAttributeError(
				path.Root("advanced_settings").AtMapKey(key),
				"Invalid advanced setting key",
				fmt.Sprintf(
					"Key %q must match the pattern {linux|mac|windows}.advanced.<setting-path> "+
						"(see Elastic Defend advanced settings documentation).",
					key,
				),
			)
		}
	}
	return diags
}

// advancedSettingsMapFromTerraform extracts a string map from the Terraform
// advanced_settings attribute. Returns nil when the attribute is null or unknown.
func advancedSettingsMapFromTerraform(ctx context.Context, settings types.Map) (map[string]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if settings.IsNull() || settings.IsUnknown() {
		return nil, diags
	}

	var result map[string]string
	d := settings.ElementsAs(ctx, &result, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	diags.Append(validateAdvancedSettingKeys(result)...)
	return result, diags
}

// setNestedValue sets value at a dot-separated path within nested maps.
func setNestedValue(root map[string]any, path string, value string) {
	parts := strings.Split(path, ".")
	current := root
	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
			return
		}
		next, ok := current[part].(map[string]any)
		if !ok || next == nil {
			next = map[string]any{}
			current[part] = next
		}
		current = next
	}
}

// nestedAdvancedToMap flattens a nested advanced object into dot-notation paths
// relative to the advanced subtree (without the OS prefix).
func nestedAdvancedToMap(prefix string, advanced map[string]any, out map[string]string) {
	for key, value := range advanced {
		fullPath := prefix + key
		switch v := value.(type) {
		case map[string]any:
			nestedAdvancedToMap(fullPath+".", v, out)
		case string:
			out[fullPath] = v
		case bool:
			out[fullPath] = fmt.Sprintf("%t", v)
		case float64:
			// JSON numbers decode as float64; preserve integer appearance when possible.
			if v == float64(int64(v)) {
				out[fullPath] = fmt.Sprintf("%d", int64(v))
			} else {
				out[fullPath] = fmt.Sprintf("%v", v)
			}
		default:
			out[fullPath] = fmt.Sprintf("%v", v)
		}
	}
}

// advancedSettingsFromPolicyData flattens policy.{os}.advanced objects from API
// policy data into Terraform advanced_settings keys.
func advancedSettingsFromPolicyData(policyData map[string]any) map[string]string {
	if policyData == nil {
		return nil
	}

	result := map[string]string{}
	for _, os := range []string{"linux", "mac", "windows"} {
		osData, ok := policyData[os].(map[string]any)
		if !ok || osData == nil {
			continue
		}
		advanced, ok := osData["advanced"].(map[string]any)
		if !ok || advanced == nil {
			continue
		}
		relative := map[string]string{}
		nestedAdvancedToMap("", advanced, relative)
		for path, value := range relative {
			result[os+advancedSettingsKeyPrefix+path] = value
		}
	}

	if len(result) == 0 {
		return map[string]string{}
	}
	return result
}

// osesFromAdvancedSettings returns the set of OS names referenced by settings keys.
func osesFromAdvancedSettings(settings map[string]string) map[string]struct{} {
	oses := map[string]struct{}{}
	for key := range settings {
		matches := advancedSettingKeyPattern.FindStringSubmatch(key)
		if len(matches) >= 2 {
			oses[matches[1]] = struct{}{}
		}
	}
	return oses
}

// mergeAdvancedSettingsIntoPolicy merges configured advanced settings into the
// Defend policy payload under policy.{os}.advanced. When settings is non-nil
// but empty, clears advanced for OSes present in priorSettings.
func mergeAdvancedSettingsIntoPolicy(policy map[string]any, settings, priorSettings map[string]string) {
	if settings == nil {
		return
	}

	oses := osesFromAdvancedSettings(settings)
	for os := range osesFromAdvancedSettings(priorSettings) {
		oses[os] = struct{}{}
	}

	for os := range oses {
		osBlock, ok := policy[os].(map[string]any)
		if !ok || osBlock == nil {
			osBlock = map[string]any{}
			policy[os] = osBlock
		}

		advanced := map[string]any{}
		for key, value := range settings {
			matches := advancedSettingKeyPattern.FindStringSubmatch(key)
			if len(matches) != 3 || matches[1] != os {
				continue
			}
			setNestedValue(advanced, matches[2], value)
		}
		osBlock["advanced"] = advanced
	}
}

// advancedSettingsMapToTerraform converts a flat settings map to a Terraform Map value.
func advancedSettingsMapToTerraform(ctx context.Context, settings map[string]string) (types.Map, diag.Diagnostics) {
	var diags diag.Diagnostics
	if settings == nil {
		return types.MapNull(types.StringType), diags
	}

	elements := make(map[string]attr.Value, len(settings))
	for key, value := range settings {
		elements[key] = types.StringValue(value)
	}

	result, d := types.MapValue(types.StringType, elements)
	diags.Append(d...)
	return result, diags
}
