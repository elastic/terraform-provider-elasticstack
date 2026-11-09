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

// Package utils holds schema helpers shared by the Alerting v2 resource
// packages: the entity-id and duration validators and the "already exists,
// import instead" diagnostic builders that every resource needs with the
// same wording.
package utils

import (
	"context"
	"regexp"
	"strconv"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var entityIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,150}$`)

// EntityIDValidator accepts the identifiers the Alerting v2 API allows for
// rules and action policies: 1-150 letters, digits, underscores, or hyphens.
func EntityIDValidator() validator.String {
	return stringvalidator.RegexMatches(
		entityIDPattern,
		"must be 1-150 characters of letters, digits, underscores, or hyphens",
	)
}

var durationPattern = regexp.MustCompile(`^(\d+)(ms|s|m|h|d|w)$`)

const (
	durationGrammarDescription = "must be an unsigned integer followed by one of ms, s, m, h, d, or w (for example 5m)"
	maxDuration                = 365 * 24 * time.Hour
)

var durationUnits = map[string]time.Duration{
	"ms": time.Millisecond,
	"s":  time.Second,
	"m":  time.Minute,
	"h":  time.Hour,
	"d":  24 * time.Hour,
	"w":  7 * 24 * time.Hour,
}

// DurationValidator accepts the Alerting v2 duration grammar: an unsigned
// integer followed by ms, s, m, h, d, or w, totalling at most 365 days.
func DurationValidator() validator.String {
	return durationValidator{}
}

// DurationValidatorAtLeast is DurationValidator with an additional lower
// bound, expressed in the same grammar (for example "5s"). It panics when
// minimum is not a valid duration so a schema typo fails at startup.
func DurationValidatorAtLeast(minimum string) validator.String {
	parsed, ok := parseDuration(minimum)
	if !ok {
		panic("utils.DurationValidatorAtLeast: invalid minimum duration " + minimum)
	}
	return durationValidator{minimum: parsed, minimumText: minimum}
}

type durationValidator struct {
	minimum     time.Duration
	minimumText string
}

func (v durationValidator) Description(_ context.Context) string {
	description := durationGrammarDescription + " and at most 365d"
	if v.minimum > 0 {
		description += " and at least " + v.minimumText
	}
	return description
}

func (v durationValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v durationValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if !typeutils.IsKnown(req.ConfigValue) {
		return
	}
	value := req.ConfigValue.ValueString()
	parsed, ok := parseDuration(value)
	if !ok {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(req.Path, durationGrammarDescription, value))
		return
	}
	if parsed > maxDuration {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(req.Path, "must be at most 365d", value))
		return
	}
	if parsed < v.minimum {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(req.Path, "must be at least "+v.minimumText, value))
	}
}

func parseDuration(value string) (time.Duration, bool) {
	match := durationPattern.FindStringSubmatch(value)
	if match == nil {
		return 0, false
	}
	amount, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return 0, false
	}
	return time.Duration(amount) * durationUnits[match[2]], true
}
