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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/dashboardacctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

const optionalRootBlocksResource = "elasticstack_kibana_dashboard.test"

func optionalRootBlocksStep(dir, title string, checks ...resource.TestCheckFunc) resource.TestStep {
	return resource.TestStep{
		ProtoV6ProviderFactories: acctest.Providers,
		ConfigDirectory:          acctest.NamedTestCaseDirectory(dir),
		ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(title)},
		Check:                    resource.ComposeTestCheckFunc(checks...),
	}
}

func optionalRootBlocksPlanOnlyStep(dir, title string) resource.TestStep {
	return resource.TestStep{
		ProtoV6ProviderFactories: acctest.Providers,
		ConfigDirectory:          acctest.NamedTestCaseDirectory(dir),
		ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(title)},
		ConfigPlanChecks: resource.ConfigPlanChecks{
			PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
		},
	}
}

func checkRootBlocksNull(blocks ...string) []resource.TestCheckFunc {
	checks := make([]resource.TestCheckFunc, 0, len(blocks))
	for _, b := range blocks {
		checks = append(checks, resource.TestCheckNoResourceAttr(optionalRootBlocksResource, b+".%"))
	}
	return checks
}

// A title-only dashboard is valid, applies without inconsistent-result errors,
// re-plans empty, and imports without drift. This is also the tripwire for a
// future Kibana version that starts returning defaults for the root blocks.
func TestAccResourceDashboardOptionalRootBlocks_titleOnly(t *testing.T) {
	title := "Test Title Only " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	checks := append([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(optionalRootBlocksResource, "id"),
		resource.TestCheckResourceAttr(optionalRootBlocksResource, "title", title),
		resource.TestCheckNoResourceAttr(optionalRootBlocksResource, "time_range.from"),
		resource.TestCheckNoResourceAttr(optionalRootBlocksResource, "refresh_interval.pause"),
		resource.TestCheckNoResourceAttr(optionalRootBlocksResource, "query.language"),
	}, checkRootBlocksNull("time_range", "refresh_interval", "query")...)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			optionalRootBlocksStep("title_only", title, checks...),
			optionalRootBlocksPlanOnlyStep("title_only", title),
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("title_only"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(title)},
				ResourceName:             optionalRootBlocksResource,
				ImportState:              true,
				ImportStateVerify:        true,
			},
		},
	})
}

// Each root block can be added to a title-only dashboard and later removed.
// Removal clears the block (PUT is a full replace) rather than resetting it to
// a Kibana default, so the block is null in state afterwards.
func TestAccResourceDashboardOptionalRootBlocks_addRemove(t *testing.T) {
	title := "Test Add Remove " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	allNull := checkRootBlocksNull("time_range", "refresh_interval", "query")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			optionalRootBlocksStep("title_only", title, allNull...),
			optionalRootBlocksStep("time_range_only", title,
				resource.TestCheckResourceAttr(optionalRootBlocksResource, "time_range.from", "now-7d"),
				resource.TestCheckResourceAttr(optionalRootBlocksResource, "time_range.to", "now"),
			),
			optionalRootBlocksPlanOnlyStep("time_range_only", title),
			optionalRootBlocksStep("title_only", title, allNull...),
			optionalRootBlocksStep("refresh_interval_only", title,
				resource.TestCheckResourceAttr(optionalRootBlocksResource, "refresh_interval.pause", "false"),
				resource.TestCheckResourceAttr(optionalRootBlocksResource, "refresh_interval.value", "30000"),
			),
			optionalRootBlocksPlanOnlyStep("refresh_interval_only", title),
			optionalRootBlocksStep("title_only", title, allNull...),
			optionalRootBlocksStep("query_only", title,
				resource.TestCheckResourceAttr(optionalRootBlocksResource, "query.language", "lucene"),
				resource.TestCheckResourceAttr(optionalRootBlocksResource, "query.text", "status:200"),
			),
			optionalRootBlocksPlanOnlyStep("query_only", title),
			optionalRootBlocksStep("title_only", title, allNull...),
			optionalRootBlocksPlanOnlyStep("title_only", title),
		},
	})
}

// Non-default values for every root block round-trip without drift.
func TestAccResourceDashboardOptionalRootBlocks_nonDefaultValues(t *testing.T) {
	title := "Test Non Default " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			optionalRootBlocksStep("non_default", title,
				resource.TestCheckResourceAttr(optionalRootBlocksResource, "time_range.from", "now-24h"),
				resource.TestCheckResourceAttr(optionalRootBlocksResource, "time_range.to", "now-1h"),
				resource.TestCheckResourceAttr(optionalRootBlocksResource, "refresh_interval.pause", "false"),
				resource.TestCheckResourceAttr(optionalRootBlocksResource, "refresh_interval.value", "45000"),
				resource.TestCheckResourceAttr(optionalRootBlocksResource, "query.language", "lucene"),
				resource.TestCheckResourceAttr(optionalRootBlocksResource, "query.text", "host.name:web-*"),
			),
			optionalRootBlocksPlanOnlyStep("non_default", title),
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("non_default"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(title)},
				ResourceName:             optionalRootBlocksResource,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore:  []string{"time_range.mode"},
			},
		},
	})
}

// A vis panel with a dashboard drilldown using use_time_range on a title-only
// dashboard (no root time_range) applies and re-plans empty.
func TestAccResourceDashboardOptionalRootBlocks_titleOnlyPanelUseTimeRange(t *testing.T) {
	title := "Test Title Only Panel " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			optionalRootBlocksStep("title_only_use_time_range", title,
				resource.TestCheckResourceAttr(optionalRootBlocksResource, "panels.0.vis_config.by_reference.drilldowns.0.dashboard.use_time_range", "true"),
				resource.TestCheckNoResourceAttr(optionalRootBlocksResource, "time_range.%"),
			),
			optionalRootBlocksPlanOnlyStep("title_only_use_time_range", title),
		},
	})
}
