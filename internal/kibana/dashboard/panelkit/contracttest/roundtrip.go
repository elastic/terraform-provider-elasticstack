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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/google/go-cmp/cmp"
)

func appendRoundtripIssues(ctx context.Context, handler iface.Handler, fixture string, skip []string, issues *[]string) {
	item0, err := ParseDashboardPanel(fixture)
	if err != nil {
		*issues = append(*issues, fmt.Sprintf("[RoundTrip] parse fixture: %v", err))
		return
	}

	var pm models.PanelModel
	diags := handler.FromAPI(ctx, &pm, nil, item0)
	if diags.HasError() {
		*issues = append(*issues, fmt.Sprintf("[RoundTrip] FromAPI: %s", summarizeDiags(diags)))
		return
	}

	item1, d2 := handler.ToAPI(pm, nil)
	diags.Append(d2...)
	if diags.HasError() {
		*issues = append(*issues, fmt.Sprintf("[RoundTrip] ToAPI: %s", summarizeDiags(diags)))
		return
	}

	if diff := diffPanelJSON(item0, item1, skip); diff != "" {
		*issues = append(*issues, "[RoundTrip] JSON differs after FromAPI→ToAPI\n"+indentLines(diff, "  "))
	}
}

func diffPanelJSON(a, b kbapi.DashboardPanelItem, skip []string) string {
	raw0, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return "marshal original: " + err.Error()
	}
	raw1, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return "marshal round-tripped: " + err.Error()
	}

	var parsed0, parsed1 any
	if err := json.Unmarshal(raw0, &parsed0); err != nil {
		return "unmarshal original: " + err.Error()
	}
	if err := json.Unmarshal(raw1, &parsed1); err != nil {
		return "unmarshal round-trip: " + err.Error()
	}

	deleteSkipFields(parsed0, skip)
	deleteSkipFields(parsed1, skip)

	return cmp.Diff(parsed0, parsed1)
}

func indentLines(s, prefix string) string {
	if s == "" {
		return ""
	}
	lines := strings.Split(strings.TrimSuffix(s, "\n"), "\n")
	for i := range lines {
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n")
}
