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

package indexmappings

import (
	"maps"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
)

// intersectMappings retains only top-level keys present in state. Within properties,
// only field names from the state's properties tree are kept at every nesting level.
//
// For other top-level keys, when the API value is semantically equal to the
// propertiesKey is the Elasticsearch mapping key whose value is a nested
// field/property map. Centralised so the intersect logic and recursion stay
// in sync.
const propertiesKey = "properties"

// declared state (FieldSemanticallyEqual), the declared value is kept so
// read-after-write matches the configuration shape. Otherwise the API value is
// stored. Plan-time drift is still handled by index.MappingsType semantic equality.
func intersectMappings(apiMappings, stateMappings map[string]any) map[string]any {
	result := make(map[string]any, len(stateMappings))
	for key, stateVal := range stateMappings {
		apiVal, ok := apiMappings[key]
		if !ok {
			// Elasticsearch may omit top-level keys that match defaults; keep the declared value.
			result[key] = stateVal
			continue
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
		if index.FieldSemanticallyEqual(stateVal, apiVal) {
			result[key] = stateVal
			continue
		}
		result[key] = apiVal
	}
	return result
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
