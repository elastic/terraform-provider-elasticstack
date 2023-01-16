package utils

import (
	"fmt"
	"time"
)

// StringIsDuration is a SchemaValidateFunc which tests to make sure the supplied string is valid duration.
func StringIsDuration(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %s to be string", k)}
	}

	if _, err := time.ParseDuration(v); err != nil {
		return nil, []error{fmt.Errorf("%q contains an invalid duration: %s", k, err)}
	}

	return nil, nil
}
