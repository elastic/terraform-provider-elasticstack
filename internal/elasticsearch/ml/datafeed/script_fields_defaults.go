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

package datafeed

import "maps"

// populateScriptFieldsDefaults ensures that all script fields have proper defaults
func populateScriptFieldsDefaults(model map[string]any) map[string]any {
	for fieldName, field := range model {
		// Copy the field
		fieldMap, ok := field.(map[string]any)
		if !ok {
			continue
		}

		resultField := make(map[string]any)
		// Copy all existing fields
		maps.Copy(resultField, fieldMap)

		// Set ignore_failure default to false if not specified
		if _, exists := resultField["ignore_failure"]; !exists {
			resultField["ignore_failure"] = false
		}

		// Set script lang default to "painless" if not specified and script exists
		if scriptInterface, exists := resultField["script"]; exists {
			if scriptMap, ok := scriptInterface.(map[string]any); ok {
				// Create a copy of the script map
				newScriptMap := make(map[string]any)
				maps.Copy(newScriptMap, scriptMap)

				// Set lang default to "painless" if not specified
				if _, langExists := newScriptMap["lang"]; !langExists {
					newScriptMap["lang"] = "painless"
				}

				resultField["script"] = newScriptMap
			}
		}

		model[fieldName] = resultField
	}

	return model
}
