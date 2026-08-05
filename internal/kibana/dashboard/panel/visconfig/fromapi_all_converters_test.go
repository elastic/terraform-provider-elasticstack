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

package visconfig_test

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/visconfig"
	"github.com/stretchr/testify/require"
)

// minimalLensChartJSON returns a minimal valid Lens chart payload for each registered kind.
// Fixtures are adapted from internal/kibana/dashboard/lenscommon/detect_test.go (chartKindsPerArm table).
func minimalLensChartJSON(t *testing.T, vizType string) string {
	t.Helper()
	switch vizType {
	case "xy":
		return `{"type":"xy","title":"t","axis":{"x":{},"y":{}},"layers":[{"type":"line","data_source":{"type":"dataView","id":"l"},` +
			`"ignore_global_filters":false,"sampling":1,"y":[{"operation":"count"}]}],"legend":{"visibility":"visible","inside":false,` +
			`"size":"auto"},"filters":[],"styling":{"line":{"curve":"linear"}},"query":{"expression":"*","language":"kql"}}`
	case "treemap":
		return `{"type":"treemap","title":"t","data_source":{"type":"dataView","id":"m"},"query":{"language":"kql","expression":""},` +
			`"legend":{"size":"small"},"metrics":[{"operation":"count"}],"group_by":[{"operation":"terms","field":"host.name","collapse_by":"avg"}]}`
	case "mosaic":
		return `{"type":"mosaic","title":"m","data_source":{"type":"dataView","id":"x"},"query":{"language":"kql","expression":""},` +
			`"legend":{"size":"small"},"metric":{"operation":"count"},"group_by":[{"operation":"terms","collapse_by":"avg",` +
			`"fields":["host.name"],"color":{"mode":"categorical","palette":"default","mapping":[],"unassigned":{"type":"color_code","value":"#D3DAE6"}}}],` +
			`"group_breakdown_by":[{"operation":"terms","collapse_by":"avg","fields":["host.name"],"color":{"mode":"categorical",` +
			`"palette":"default","mapping":[],"unassigned":{"type":"color_code","value":"#D3DAE6"}}}]}`
	case "data_table":
		return `{"type":"data_table","data_source":{"type":"dataView","id":"i"},"query":{"language":"kql","expression":"*"},` +
			`"styling":{"density":{"mode":"default","height":{"header":{"type":"auto"},"value":{"type":"auto"}}}},"metrics":[],` +
			`"time_range":{"from":"now-7d","to":"now"}}`
	case "tag_cloud":
		return `{"type":"tag_cloud","data_source":{"index":"i"},"query":{"expression":"*","language":"kql"},` +
			`"metric":{"operation":{"operation_type":"count"}},"tag_by":{"operation":{"operation_type":"terms"},"field":"t"},` +
			`"styling":{},"filters":[],"time_range":{"from":"now-7d","to":"now"}}`
	case "heatmap":
		return `{"type":"heatmap","data_source":{"type":"dataView","id":"m"},"query":{"expression":"*","language":"kql"},` +
			`"axis":{"x":{},"y":{}},"styling":{"cells":{}},"legend":{"size":"m"},"metric":{"operation":"count"},` +
			`"x":{"operation":"filters","filters":[{"label":"All","filter":{"query":"*","language":"kql"}}]}}`
	case "region_map":
		return `{"type":"region_map","data_source":{"type":"dataView","id":"m"},"query":{"language":"kql","expression":"*"},` +
			`"metric":{"operation":"count"},"region":{"operation":"filters","filters":[{"filter":{"query":"*","language":"kql"},` +
			`"label":"A"}]}}`
	case "legacy_metric":
		return `{"type":"legacy_metric","title":"l","data_source":{"type":"data_view_spec","index_pattern":"m"},` +
			`"query":{"language":"kql","query":"*"},"metric":{"operation":"count","format":{"type":"number"}}}`
	case "metric":
		return `{"type":"metric","title":"M","query":{"expression":"*","language":"kql"},"metrics":[]}`
	case "pie":
		return `{"type":"pie","data_source":{},"query":{"expression":"*","language":"kql"},"styling":{},"metrics":[],"group_by":[]}`
	case "gauge":
		return `{"type":"gauge","data_source":{"type":"dataView","id":"m"},"query":{"expression":"*","language":"kql"},"metric":{"operation":"count"}}`
	case "waffle":
		return `{"type":"waffle","data_source":{"type":"dataView","id":"m"},"query":{"language":"kql","query":""},` +
			`"legend":{"size":"medium","visible":"auto"},"styling":{"values":{}},"metrics":[{"operation":"count"}]}`

	default:
		t.Fatalf("no minimal fixture for viz type %q — add one adapted from lenscommon/detect_test.go", vizType)
		return ""
	}
}

func assertExactlyOneLensChartBlock(t *testing.T, want lenscommon.VizConverter, blocks *models.LensByValueChartBlocks) {
	t.Helper()
	require.True(t, want.HandlesBlocks(blocks), "converter %q should recognize its chart block", want.VizType())
	for _, c := range lenscommon.All() {
		if c.VizType() == want.VizType() {
			continue
		}
		require.False(t, c.HandlesBlocks(blocks), "unexpected secondary chart match from converter %q while testing %q",
			c.VizType(), want.VizType())
	}
}

func TestHandler_FromAPI_byValue_allRegisteredLensCharts(t *testing.T) {
	ctx := iface.WithEnclosingDashboard(context.Background(), &models.DashboardModel{})
	for _, c := range lenscommon.All() {
		t.Run(c.VizType(), func(t *testing.T) {
			item := mustVisPanelItem(t, minimalLensChartJSON(t, c.VizType()))
			var pm models.PanelModel
			diags := visconfig.Handler{}.FromAPI(ctx, &pm, nil, item)
			require.False(t, diags.HasError(), "%s", diags)
			require.NotNil(t, pm.VisConfig)
			require.NotNil(t, pm.VisConfig.ByValue)
			assertExactlyOneLensChartBlock(t, c, &pm.VisConfig.ByValue.LensByValueChartBlocks)
		})
	}
}
