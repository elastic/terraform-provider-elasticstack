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
	"maps"
)

// propertiesKey is the Elasticsearch mapping key whose value is a nested
// field/property map. Centralised so the intersect logic and recursion stay
// in sync.
const propertiesKey = "properties"

// dynamicTemplatesKey is the Elasticsearch mapping key whose value is an
// ordered array of named dynamic templates.
const dynamicTemplatesKey = "dynamic_templates"

// IntersectMappings retains only top-level keys present in state. Within properties,
// only field names from the state's properties tree are kept at every nesting level.
// Within dynamic_templates, only template names declared in state are kept,
// in the API's relative order of those names.
//
// For other top-level keys, when the API value is semantically equal to the
// declared state (FieldSemanticallyEqual), the declared value is kept so
// read-after-write matches the configuration shape. Otherwise the API value is
// stored. Plan-time drift is still handled by MappingsType semantic equality.
func IntersectMappings(apiMappings, stateMappings map[string]any) map[string]any {
	result := make(map[string]any, len(stateMappings))
	for key, stateVal := range stateMappings {
		apiVal, ok := apiMappings[key]
		if !ok {
			if key == dynamicTemplatesKey {
				// API omitted dynamic_templates entirely (e.g. every declared
				// template was removed out-of-band). Persist an empty array so
				// Framework semantic equality cannot re-pin the prior declared
				// value — an omitted key is indistinguishable from "never declared".
				result[key] = []any{}
				continue
			}
			// Elasticsearch may omit top-level keys that match defaults; keep the declared value.
			result[key] = stateVal
			continue
		}
		if key == dynamicTemplatesKey {
			if templates, ok := intersectDynamicTemplates(apiVal, stateVal); ok {
				result[key] = templates
				continue
			}
			// Unparseable shape — fall through to FieldSemanticallyEqual/passthrough.
		}
		if key == propertiesKey {
			apiProps, apiOK := apiVal.(map[string]any)
			stateProps, stateOK := stateVal.(map[string]any)
			if apiOK && stateOK {
				if intersected := intersectProperties(apiProps, stateProps); len(intersected) > 0 {
					result[key] = intersected
				}
				continue
			}
		}
		if FieldSemanticallyEqual(stateVal, apiVal) {
			result[key] = stateVal
			continue
		}
		result[key] = apiVal
	}
	return result
}

// intersectDynamicTemplates filters the API's dynamic_templates array down to
// the names declared in state, using the API's definition for each retained
// name and preserving the API's relative order of those names. Names absent
// from the API are omitted. ok is false when either side cannot be parsed via
// dynamicTemplatesByName, signalling the caller to pass through the API value.
func intersectDynamicTemplates(apiVal, stateVal any) (templates []any, ok bool) {
	apiTemplates, apiOK := dynamicTemplatesByName(apiVal)
	if !apiOK {
		return nil, false
	}
	stateTemplates, stateOK := dynamicTemplatesByName(stateVal)
	if !stateOK {
		return nil, false
	}

	result := make([]any, 0, len(apiTemplates))
	for _, name := range dynamicTemplateNamesInOrder(apiVal) {
		if _, declared := stateTemplates[name]; !declared {
			continue
		}
		result = append(result, map[string]any{name: apiTemplates[name]})
	}
	return result, true
}

func intersectProperties(apiProps, stateProps map[string]any) map[string]any {
	if len(stateProps) == 0 {
		return nil
	}

	result := make(map[string]any, len(stateProps))
	for fieldName, stateField := range stateProps {
		apiField, ok := apiProps[fieldName]
		if !ok {
			continue
		}

		apiMap, apiIsMap := apiField.(map[string]any)
		stateMap, stateIsMap := stateField.(map[string]any)
		if !apiIsMap || !stateIsMap {
			result[fieldName] = apiField
			continue
		}

		apiNested, apiHasProps := apiMap[propertiesKey]
		stateNested, stateHasProps := stateMap[propertiesKey]
		if apiHasProps && stateHasProps {
			apiNestedMap, apiNestedOK := apiNested.(map[string]any)
			stateNestedMap, stateNestedOK := stateNested.(map[string]any)
			if apiNestedOK && stateNestedOK {
				out := make(map[string]any, len(apiMap))
				maps.Copy(out, apiMap)
				if intersected := intersectProperties(apiNestedMap, stateNestedMap); len(intersected) > 0 {
					out[propertiesKey] = intersected
				} else {
					delete(out, propertiesKey)
				}
				result[fieldName] = out
				continue
			}
		}

		result[fieldName] = apiField
	}
	return result
}
