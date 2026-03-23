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

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccResourceMLDatafeedState_basic(t *testing.T) {
	jobID := fmt.Sprintf("test-job-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	datafeedID := fmt.Sprintf("test-datafeed-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	indexName := fmt.Sprintf("test-datafeed-index-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("started"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
					"index_name":  config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed_state.test", "datafeed_id", datafeedID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed_state.test", "state", "started"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed_state.test", "force", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("stopped"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
					"index_name":  config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed_state.test", "datafeed_id", datafeedID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed_state.test", "state", "stopped"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed_state.test", "force", "false"),
				),
			},
		},
	})
}

func TestAccResourceMLDatafeedState_import(t *testing.T) {
	jobID := fmt.Sprintf("test-job-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	datafeedID := fmt.Sprintf("test-datafeed-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	indexName := fmt.Sprintf("test-datafeed-index-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
					"index_name":  config.StringVariable(indexName),
				},
			},
			{
				ProtoV6ProviderFactories:             acctest.Providers,
				ConfigDirectory:                      acctest.NamedTestCaseDirectory("create"),
				ResourceName:                         "elasticstack_elasticsearch_ml_datafeed_state.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "datafeed_id",
				ImportStateVerifyIgnore:              []string{"force", "datafeed_timeout", "id"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["elasticstack_elasticsearch_ml_datafeed_state.test"]
					if !ok {
						return "", fmt.Errorf("not found: %s", "elasticstack_elasticsearch_ml_datafeed_state.test")
					}
					return rs.Primary.Attributes["datafeed_id"], nil
				},
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
					"index_name":  config.StringVariable(indexName),
				},
			},
		},
	})
}

func TestAccResourceMLDatafeedState_withTimes(t *testing.T) {
	jobID := fmt.Sprintf("test-job-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	datafeedID := fmt.Sprintf("test-datafeed-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	indexName := fmt.Sprintf("test-datafeed-index-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_times"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
					"index_name":  config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed_state.test", "datafeed_id", datafeedID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed_state.test", "state", "started"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed_state.test", "start", "2024-01-01T00:00:00Z"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed_state.test", "end", "2024-01-02T00:00:00Z"),
				),
			},
		},
	})
}

// TestAccResourceMLDatafeedState_stoppedThenStarted verifies the plan modifier
// produces correct plan values for `start` during create and state transitions.
//
// The Terraform CLI errors reported in #1866 and #1867 do not reproduce
// consistently in acceptance tests. On main, the plan modifier's early return
// on Create leaves `start` as unknown, but the apply resolves it to null via
// SetStartAndEndFromAPI — an unknown→null resolution that the Terraform
// protocol considers valid. The reported errors appear to be triggered by a
// subtle interaction between the timetypes.RFC3339 custom type and the
// framework's state serialization that does not manifest in a clean test
// environment.
//
// Plan checks verify the fix produces the correct plan values upfront,
// eliminating the dependency on the fragile unknown→null conversion:
//   - Step 1: `start` must be known-null in the plan for a stopped datafeed
//     (without fix: unknown, relying on apply-time resolution).
//   - Step 2: `start` must be unknown in the plan when transitioning to
//     started (without fix: null, inherited from stopped state by
//     UseStateForUnknown before SetUnknownIfStateHasChanges can override it).
func TestAccResourceMLDatafeedState_stoppedThenStarted(t *testing.T) {
	jobID := fmt.Sprintf("test-job-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	datafeedID := fmt.Sprintf("test-datafeed-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	indexName := fmt.Sprintf("test-datafeed-index-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	datafeedStateAddr := "elasticstack_elasticsearch_ml_datafeed_state.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("stopped"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
					"index_name":  config.StringVariable(indexName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(datafeedStateAddr, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(datafeedStateAddr, tfjsonpath.New("start"), knownvalue.Null()),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datafeedStateAddr, "datafeed_id", datafeedID),
					resource.TestCheckResourceAttr(datafeedStateAddr, "state", "stopped"),
					resource.TestCheckNoResourceAttr(datafeedStateAddr, "start"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("started"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
					"index_name":  config.StringVariable(indexName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(datafeedStateAddr, plancheck.ResourceActionUpdate),
						plancheck.ExpectUnknownValue(datafeedStateAddr, tfjsonpath.New("start")),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datafeedStateAddr, "datafeed_id", datafeedID),
					resource.TestCheckResourceAttr(datafeedStateAddr, "state", "started"),
				),
			},
		},
	})
}

func TestAccResourceMLDatafeedState_multiStep(t *testing.T) {
	jobID := fmt.Sprintf("test-job-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	datafeedID := fmt.Sprintf("test-datafeed-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	indexName := fmt.Sprintf("test-datafeed-index-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("closed_stopped"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
					"index_name":  config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed_state.nginx", "state", "stopped"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed_state.nginx", "force", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("job_opened"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
					"index_name":  config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed_state.nginx", "state", "stopped"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed_state.nginx", "force", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("started_no_time"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
					"index_name":  config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed_state.nginx", "state", "started"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed_state.nginx", "force", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("stopped_job_open"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
					"index_name":  config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed_state.nginx", "state", "stopped"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed_state.nginx", "force", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("started_with_time"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
					"index_name":  config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed_state.nginx", "state", "started"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed_state.nginx", "force", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed_state.nginx", "start", "2025-12-01T00:00:00+01:00"),
				),
			},
		},
	})
}
