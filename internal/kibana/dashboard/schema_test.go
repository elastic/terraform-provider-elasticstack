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

package dashboard

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/require"
)

// Panel attributes that participate in sibling mutual exclusion are exactly
// panelConfigNames; structural attributes (type, grid, id) are excluded (OpenSpec task 7.3).
const (
	attrPanelType        = "type"
	attrPanelGrid        = "grid"
	attrPanelID          = "id"
	expectedPanelConfigs = 16 // design D9: API panel kinds + universal config_json (plus image, slo_alerts, discover_session)
)

func Test_panelConfigNames_matchesPanelSchemaAttributes(t *testing.T) {
	names := panelConfigNames()
	require.Len(t, names, expectedPanelConfigs,
		"design D9 expects exactly %d top-level panel-config names", expectedPanelConfigs)

	seenNames := make(map[string]struct{}, len(names))
	for _, n := range names {
		_, dup := seenNames[n]
		require.False(t, dup, "panelConfigNames contains duplicate entry %q", n)
		seenNames[n] = struct{}{}
	}

	panel := getPanelSchema()
	panelConfigKeys := seenNames

	for _, n := range names {
		_, ok := panel.Attributes[n]
		require.True(t, ok,
			"panel object schema missing attribute %q listed in panelConfigNames", n)
	}

	skippedStructural := map[string]struct{}{
		attrPanelType: {},
		attrPanelGrid: {},
		attrPanelID:   {},
	}

	for attrName := range panel.Attributes {
		if _, skip := skippedStructural[attrName]; skip {
			continue
		}
		_, ok := panelConfigKeys[attrName]
		require.True(t, ok,
			"panel schema has panel-config attribute %q that is absent from panelConfigNames", attrName)
	}

	require.Len(t, panel.Attributes, len(skippedStructural)+len(names),
		"panel.Attributes size should equal structural attrs plus panelConfigNames entries")
}

func Test_getPanelSchema_registeredHandlerConfigKeys(t *testing.T) {
	panel := getPanelSchema()
	for _, h := range AllHandlers() {
		key := h.PanelType() + "_config"
		_, ok := panel.Attributes[key]
		require.True(t, ok, "registry handler schema missing panel attribute %q", key)
		_, inner := panel.Attributes[key].(schema.SingleNestedAttribute)
		require.True(t, inner, "handler %q SchemaAttribute must be SingleNestedAttribute for key %s", h.PanelType(), key)
	}
}

func Test_sibling_conflict_paths_exclude_only_self(t *testing.T) {
	names := panelConfigNames()
	for _, h := range AllHandlers() {
		self := h.PanelType() + "_config"
		pathExprs := panelkit.SiblingTypedPanelConfigConflictPathsExcept(self, names)
		require.Len(t, pathExprs, len(names)-1, "panel %q sibling conflict coverage", h.PanelType())
	}
}
