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

package validators

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// elasticDurationRegexp matches the Elastic time-unit duration format
// documented at
// https://www.elastic.co/guide/en/elasticsearch/reference/current/api-conventions.html#time-units
// (d, h, m, s, ms, micros, nanos).
var elasticDurationRegexp = regexp.MustCompile(`^[0-9]+(?:\.[0-9]+)?(?:d|h|m|s|ms|micros|nanos)$`)

// ElasticDuration returns a Plugin Framework string validator that ensures the
// value is a valid Elastic time-unit duration (d, h, m, s, ms, micros, nanos).
// Null and unknown values are skipped.
func ElasticDuration() validator.String {
	return elasticDurationValidator{}
}

type elasticDurationValidator struct{}

var _ validator.String = elasticDurationValidator{}

func (v elasticDurationValidator) Description(_ context.Context) string {
	return "must be a valid Elastic duration (e.g. 1d, 2h, 30m, 60s, 500ms, 1micros, 1nanos)"
}

func (v elasticDurationValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v elasticDurationValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	val := req.ConfigValue
	if val.IsNull() || val.IsUnknown() {
		return
	}
	s := val.ValueString()
	if s == "" {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid Elastic duration", "duration must not be empty")
		return
	}
	if !elasticDurationRegexp.MatchString(s) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Elastic duration",
			fmt.Sprintf("%q is not a valid Elastic time-unit duration", s),
		)
	}
}

// unitDurationRegexp builds a regexp matching one or more digits followed by
// a single unit drawn from units, e.g. unitDurationRegexp("mhd") matches
// "5m", "3h", "6d". units is inserted verbatim into a character class, so
// callers must only pass literal unit letters (no regexp metacharacters).
func unitDurationRegexp(units string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(`^\d+[%s]$`, units))
}

// DurationWithUnits returns a Plugin Framework string validator that accepts
// a positive integer followed by a single unit character drawn from units
// (e.g. units="mhd" accepts "5m", "3h", "6d"). message is used as the
// validation failure detail. Null and unknown values are skipped.
func DurationWithUnits(units string, message string) validator.String {
	return unitDurationValidator{regexp: unitDurationRegexp(units), message: message}
}

type unitDurationValidator struct {
	regexp  *regexp.Regexp
	message string
}

var _ validator.String = unitDurationValidator{}

func (v unitDurationValidator) Description(_ context.Context) string {
	return v.message
}

func (v unitDurationValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v unitDurationValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	val := req.ConfigValue
	if val.IsNull() || val.IsUnknown() {
		return
	}
	if !v.regexp.MatchString(val.ValueString()) {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid Attribute Value", v.message)
	}
}

// ParseUnitDuration parses a "digits + unit letter" duration string such as
// "5m" or "30s" into its numeric value and unit, validating that unit is one
// of the characters in units. It is the parsing counterpart to
// DurationWithUnits, for callers that need the decomposed value rather than
// just a validation pass/fail.
func ParseUnitDuration(s string, units string) (value int, unit string, err error) {
	if !unitDurationRegexp(units).MatchString(s) {
		return 0, "", fmt.Errorf("%q does not match the required format: digits followed by one of [%s]", s, units)
	}
	numeric := s[:len(s)-1]
	unit = s[len(s)-1:]
	value, err = strconv.Atoi(numeric)
	if err != nil {
		return 0, "", fmt.Errorf("failed to parse duration value %q: %w", numeric, err)
	}
	return value, unit, nil
}
