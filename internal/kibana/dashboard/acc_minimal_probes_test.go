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

package dashboard_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/dashboardacctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// The probes in this file deliberately omit every optional attribute on each
// typed Lens panel block to flush out latent "Provider produced inconsistent
// result after apply" diagnostics caused by Kibana injecting server-side
// defaults that the plan never wrote. The pattern that originally surfaced
// these issues is documented in REQ-011 of openspec/specs/kibana-dashboard
// and tracked in https://github.com/elastic/terraform-provider-elasticstack/issues/3402.
//
// Each test does an apply followed by a plan-only step so both failure modes
// are caught:
//   - inconsistent-result errors during apply (write-time)
//   - drift on a subsequent plan (read-time)

const minimalProbeDashboardPrefix = "Test Lens Minimal Probe "

func runMinimalLensProbe(t *testing.T, title string, body string) {
	t.Helper()
	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	config := fmt.Sprintf(`
resource "elasticstack_kibana_dashboard" "probe" {
  title = %q
  time_range = { from = "now-15m", to = "now" }
  refresh_interval = { pause = true, value = 0 }
  query = { language = "kql", text = "" }
  panels = [{
    type = "vis"
    grid = { x = 0, y = 0, w = 24, h = 15 }
    vis_config = {
      by_value = {
%s
      }
    }
  }]
}
`, title, body)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				Config:                   config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.probe", "id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				Config:                   config,
				PlanOnly:                 true,
			},
		},
	})
}

func newProbeTitle(suffix string) string {
	return minimalProbeDashboardPrefix + suffix + " " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
}

func TestAccLensMinimalProbe_LegacyMetric(t *testing.T) {
	runMinimalLensProbe(t, newProbeTitle("LegacyMetric"), `
legacy_metric_config = {
  data_source_json = jsonencode({ type = "data_view_spec", index_pattern = "metrics-*" })
  query            = { expression = "" }
  metric_json      = jsonencode({ operation = "count" })
}
`)
}

func TestAccLensMinimalProbe_Gauge(t *testing.T) {
	runMinimalLensProbe(t, newProbeTitle("Gauge"), `
gauge_config = {
  data_source_json = jsonencode({ type = "data_view_spec", index_pattern = "metrics-*" })
  query            = { expression = "" }
  metric_json      = jsonencode({ operation = "count" })
  styling          = {}
}
`)
}

func TestAccLensMinimalProbe_Heatmap(t *testing.T) {
	runMinimalLensProbe(t, newProbeTitle("Heatmap"), `
heatmap_config = {
  data_source_json = jsonencode({ type = "data_view_spec", index_pattern = "metrics-*" })
  query            = { expression = "" }
  metric_json      = jsonencode({ operation = "count" })
  x_axis_json      = jsonencode({
    operation = "filters"
    filters   = [{ label = "All", filter = { expression = "*", language = "kql" } }]
  })
  axis = {
    x = { title = { value = "X" }, labels = {} }
    y = { title = { value = "Y" }, labels = {} }
  }
  styling = { cells = { labels = {} } }
  legend  = { size = "m" }
}
`)
}

func TestAccLensMinimalProbe_Tagcloud(t *testing.T) {
	runMinimalLensProbe(t, newProbeTitle("Tagcloud"), `
tagcloud_config = {
  data_source_json = jsonencode({ type = "data_view_spec", index_pattern = "metrics-*" })
  query            = { expression = "" }
  metric_json      = jsonencode({ operation = "count" })
  tag_by_json      = jsonencode({ operation = "terms", fields = ["host.name"], limit = 10 })
}
`)
}

func TestAccLensMinimalProbe_RegionMap(t *testing.T) {
	runMinimalLensProbe(t, newProbeTitle("RegionMap"), `
region_map_config = {
  data_source_json = jsonencode({ type = "data_view_spec", index_pattern = "metrics-*" })
  query            = { expression = "" }
  metric_json      = jsonencode({ operation = "count" })
  region_json      = jsonencode({
    operation = "filters"
    filters   = [{ label = "All", filter = { expression = "*", language = "kql" } }]
  })
}
`)
}

func TestAccLensMinimalProbe_Datatable(t *testing.T) {
	runMinimalLensProbe(t, newProbeTitle("Datatable"), `
datatable_config = {
  no_esql = {
    data_source_json = jsonencode({ type = "data_view_spec", index_pattern = "metrics-*" })
    query            = { expression = "" }
    metrics          = [{ config_json = jsonencode({ operation = "count" }) }]
    styling          = { density = { mode = "default" } }
  }
}
`)
}

func TestAccLensMinimalProbe_Pie(t *testing.T) {
	runMinimalLensProbe(t, newProbeTitle("Pie"), `
pie_chart_config = {
  data_source_json = jsonencode({ type = "data_view_spec", index_pattern = "metrics-*" })
  query            = { expression = "" }
  metrics          = [{ config_json = jsonencode({ operation = "count" }) }]
  group_by         = [{ config_json = jsonencode({ operation = "terms", fields = ["host.name"], limit = 5 }) }]
}
`)
}

func TestAccLensMinimalProbe_Treemap(t *testing.T) {
	runMinimalLensProbe(t, newProbeTitle("Treemap"), `
treemap_config = {
  data_source_json = jsonencode({ type = "data_view_spec", index_pattern = "metrics-*" })
  query            = { expression = "" }
  group_by_json    = jsonencode([{ operation = "terms", fields = ["host.name"], limit = 5 }])
  metrics_json     = jsonencode([{ operation = "count" }])
  legend           = { size = "m" }
}
`)
}

func TestAccLensMinimalProbe_Mosaic(t *testing.T) {
	runMinimalLensProbe(t, newProbeTitle("Mosaic"), `
mosaic_config = {
  data_source_json        = jsonencode({ type = "data_view_spec", index_pattern = "metrics-*" })
  query                   = { expression = "" }
  group_by_json           = jsonencode([{ operation = "terms", fields = ["host.name"], limit = 5 }])
  group_breakdown_by_json = jsonencode([{ operation = "terms", fields = ["service.name"], limit = 5 }])
  metrics_json            = jsonencode([{ operation = "count" }])
  legend                  = { size = "m" }
}
`)
}

func TestAccLensMinimalProbe_Waffle(t *testing.T) {
	runMinimalLensProbe(t, newProbeTitle("Waffle"), `
waffle_config = {
  data_source_json = jsonencode({ type = "data_view_spec", index_pattern = "metrics-*" })
  query            = { expression = "" }
  metrics          = [{ config_json = jsonencode({ operation = "count" }) }]
  legend           = { size = "m" }
}
`)
}

func TestAccLensMinimalProbe_Metric(t *testing.T) {
	runMinimalLensProbe(t, newProbeTitle("Metric"), `
metric_chart_config = {
  data_source_json = jsonencode({ type = "data_view_spec", index_pattern = "metrics-*" })
  query            = { expression = "" }
  metrics          = [{
    config_json = jsonencode({
      type      = "primary"
      operation = "count"
      format    = { type = "number" }
    })
  }]
}
`)
}
