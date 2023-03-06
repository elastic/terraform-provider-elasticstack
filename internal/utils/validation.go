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

// StringIsElasticDuration is a SchemaValidateFunc which tests to make sure the supplied string is valid duration using Elastic time units:
// d, h, m, s, ms, micros, nanos. (see https://www.elastic.co/guide/en/elasticsearch/reference/current/api-conventions.html#time-units)
func StringIsElasticDuration(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %s to be string", k)}
	}

	if v == "" {
		return nil, []error{fmt.Errorf("%q contains an invalid duration: [empty]", k)}
	}

	firstPartCount := 0
	for v != "" {
		// first part must contain only characters in range [0-9] and .
		if ('0' <= v[0] && v[0] <= '9') || v[0] == '.' {
			v = v[1:]
			firstPartCount++
			continue
		}

		if firstPartCount == 0 {
			return nil, []error{fmt.Errorf("%q contains an invalid duration: should start with a numeric value", k)}
		}

		if !isValidElasticTimeUnit(v) {
			return nil, []error{fmt.Errorf("%q contains an invalid duration: unrecognized time unit [%s]", k, v)}
		}

		break
	}

	return nil, nil
}

func isValidElasticTimeUnit(timeUnit string) bool {
	switch timeUnit {
	case
		"d",
		"h",
		"m",
		"s",
		"ms",
		"micros",
		"nanos":
		return true
	}
	return false
}
