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

package index

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// mappingDiffResult captures the outcome of comparing state and config mappings
// for the purpose of detecting user-owned changes that require replacement.
type mappingDiffResult struct {
	// RequiresReplace is true if a user-owned field's type changed or if
	// properties were removed entirely from config while present in state.
	RequiresReplace bool
	// RemovedFields tracks paths of fields present in state but absent from config.
	RemovedFields []string
	// Diags contains warning diagnostics for removed fields.
	Diags diag.Diagnostics
}

// compareMappingsForPlan compares state mappings against config mappings
// and returns whether replacement is required, which fields were removed,
// and any diagnostics. This is used by the plan modifier.
func compareMappingsForPlan(stateMappings, cfgMappings map[string]any) mappingDiffResult {
	var result mappingDiffResult

	if stateProps, ok := stateMappings["properties"]; ok {
		cfgProps, ok := cfgMappings["properties"]
		if !ok {
			result.RequiresReplace = true
			return result
		}
		result = walkPropertiesForPlan(path.Root("mappings").AtMapKey("properties"), stateProps.(map[string]any), cfgProps.(map[string]any))
	}

	return result
}

// walkPropertiesForPlan recursively walks state properties against config properties
// to detect type changes and removed fields.
func walkPropertiesForPlan(initialPath path.Path, stateProps, cfgProps map[string]any) mappingDiffResult {
	var result mappingDiffResult

	for fieldName, stateFieldRaw := range stateProps {
		stateField, ok := stateFieldRaw.(map[string]any)
		if !ok {
			continue
		}

		currentPath := initialPath.AtMapKey(fieldName)
		cfgFieldRaw, cfgHas := cfgProps[fieldName]

		if !cfgHas {
			result.RemovedFields = append(result.RemovedFields, currentPath.String())
			result.Diags.AddAttributeWarning(
				path.Root("mappings"),
				fmt.Sprintf("removing field [%s] in mappings is ignored.", currentPath),
				"Elasticsearch will maintain the current field in it's mapping. Re-index to remove the field completely",
			)
			continue
		}

		cfgField, ok := cfgFieldRaw.(map[string]any)
		if !ok {
			continue
		}

		// Check type
		stateType, stateHasType := stateField["type"]
		cfgType, cfgHasType := cfgField["type"]

		if stateHasType && cfgHasType {
			if !reflect.DeepEqual(stateType, cfgType) {
				result.RequiresReplace = true
				return result
			}

			// semantic_text special handling: not a replacement if model_settings differs
			if stateType == "semantic_text" {
				if _, stateHasModelSettings := stateField["model_settings"]; stateHasModelSettings {
					if _, cfgHasModelSettings := cfgField["model_settings"]; !cfgHasModelSettings {
						// ES auto-populates model_settings - not a replacement trigger
						continue
					}
				}
			}
			continue
		} else if stateHasType || cfgHasType {
			result.RequiresReplace = true
			return result
		}

		// Check nested properties
		if stateNested, stateHasNested := stateField["properties"]; stateHasNested {
			cfgNested, cfgHasNested := cfgField["properties"]
			if !cfgHasNested {
				result.RemovedFields = append(result.RemovedFields, currentPath.AtMapKey("properties").String())
				result.Diags.AddAttributeWarning(
					path.Root("mappings"),
					fmt.Sprintf("removing field [%s] in mappings is ignored.", currentPath.AtMapKey("properties")),
					"Elasticsearch will maintain the current field in it's mapping. Re-index to remove the field completely",
				)
				continue
			}
			nestedResult := walkPropertiesForPlan(currentPath.AtMapKey("properties"), stateNested.(map[string]any), cfgNested.(map[string]any))
			result.Diags.Append(nestedResult.Diags...)
			if nestedResult.RequiresReplace {
				result.RequiresReplace = true
				return result
			}
			result.RemovedFields = append(result.RemovedFields, nestedResult.RemovedFields...)
		}
	}

	return result
}

// mappingsSemanticallyEqual compares user-owned mappings against API mappings.
// It returns true when the API value is a non-drifting superset of user intent,
// meaning:
//   - All user-owned properties exist in the API with matching types
//   - Template-injected extras (extra properties, dynamic_templates, _meta, etc.) are allowed
//   - semantic_text model_settings auto-populated by ES are allowed
func mappingsSemanticallyEqual(userMappings, apiMappings map[string]any) bool {
	if len(userMappings) == 0 && len(apiMappings) == 0 {
		return true
	}

	for key, userVal := range userMappings {
		apiVal, ok := apiMappings[key]
		if !ok {
			return false
		}

		if key == "properties" {
			userProps, ok := userVal.(map[string]any)
			if !ok {
				return false
			}
			apiProps, ok := apiVal.(map[string]any)
			if !ok {
				return false
			}
			return propertiesSemanticallyEqual(userProps, apiProps)
		}

		if !reflect.DeepEqual(userVal, apiVal) {
			return false
		}
	}

	return true
}

// propertiesSemanticallyEqual recursively checks that all user-owned properties
// exist in the API with semantically equal definitions.
func propertiesSemanticallyEqual(userProps, apiProps map[string]any) bool {
	for fieldName, userFieldRaw := range userProps {
		apiFieldRaw, ok := apiProps[fieldName]
		if !ok {
			return false
		}
		if !fieldSemanticallyEqual(userFieldRaw, apiFieldRaw) {
			return false
		}
	}
	return true
}

// fieldSemanticallyEqual checks if two field definitions are semantically equal,
// allowing for ES-auto-populated values such as semantic_text model_settings.
func fieldSemanticallyEqual(userFieldRaw, apiFieldRaw any) bool {
	userField, ok := userFieldRaw.(map[string]any)
	if !ok {
		return false
	}
	apiField, ok := apiFieldRaw.(map[string]any)
	if !ok {
		return false
	}

	// Determine the field type from user intent
	userType, userHasType := userField["type"]
	apiType, apiHasType := apiField["type"]

	if userHasType && apiHasType {
		if !reflect.DeepEqual(userType, apiType) {
			return false
		}
	} else if userHasType || apiHasType {
		return false
	}

	_, userHasModelSettings := userField["model_settings"]

	for key, userVal := range userField {
		// For semantic_text fields, allow API to have model_settings that the user didn't specify
		if key == "model_settings" && userHasType && userType == "semantic_text" && !userHasModelSettings {
			continue
		}

		apiVal, ok := apiField[key]
		if !ok {
			return false
		}

		if key == "properties" {
			userProps, ok := userVal.(map[string]any)
			if !ok {
				return false
			}
			apiProps, ok := apiVal.(map[string]any)
			if !ok {
				return false
			}
			if !propertiesSemanticallyEqual(userProps, apiProps) {
				return false
			}
			continue
		}

		if !reflect.DeepEqual(userVal, apiVal) {
			return false
		}
	}

	return true
}
