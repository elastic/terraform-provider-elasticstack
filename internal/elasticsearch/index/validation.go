package index

import (
	"encoding/json"
	"fmt"
)

func stringIsJSONObject(i interface{}, s string) (warnings []string, errors []error) {
	iStr, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", s))
		return warnings, errors
	}

	m := map[string]interface{}{}
	if err := json.Unmarshal([]byte(iStr), &m); err != nil {
		errors = append(errors, fmt.Errorf("expected %s to be a JSON object. Check the documentation for the expected format. %w", s, err))
		return
	}

	return
}
