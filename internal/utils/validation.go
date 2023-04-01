package utils

import (
	"fmt"
	"regexp"
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

	r := regexp.MustCompile(`^[0-9]+(?:\.[0-9]+)?(?:d|h|m|s|ms|micros|nanos)$`)

	if !r.MatchString(v) {
		return nil, []error{fmt.Errorf("%q contains an invalid duration: not conforming to Elastic time-units format", k)}
	}

	return nil, nil
}
