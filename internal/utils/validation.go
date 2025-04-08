package utils

import (
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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

// StringIsHours is a SchemaValidateFunc which tests to make sure the supplied string is in the required format of HH:mm
func StringIsHours(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %s to be string", k)}
	}

	if v == "" {
		return nil, []error{fmt.Errorf("%q string is not a valid time in HH:mm format: [empty]", k)}
	}

	r := regexp.MustCompile(`^([0-9]|0[0-9]|1[0-9]|2[0-3]):[0-5][0-9]$`)

	if !r.MatchString(v) {
		return nil, []error{fmt.Errorf("%q string is not a valid time in HH:mm format", k)}
	}

	return nil, nil
}

// Avoid lint error on deprecated SchemaValidateFunc usage.
//
//nolint:staticcheck
func StringIsAlertingDuration() schema.SchemaValidateFunc {
	r := regexp.MustCompile(`^[1-9][0-9]*(?:d|h|m|s)$`)
	return validation.StringMatch(r, "string is not a valid Alerting duration in seconds (s), minutes (m), hours (h), or days (d)")
}

// Avoid lint error on deprecated SchemaValidateFunc usage.
//
//nolint:staticcheck
func StringIsMaintenanceWindowOnWeekDay() schema.SchemaValidateFunc {
	r := regexp.MustCompile(`^(((\+|-)[1-5])?(MO|TU|WE|TH|FR|SA|SU))$`)
	return validation.StringMatch(r, "string is not a valid OnWeekDay. Accepted values are specific days of the week (`[MO,TU,WE,TH,FR,SA,SU]`) or nth day of month (`[+1MO, -3FR, +2WE, -4SA, -5SU]`).")
}

// Avoid lint error on deprecated SchemaValidateFunc usage.
//
//nolint:staticcheck
func StringIsMaintenanceWindowIntervalFrequency() schema.SchemaValidateFunc {
	r := regexp.MustCompile(`^[1-9][0-9]*(?:d|w|M|y)$`)
	return validation.StringMatch(r, "string is not a valid interval/frequency. Allowed values are in the `<integer><unit>` format. `<unit>` is one of `d`, `w`, `M`, or `y` for days, weeks, months, years. For example: `15d`, `2w`, `3m`, `1y`.")
}
