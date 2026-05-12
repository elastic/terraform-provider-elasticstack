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

	"github.com/stretchr/testify/require"
)

// Panel attributes that participate in sibling mutual exclusion are exactly
// panelConfigNames; structural attributes (type, grid, id) are excluded (OpenSpec task 7.3).
const (
	attrPanelType        = "type"
	attrPanelGrid        = "grid"
	attrPanelID          = "id"
	expectedPanelConfigs = 13 // design D9: API panel kinds + universal config_json
)

func Test_panelConfigNames_matchesPanelSchemaAttributes(t *testing.T) {
	require.Len(t, panelConfigNames, expectedPanelConfigs,
		"design D9 expects exactly %d top-level panel-config names", expectedPanelConfigs)

	seenNames := make(map[string]struct{}, len(panelConfigNames))
	for _, n := range panelConfigNames {
		_, dup := seenNames[n]
		require.False(t, dup, "panelConfigNames contains duplicate entry %q", n)
		seenNames[n] = struct{}{}
	}

	panel := getPanelSchema()
	panelConfigKeys := seenNames // same keys

	for _, n := range panelConfigNames {
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

	require.Len(t, panel.Attributes, len(skippedStructural)+len(panelConfigNames),
		"panel.Attributes size should equal structural attrs plus panelConfigNames entries")
}
