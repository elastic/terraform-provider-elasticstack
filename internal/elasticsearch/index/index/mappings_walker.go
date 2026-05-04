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
	"maps"
	"reflect"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

const semanticTextType = index.SemanticTextMappingType

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

	if statePropsRaw, ok := stateMappings["properties"]; ok {
		cfgPropsRaw, ok := cfgMappings["properties"]
		if !ok {
			result.RequiresReplace = true
			return result
		}

		stateProps, stateOK := statePropsRaw.(map[string]any)
		cfgProps, cfgOK := cfgPropsRaw.(map[string]any)
		if !stateOK || !cfgOK {
			// Invalid properties structure; schema validation normally prevents this.
			return result
		}

		result = walkPropertiesForPlan(path.Root("mappings").AtMapKey("properties"), stateProps, cfgProps)
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
				"Elasticsearch will maintain the current field in its mapping. Re-index to remove the field completely",
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

			continue
		} else if stateHasType || cfgHasType {
			result.RequiresReplace = true
			return result
		}

		// Check nested properties
		if stateNestedRaw, stateHasNested := stateField["properties"]; stateHasNested {
			cfgNestedRaw, cfgHasNested := cfgField["properties"]
			if !cfgHasNested {
				result.RemovedFields = append(result.RemovedFields, currentPath.AtMapKey("properties").String())
				result.Diags.AddAttributeWarning(
					path.Root("mappings"),
					fmt.Sprintf("removing field [%s] in mappings is ignored.", currentPath.AtMapKey("properties")),
					"Elasticsearch will maintain the current field in its mapping. Re-index to remove the field completely",
				)
				continue
			}
			stateNested, stateOK := stateNestedRaw.(map[string]any)
			cfgNested, cfgOK := cfgNestedRaw.(map[string]any)
			if !stateOK || !cfgOK {
				continue
			}
			nestedResult := walkPropertiesForPlan(currentPath.AtMapKey("properties"), stateNested, cfgNested)
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

// mergeMappingsForPlan merges state mappings into config mappings to produce a
// planned value that preserves state-only fields (those retained by ES or
// injected by templates).
func mergeMappingsForPlan(stateMappings, cfgMappings map[string]any) map[string]any {
	merged := make(map[string]any, len(cfgMappings))
	maps.Copy(merged, cfgMappings)

	// Merge properties recursively
	if stateProps, ok := stateMappings["properties"].(map[string]any); ok {
		cfgProps, ok := merged["properties"].(map[string]any)
		if ok {
			merged["properties"] = mergeProperties(stateProps, cfgProps)
		}
	}

	// Copy known top-level template-injected keys from state that are not in config.
	for _, key := range []string{"dynamic_templates", "_meta", "runtime"} {
		if val, exists := stateMappings[key]; exists {
			if _, inConfig := merged[key]; !inConfig {
				merged[key] = val
			}
		}
	}

	return merged
}

func mergeProperties(stateProps, cfgProps map[string]any) map[string]any {
	merged := make(map[string]any, len(cfgProps))
	maps.Copy(merged, cfgProps)

	for fieldName, stateFieldRaw := range stateProps {
		stateField, ok := stateFieldRaw.(map[string]any)
		if !ok {
			continue
		}

		cfgFieldRaw, cfgHas := merged[fieldName]
		if !cfgHas {
			// Field retained by Elasticsearch
			merged[fieldName] = stateFieldRaw
			continue
		}

		cfgField, ok := cfgFieldRaw.(map[string]any)
		if !ok {
			continue
		}

		mergedField := make(map[string]any, len(cfgField))
		maps.Copy(mergedField, cfgField)

		// Check type
		stateType, stateHasType := stateField["type"]
		cfgType, cfgHasType := cfgField["type"]

		if stateHasType && cfgHasType && reflect.DeepEqual(stateType, cfgType) {
			// For semantic_text fields, copy model_settings from state if not in config
			if stateType == semanticTextType {
				if modelSettings, exists := stateField["model_settings"]; exists {
					if _, configHasModelSettings := cfgField["model_settings"]; !configHasModelSettings {
						mergedField["model_settings"] = modelSettings
					}
				}
			}
		}

		// Recursively merge nested properties
		if stateNestedRaw, stateHasNested := stateField["properties"]; stateHasNested {
			cfgNestedRaw, cfgHasNested := cfgField["properties"]
			if cfgHasNested {
				stateNested, stateOK := stateNestedRaw.(map[string]any)
				cfgNested, cfgOK := cfgNestedRaw.(map[string]any)
				if stateOK && cfgOK {
					mergedField["properties"] = mergeProperties(stateNested, cfgNested)
				} else {
					mergedField["properties"] = cfgNestedRaw
				}
			} else {
				mergedField["properties"] = stateNestedRaw
			}
		}

		merged[fieldName] = mergedField
	}

	return merged
}

// mappingsSemanticallyEqual is retained here as a package-level helper so that
// existing tests in this package can call it directly.  The implementation now
// delegates to the shared customtypes helper.
func mappingsSemanticallyEqual(userMappings, apiMappings map[string]any) bool {
	return index.MappingsSemanticallyEqual(userMappings, apiMappings)
}
