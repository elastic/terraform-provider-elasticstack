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

package customtypes

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable = (*AlertingDurationType)(nil)
)

// AlertingDurationUnits is the set of time-unit suffixes a particular
// AlertingDurationType attribute will accept (for example "s, m, h, d" for a
// rule check interval, or "d, w, M, y" for a recurring-schedule frequency).
// Use one of the package-level presets (AlertingDurationUnitsSubDay,
// IntervalFrequencyUnits, AllAlertingDurationUnits) when declaring a schema
// attribute. Unit sets are compared structurally for type equality.
type AlertingDurationUnits struct {
	// units is the sorted, deduplicated list of allowed unit suffix bytes.
	units string
}

// Allowed reports whether the given unit byte (e.g. 'd') is permitted by this
// unit set. A zero-value AlertingDurationUnits permits every unit the parser
// understands; this keeps the type usable in contexts where it is constructed
// without a schema (for example values produced via the New* helpers when
// reading API responses).
func (u AlertingDurationUnits) Allowed(unit byte) bool {
	if u.units == "" {
		return strings.IndexByte(allParserUnits, unit) >= 0
	}
	return strings.IndexByte(u.units, unit) >= 0
}

// Describe returns a human readable rendering of the allowed units, suitable
// for embedding in validation diagnostics.
func (u AlertingDurationUnits) Describe() string {
	units := u.units
	if units == "" {
		units = allParserUnits
	}
	return renderUnitList(units)
}

// String returns a stable representation used for type identity (Equal) and
// debug output.
func (u AlertingDurationUnits) String() string {
	if u.units == "" {
		return allParserUnits
	}
	return u.units
}

// Equal reports whether two unit sets accept exactly the same units.
func (u AlertingDurationUnits) Equal(o AlertingDurationUnits) bool {
	return u.String() == o.String()
}

// allParserUnits is the master list of units understood by parseAlertingDuration,
// in canonical order. Used as the fallback for zero-value AlertingDurationUnits.
const allParserUnits = "smhdwMy"

// newAlertingDurationUnits constructs a unit set from the given suffix bytes,
// preserving the canonical ordering defined by allParserUnits so that
// String/Equal are stable regardless of caller input order.
func newAlertingDurationUnits(units ...byte) AlertingDurationUnits {
	var b strings.Builder
	for i := 0; i < len(allParserUnits); i++ {
		c := allParserUnits[i]
		for _, u := range units {
			if u == c {
				b.WriteByte(c)
				break
			}
		}
	}
	return AlertingDurationUnits{units: b.String()}
}

// renderUnitList formats a unit string like "smhd" as "`s`, `m`, `h`, or `d`".
func renderUnitList(units string) string {
	switch len(units) {
	case 0:
		return ""
	case 1:
		return fmt.Sprintf("`%c`", units[0])
	case 2:
		return fmt.Sprintf("`%c` or `%c`", units[0], units[1])
	}
	var b strings.Builder
	for i := 0; i < len(units)-1; i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "`%c`", units[i])
	}
	fmt.Fprintf(&b, ", or `%c`", units[len(units)-1])
	return b.String()
}

var (
	// AlertingDurationUnitsSubDay matches the Kibana alerting duration shape
	// used by rule check intervals, throttles, and maintenance-window durations:
	// integers suffixed with `s` (seconds), `m` (minutes), `h` (hours), or
	// `d` (days). Equivalent to the legacy `StringIsAlertingDuration` validator.
	AlertingDurationUnitsSubDay = newAlertingDurationUnits('s', 'm', 'h', 'd')

	// IntervalFrequencyUnits matches Kibana's recurring-schedule interval shape
	// used by maintenance-window `recurring.every`: integers suffixed with
	// `d` (days), `w` (weeks), `M` (months), or `y` (years). Equivalent to the
	// legacy `StringIsMaintenanceWindowIntervalFrequency` validator.
	IntervalFrequencyUnits = newAlertingDurationUnits('d', 'w', 'M', 'y')

	// AllAlertingDurationUnits accepts every unit the parser understands and is
	// primarily useful in tests. Prefer the more specific sets in production
	// schemas.
	AllAlertingDurationUnits = newAlertingDurationUnits('s', 'm', 'h', 'd', 'w', 'M', 'y')
)

// AlertingDurationType is a custom string type for Kibana alerting duration
// strings (for example `1d`, `24h`, `1w`). It extends basetypes.StringType
// with semantic equality so that durations which are numerically equivalent —
// such as `1d` and `24h`, or `1w` and `7d` — are treated as equal by
// Terraform plan/state comparisons. Each attribute selects a Units set to
// declare which suffix letters are valid; validation diagnostics are produced
// by the type itself so schemas no longer need a separate regex validator.
type AlertingDurationType struct {
	basetypes.StringType

	// Units restricts which suffix letters are accepted by ValidateAttribute.
	// Leave empty to accept every unit the parser understands; set to one of
	// the package-level presets (e.g. AlertingDurationUnitsSubDay) to mirror
	// a specific Kibana API contract.
	Units AlertingDurationUnits
}

func (t AlertingDurationType) String() string {
	return fmt.Sprintf("kibanacustomtypes.AlertingDurationType[%s]", t.Units)
}

func (t AlertingDurationType) ValueType(_ context.Context) attr.Value {
	return AlertingDuration{units: t.Units}
}

func (t AlertingDurationType) Equal(o attr.Type) bool {
	other, ok := o.(AlertingDurationType)
	if !ok {
		return false
	}
	if !t.Units.Equal(other.Units) {
		return false
	}
	return t.StringType.Equal(other.StringType)
}

func (t AlertingDurationType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return AlertingDuration{StringValue: in, units: t.Units}, nil
}

func (t AlertingDurationType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}
