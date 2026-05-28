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
//
// NOTE: All probes here currently FAIL against a real Kibana 9.4+ stack — they
// are committed as TODO/regression markers. Each test calls t.Skip with the
// specific field paths Kibana injects. Remove the Skip in the test guarding a
// given panel's fix to turn the probe back on as a passing regression test.

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
	t.Skip("TODO: data_source_json drift — Kibana injects time_field=\"@timestamp\" on read-back; see acc_minimal_probes_test.go banner")
	runMinimalLensProbe(t, newProbeTitle("LegacyMetric"), `
legacy_metric_config = {
  data_source_json = jsonencode({ type = "data_view_spec", index_pattern = "metrics-*" })
  query            = { expression = "" }
  metric_json      = jsonencode({ operation = "count" })
}
`)
}

func TestAccLensMinimalProbe_Gauge(t *testing.T) {
	t.Skip("TODO: data_source_json drift (time_field) + styling.shape_json default {\"type\":\"bullet\"} injected by Kibana when omitted")
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
	t.Skip("TODO: data_source_json drift + axis.*.labels.visible, axis.*.title.visible, styling.cells.labels.visible, legend.visibility server defaults need null-preservation")
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
	t.Skip("TODO: data_source_json drift + orientation=\"horizontal\" and font_size={min=18,max=72} injected by Kibana when omitted")
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
	t.Skip("TODO: data_source_json drift — Kibana injects time_field=\"@timestamp\" on read-back")
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
	t.Skip("TODO: data_source_json drift + metrics[*].config_json default normalization (color, empty_as_null, format.{decimals,compact}) needed")
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
	t.Skip("TODO: data_source_json drift + label_position=\"outside\" default + group_by[*].config_json default normalization (color, rank_by, limit)")
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
	t.Skip("TODO: data_source_json drift + legend.visible=\"auto\" + value_display={mode=\"percentage\"} block + group_by_json default normalization (color, rank_by)")
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
	t.Skip("TODO: data_source_json drift + legend.visible=\"auto\" + value_display={mode=\"percentage\"} block + group_by_json / group_breakdown_by_json default normalization")
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
	t.Skip("TODO: data_source_json drift — Kibana injects time_field=\"@timestamp\" on read-back")
	runMinimalLensProbe(t, newProbeTitle("Waffle"), `
waffle_config = {
  data_source_json = jsonencode({ type = "data_view_spec", index_pattern = "metrics-*" })
  query            = { expression = "" }
  metrics          = [{ config_json = jsonencode({ operation = "count" }) }]
  legend           = { size = "m" }
}
`)
}

// TestAccLensMinimalProbe_Metric is the baseline — the lensmetric panel was
// already hardened for this class of bug in #2355, so its minimal probe
// passes. It is intentionally not skipped and serves as a control showing
// the probe harness itself is correct.
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
