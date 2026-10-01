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
	"slices"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/dashboardacctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const optionalRootBlocksResource = "elasticstack_kibana_dashboard.test"

func baseStep(dir, title string) resource.TestStep {
	return resource.TestStep{
		ProtoV6ProviderFactories: acctest.Providers,
		ConfigDirectory:          acctest.NamedTestCaseDirectory(dir),
		ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(title)},
	}
}

func optionalRootBlocksStep(dir, title string, checks ...resource.TestCheckFunc) resource.TestStep {
	step := baseStep(dir, title)
	step.Check = resource.ComposeTestCheckFunc(checks...)
	return step
}

func optionalRootBlocksImportStep(dir, title string, ignore ...string) resource.TestStep {
	step := baseStep(dir, title)
	step.ResourceName = optionalRootBlocksResource
	step.ImportState = true
	step.ImportStateVerify = true
	step.ImportStateVerifyIgnore = ignore
	return step
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
	}, checkRootBlocksNull("time_range", "refresh_interval", "query")...)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			optionalRootBlocksStep("title_only", title, checks...),
			optionalRootBlocksImportStep("title_only", title),
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
			optionalRootBlocksStep("title_only", title, allNull...),
			optionalRootBlocksStep("refresh_interval_only", title,
				resource.TestCheckResourceAttr(optionalRootBlocksResource, "refresh_interval.pause", "false"),
				resource.TestCheckResourceAttr(optionalRootBlocksResource, "refresh_interval.value", "30000"),
			),
			optionalRootBlocksStep("title_only", title, allNull...),
			optionalRootBlocksStep("query_only", title,
				resource.TestCheckResourceAttr(optionalRootBlocksResource, "query.language", "lucene"),
				resource.TestCheckResourceAttr(optionalRootBlocksResource, "query.text", "status:200"),
				resource.TestCheckNoResourceAttr(optionalRootBlocksResource, "query.json"),
			),
			optionalRootBlocksStep("title_only", title, allNull...),
			optionalRootBlocksStep("query_json", title,
				resource.TestCheckResourceAttr(optionalRootBlocksResource, "query.language", "kql"),
				resource.TestCheckResourceAttrSet(optionalRootBlocksResource, "query.json"),
				resource.TestCheckNoResourceAttr(optionalRootBlocksResource, "query.text"),
			),
			optionalRootBlocksStep("title_only", title, allNull...),
			optionalRootBlocksStep("refresh_paused_zero", title,
				resource.TestCheckResourceAttr(optionalRootBlocksResource, "refresh_interval.pause", "true"),
				resource.TestCheckResourceAttr(optionalRootBlocksResource, "refresh_interval.value", "0"),
			),
			optionalRootBlocksStep("title_only", title, allNull...),
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
			optionalRootBlocksImportStep("non_default", title, "time_range.mode"),
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
			optionalRootBlocksStep("title_only_use_time_range", title, slices.Concat(
				[]resource.TestCheckFunc{
					resource.TestCheckResourceAttr(optionalRootBlocksResource, "panels.0.vis_config.by_reference.drilldowns.0.dashboard.use_time_range", "true"),
				},
				checkRootBlocksNull("time_range"),
			)...),
		},
	})
}

// Removing only some blocks keeps the others: the PUT is a full replace of the
// payload the provider sends, so each omitted block is cleared individually.
func TestAccResourceDashboardOptionalRootBlocks_partialRemoval(t *testing.T) {
	title := "Test Partial Removal " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	r := optionalRootBlocksResource
	timeRange := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(r, "time_range.from", "now-24h"),
		resource.TestCheckResourceAttr(r, "time_range.to", "now-1h"),
	}
	refresh := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(r, "refresh_interval.pause", "false"),
		resource.TestCheckResourceAttr(r, "refresh_interval.value", "45000"),
	}
	query := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(r, "query.language", "lucene"),
		resource.TestCheckResourceAttr(r, "query.text", "host.name:web-*"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			optionalRootBlocksStep("non_default", title, slices.Concat(timeRange, refresh, query)...),
			optionalRootBlocksStep("partial_no_query", title,
				slices.Concat(timeRange, refresh, checkRootBlocksNull("query"))...),
			optionalRootBlocksStep("partial_time_range_only", title,
				slices.Concat(timeRange, checkRootBlocksNull("refresh_interval", "query"))...),
		},
	})
}

// time_range.mode lifecycle: switching modes, removing mode while the block
// stays, and removing then re-adding the whole block.
func TestAccResourceDashboardOptionalRootBlocks_timeRangeMode(t *testing.T) {
	title := "Test Time Range Mode " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	r := optionalRootBlocksResource
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			optionalRootBlocksStep("mode_relative", title,
				resource.TestCheckResourceAttr(r, "time_range.mode", "relative")),
			optionalRootBlocksStep("mode_absolute", title,
				resource.TestCheckResourceAttr(r, "time_range.from", "2024-01-01T00:00:00.000Z"),
				resource.TestCheckResourceAttr(r, "time_range.mode", "absolute")),
			// Mode removed while the block stays.
			optionalRootBlocksStep("time_range_only", title,
				resource.TestCheckResourceAttr(r, "time_range.from", "now-7d"),
				resource.TestCheckNoResourceAttr(r, "time_range.mode")),
			// Block with mode -> no block -> block with mode again.
			optionalRootBlocksStep("mode_relative", title,
				resource.TestCheckResourceAttr(r, "time_range.mode", "relative")),
			optionalRootBlocksStep("title_only", title, checkRootBlocksNull("time_range")...),
			optionalRootBlocksStep("mode_relative", title,
				resource.TestCheckResourceAttr(r, "time_range.mode", "relative")),
		},
	})
}
