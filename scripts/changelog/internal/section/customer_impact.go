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

package section

import (
	"encoding/json"
	"fmt"
)

const (
	impactLiteralNone        = "none"
	impactLiteralFix         = "fix"
	impactLiteralEnhancement = "enhancement"
	impactLiteralBreaking    = "breaking"
)

// CustomerImpact mirrors the changelog contract's discrete impact values.
// Explicit "Customer impact omitted" differs from Customer impact: none; use ImpactPresent and/or ImpactRaw
// alongside the enum (ImpactNone mirrors none but is ambiguous when ImpactPresent is false).
type CustomerImpact int

const (
	ImpactNone CustomerImpact = iota
	ImpactFix
	ImpactEnhancement
	ImpactBreaking
)

var customerImpactStrings = map[CustomerImpact]string{
	ImpactNone:        impactLiteralNone,
	ImpactFix:         impactLiteralFix,
	ImpactEnhancement: impactLiteralEnhancement,
	ImpactBreaking:    impactLiteralBreaking,
}

func customerImpactIDs() []string {
	return []string{impactLiteralNone, impactLiteralFix, impactLiteralEnhancement, impactLiteralBreaking}
}

// ParseCustomerImpact maps a changelog line value to CustomerImpact when it is exactly one of
// none|fix|enhancement|breaking (case-sensitive, per workflow contract parity with JS helpers).
func ParseCustomerImpact(s string) (CustomerImpact, bool) {
	switch s {
	case impactLiteralNone:
		return ImpactNone, true
	case impactLiteralFix:
		return ImpactFix, true
	case impactLiteralEnhancement:
		return ImpactEnhancement, true
	case impactLiteralBreaking:
		return ImpactBreaking, true
	default:
		return ImpactNone, false
	}
}

// RequiresSummary mirrors validateChangelogSection: summary is required unless impact is exactly "none".
func (i CustomerImpact) RequiresSummary() bool {
	return i != ImpactNone
}

func (i CustomerImpact) String() string {
	if s, ok := customerImpactStrings[i]; ok {
		return s
	}
	return ""
}

func (i CustomerImpact) MarshalJSON() ([]byte, error) {
	s := i.String()
	if s == "" {
		return nil, fmt.Errorf("section: cannot marshal unknown CustomerImpact value %d", int(i))
	}
	return json.Marshal(s)
}

func (i *CustomerImpact) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	v, ok := ParseCustomerImpact(s)
	if !ok {
		return fmt.Errorf("section: invalid CustomerImpact JSON value %q", s)
	}
	*i = v
	return nil
}
