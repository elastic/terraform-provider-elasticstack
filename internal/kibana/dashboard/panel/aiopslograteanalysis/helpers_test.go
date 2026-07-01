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

package aiopslograteanalysis_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func stringVal(s string) types.String { return types.StringValue(s) }
func boolVal(b bool) types.Bool       { return types.BoolValue(b) }
func stringNull() types.String        { return types.StringNull() }
func boolNull() types.Bool            { return types.BoolNull() }

func configJSONSet(s string) customtypes.JSONWithDefaultsValue[map[string]any] {
	return customtypes.NewJSONWithDefaultsValue(s, func(m map[string]any) map[string]any { return m })
}

func configMap(t *testing.T, item kbapi.DashboardPanelItem) map[string]any {
	t.Helper()
	raw, err := json.Marshal(item)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(raw, &m))
	cfg, ok := m["config"].(map[string]any)
	require.True(t, ok, "config should be object")
	return cfg
}

func diagSummary(diags diag.Diagnostics) string {
	if diags == nil {
		return ""
	}
	var b strings.Builder
	for _, d := range diags {
		b.WriteString(d.Severity().String())
		b.WriteString(": ")
		b.WriteString(d.Summary())
		if dt := d.Detail(); dt != "" {
			b.WriteString(" — ")
			b.WriteString(dt)
		}
		b.WriteString("\n")
	}
	return b.String()
}
