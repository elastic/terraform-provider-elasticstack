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

package kibanacustomtypes

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.StringValuable                   = (*AlertingDuration)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*AlertingDuration)(nil)
	_ xattr.ValidateableAttribute                = (*AlertingDuration)(nil)
)

// AlertingDuration is a custom StringValue for Kibana alerting duration
// strings. The parser supports every unit suffix used by Kibana's alerting
// and maintenance-window APIs: `s` (seconds), `m` (minutes), `h` (hours),
// `d` (days), `w` (weeks), `M` (months, approximated as 30 days), and
// `y` (years, approximated as 365 days). Which subset of those units is
// actually permitted at validation time is determined by the units field,
// which is propagated from the owning AlertingDurationType.
//
// Semantic equality compares the parsed durations rather than the raw
// strings, so values like `1d` and `24h` compare equal regardless of the
// allowed unit set.
type AlertingDuration struct {
	basetypes.StringValue

	// units restricts which suffix letters ValidateAttribute will accept. It
	// is set by AlertingDurationType.ValueFromString so values that enter the
	// model through the framework's normal conversion path carry the
	// schema-declared restriction. A zero-value units permits every unit the
	// parser understands (used by direct constructors like
	// NewAlertingDurationValue, where validation does not apply).
	units AlertingDurationUnits
}

func (v AlertingDuration) Type(_ context.Context) attr.Type {
	return AlertingDurationType{Units: v.units}
}

func (v AlertingDuration) Equal(o attr.Value) bool {
	other, ok := o.(AlertingDuration)
	if !ok {
		return false
	}
	return v.StringValue.Equal(other.StringValue)
}

func (v AlertingDuration) ValidateAttribute(_ context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if v.IsNull() || v.IsUnknown() {
		return
	}

	raw := v.ValueString()
	if _, err := parseAlertingDuration(raw); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Kibana alerting duration value",
			fmt.Sprintf("A string value was provided that is not a valid Kibana alerting duration: %s. "+
				"Expected an unsigned integer followed by one of %s.",
				err.Error(), v.units.Describe()),
		)
		return
	}

	unit := raw[len(raw)-1]
	if !v.units.Allowed(unit) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Unsupported Kibana alerting duration unit",
			fmt.Sprintf("This attribute does not accept the unit %q in %q. Allowed units: %s.",
				string(unit), raw, v.units.Describe()),
		)
	}
}

// StringSemanticEquals reports whether two AlertingDuration values parse to
// the same time.Duration. Null/unknown values are equal only to values of the
// same state.
func (v AlertingDuration) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(AlertingDuration)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)
		return false, diags
	}

	if v.IsNull() {
		return newValue.IsNull(), diags
	}
	if v.IsUnknown() {
		return newValue.IsUnknown(), diags
	}

	vParsed, d := v.Parse()
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	newParsed, d := newValue.Parse()
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	return vParsed == newParsed, diags
}

// Parse returns the parsed time.Duration. Null/unknown values produce an
// error diagnostic.
func (v AlertingDuration) Parse() (time.Duration, diag.Diagnostics) {
	var diags diag.Diagnostics

	if v.IsNull() {
		diags.Append(diag.NewErrorDiagnostic("AlertingDuration Parse error", "alerting duration string value is null"))
		return 0, diags
	}
	if v.IsUnknown() {
		diags.Append(diag.NewErrorDiagnostic("AlertingDuration Parse Error", "alerting duration string value is unknown"))
		return 0, diags
	}

	duration, err := parseAlertingDuration(v.ValueString())
	if err != nil {
		diags.Append(diag.NewErrorDiagnostic("AlertingDuration Parse Error", err.Error()))
	}
	return duration, diags
}

// parseAlertingDuration parses Kibana alerting duration strings of the form
// <unsigned-int><unit> where unit is one of s, m, h, d, w, M, y. Months and
// years use fixed approximations (30 and 365 days) so semantic equality
// remains deterministic and direction-independent.
func parseAlertingDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, fmt.Errorf("empty duration string")
	}

	// Unit is the final byte (M is upper-case to disambiguate from minutes).
	unit := s[len(s)-1]
	numPart := s[:len(s)-1]

	n, err := strconv.ParseUint(numPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: numeric prefix is not an unsigned integer", s)
	}
	if n == 0 {
		return 0, fmt.Errorf("invalid duration %q: must be greater than zero", s)
	}
	if strings.HasPrefix(numPart, "0") && numPart != "0" {
		return 0, fmt.Errorf("invalid duration %q: leading zeros are not allowed", s)
	}

	var unitDuration time.Duration
	switch unit {
	case 's':
		unitDuration = time.Second
	case 'm':
		unitDuration = time.Minute
	case 'h':
		unitDuration = time.Hour
	case 'd':
		unitDuration = 24 * time.Hour
	case 'w':
		unitDuration = 7 * 24 * time.Hour
	case 'M':
		unitDuration = 30 * 24 * time.Hour
	case 'y':
		unitDuration = 365 * 24 * time.Hour
	default:
		return 0, fmt.Errorf("invalid duration %q: unrecognised unit %q (expected one of s, m, h, d, w, M, y)", s, string(unit))
	}

	if n > uint64(math.MaxInt64) {
		return 0, fmt.Errorf("invalid duration %q: value is too large", s)
	}
	maxNForUnit := uint64(math.MaxInt64 / int64(unitDuration))
	if n > maxNForUnit {
		return 0, fmt.Errorf("invalid duration %q: value is too large for unit %q", s, string(unit))
	}

	return time.Duration(n) * unitDuration, nil
}

func NewAlertingDurationNull() AlertingDuration {
	return AlertingDuration{StringValue: basetypes.NewStringNull()}
}

func NewAlertingDurationUnknown() AlertingDuration {
	return AlertingDuration{StringValue: basetypes.NewStringUnknown()}
}

func NewAlertingDurationValue(value string) AlertingDuration {
	return AlertingDuration{StringValue: basetypes.NewStringValue(value)}
}

func NewAlertingDurationPointerValue(value *string) AlertingDuration {
	return AlertingDuration{StringValue: basetypes.NewStringPointerValue(value)}
}
