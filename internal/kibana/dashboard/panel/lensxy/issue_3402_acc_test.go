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

package lensxy_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/dashboardacctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccReproduceIssue3402 reproduces the bug where using bar_stacked as the
// XY chart layer type with fitting.type = "none" causes Kibana to return an
// empty string for fitting.type on read-back, leading to
// "Provider produced inconsistent result after apply".
//
// Root cause: xyFittingFromAPI uses typeutils.StringishValue (which maps "" to
// types.StringValue("") — a known non-null value) instead of
// NonEmptyStringishValue. The alignment helper alignXYFittingStateFromPlan only
// restores plan values when state is null, so the empty-string return slips
// through unchanged.
//
// Related to: https://github.com/elastic/terraform-provider-elasticstack/issues/3402
func TestAccReproduceIssue3402(t *testing.T) {
	dashboardTitle := "Repro Issue 3402 bar_stacked " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				// When bar_stacked is used, Kibana returns fitting.type="" on read-back
				// even though "none" was written. The provider detects the plan/state
				// divergence and reports "Provider produced inconsistent result after apply".
				ProtoV6ProviderFactories: acctest.Providers,
				Config:                   testAccIssue3402Config(dashboardTitle),
				ExpectError:              regexp.MustCompile(`inconsistent result after apply`),
			},
		},
	})
}

func testAccIssue3402Config(title string) string {
	return fmt.Sprintf(`
resource "elasticstack_kibana_dashboard" "repro_3402" {
  title = %q
  time_range = {
    from = "now-15m"
    to   = "now"
  }
  refresh_interval = {
    pause = true
    value = 0
  }
  query = {
    language = "kql"
    text     = ""
  }
  panels = [{
    type = "vis"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 15
    }
    vis_config = {
      by_value = {
        xy_chart_config = {
          axis = {
            y = {
              domain_json = jsonencode({ type = "fit" })
            }
          }
          decorations = {}
          fitting = {
            type = "none"
          }
          layers = [{
            type = "bar_stacked"
            data_layer = {
              data_source_json = jsonencode({
                type          = "data_view_spec"
                index_pattern = "metrics-*"
              })
              y = [{
                config_json = jsonencode({
                  operation     = "count"
                  empty_as_null = true
                })
              }]
            }
          }]
          legend = {}
          query  = { expression = "" }
        }
      }
    }
  }]
}
`, title)
}
