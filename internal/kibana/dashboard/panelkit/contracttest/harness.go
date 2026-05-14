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

package contracttest

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
)

// Config holds a raw dashboard panel fixture (one JSON object matching kbapi DashboardPanelItem).
type Config struct {
	FullAPIResponse string
	SkipFields      []string
	// OmitValidateRequiredZero skips appendValidateRequiredZeroIssues entirely for this panel (not per-attribute).
	//
	// Use only when the panel has a shallow required Terraform attribute whose type is non-scalar (list, map, or
	// object) and attrsForShallowFixture cannot synthesize a typed null/zero from the raw JSON fixture.
	//
	// TODO(contracttest): generalize required-zero synthesis for list/map/object shallow required attributes so
	// callers need not set OmitValidateRequiredZero; then deprecate this field.
	OmitValidateRequiredZero bool
	// OmitRequiredLeafPresence disables the required-leaf-presence harness phase that checks every required Terraform leaf
	// path against raw fixture.config navigation. Enable only when the panel's exported API JSON layout does not align with
	// Terraform's nested paths (flat config keys vs nested single|groups, or optional filters/drilldown trees only present
	// in schema) such that navigateStructByTFSegments cannot resolve the fixture shape to those paths yet. Each caller
	// must document the specific JSON-vs-HCL divergence beside the literal; see todo in contracttest/tfreflect.go.
	OmitRequiredLeafPresence bool
}

func Run(t *testing.T, handler iface.Handler, cfg Config) {
	t.Helper()
	ctx := context.Background()
	for _, msg := range runChecks(ctx, handler, cfg) {
		t.Error(msg)
	}
}

func runChecks(ctx context.Context, handler iface.Handler, cfg Config) []string {
	block := handler.PanelType() + "_config"

	if cfg.FullAPIResponse == "" {
		return []string{"[Harness] FullAPIResponse must be non-empty JSON"}
	}
	if _, err := ParseDashboardPanel(cfg.FullAPIResponse); err != nil {
		return []string{"[Harness] parse FullAPIResponse: " + err.Error()}
	}

	var issues []string

	appendOuterSchemaIssues(handler, &issues)
	if !cfg.OmitRequiredLeafPresence {
		appendRequiredJSONPresenceIssues(handler, cfg.FullAPIResponse, &issues)
	}
	if panelkit.HasPanelConfigBlock(block) && !cfg.OmitValidateRequiredZero {
		appendValidateRequiredZeroIssues(ctx, handler, cfg.FullAPIResponse, &issues)
	}

	appendRoundtripIssues(ctx, handler, cfg.FullAPIResponse, cfg.SkipFields, &issues)

	appendReflectIssues(ctx, handler, cfg.FullAPIResponse, &issues)

	appendNullPreserveIssues(ctx, handler, cfg.FullAPIResponse, cfg.SkipFields, &issues)

	return issues
}
