package datafeed

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
		for k, v := range fieldMap {
			resultField[k] = v
		}

		// Set ignore_failure default to false if not specified
		if _, exists := resultField["ignore_failure"]; !exists {
			resultField["ignore_failure"] = false
		}

		// Set script lang default to "painless" if not specified and script exists
		if scriptInterface, exists := resultField["script"]; exists {
			if scriptMap, ok := scriptInterface.(map[string]any); ok {
				// Create a copy of the script map
				newScriptMap := make(map[string]any)
				for k, v := range scriptMap {
					newScriptMap[k] = v
				}

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
