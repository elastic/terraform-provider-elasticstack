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

package datafeedstate_test

// TestAccResourceMLDatafeedState_effectiveSearchEndPopulated verifies that
// effective_search_end is populated with a real value for a started datafeed
// whose running_state.real_time_configured is false. Providing an explicit
// `end` after the indexed document's timestamp makes real_time_configured
// false and gives Elasticsearch a finite lookback whose search_interval.end_ms
// matches that configured end.
//
// Lookback of this short empty range finishes in about a second, so the write
// path must snapshot search_interval immediately after start (see
// asyncutils.WaitForStateTransition). Check functions run against that
// apply-time state. The post-apply refresh plan is non-empty because
// Elasticsearch has then stopped the datafeed (and typically closed the job).
//
// Gated to ES >= 8.1.0 for the same reason as
// TestAccResourceMLDatafeedState_explicitStartPreserved: 8.0.x does not
// reliably expose running_state.search_interval.

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceMLDatafeedState_effectiveSearchEndPopulated(t *testing.T) {
	jobID := fmt.Sprintf("test-job-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	datafeedID := fmt.Sprintf("datafeed-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	indexName := fmt.Sprintf("test-index-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	const docTimestamp = "2022-01-01T00:10:00Z"
	const plannedStart = "2022-01-01T00:00:00Z"
	const plannedEnd = "2022-01-01T01:00:00Z"

	configVars := config.Variables{
		"job_id":      config.StringVariable(jobID),
		"datafeed_id": config.StringVariable(datafeedID),
		"index_name":  config.StringVariable(indexName),
	}

	fullConfigVars := config.Variables{
		"job_id":        config.StringVariable(jobID),
		"datafeed_id":   config.StringVariable(datafeedID),
		"index_name":    config.StringVariable(indexName),
		"planned_start": config.StringVariable(plannedStart),
		"planned_end":   config.StringVariable(plannedEnd),
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			versionutils.SkipIfUnsupported(t, minVersionIssue2353, versionutils.FlavorAny)
		},
		Steps: []resource.TestStep{
			{
				// Step 1: create prerequisite resources (index, job, job state,
				// datafeed). After apply, index a document so the datafeed has
				// data to consume in step 2.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("setup"),
				ConfigVariables:          configVars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index.test", "id"),
					testAccIssue2353IndexDocument(indexName, docTimestamp),
				),
			},
			{
				// Step 2: start the datafeed with an explicit end after the
				// indexed document. real_time_configured is false, and
				// effective_search_end is surfaced from
				// running_state.search_interval.end_ms (the configured end).
				// ExpectNonEmptyPlan: lookback completes before the post-apply
				// refresh, so the refresh plan shows state stopped→started.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("full"),
				ConfigVariables:          fullConfigVars,
				ExpectNonEmptyPlan:       true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlDatafeedStateResourceName, "state", "started"),
					resource.TestCheckResourceAttr(mlDatafeedStateResourceName, "end", plannedEnd),
					resource.TestCheckResourceAttr(mlDatafeedStateResourceName, "effective_search_end", plannedEnd),
				),
			},
		},
	})
}
