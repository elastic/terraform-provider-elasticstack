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
	"os"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const mlDatafeedStateResourceName = "elasticstack_elasticsearch_ml_datafeed_state.test"

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
					resource.TestCheckResourceAttr(mlDatafeedStateResourceName, "datafeed_id", datafeedID),
					resource.TestCheckResourceAttr(mlDatafeedStateResourceName, "state", "started"),
					resource.TestCheckResourceAttr(mlDatafeedStateResourceName, "force", "false"),
					resource.TestCheckResourceAttrSet(mlDatafeedStateResourceName, "id"),
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
					resource.TestCheckResourceAttr(mlDatafeedStateResourceName, "datafeed_id", datafeedID),
					resource.TestCheckResourceAttr(mlDatafeedStateResourceName, "state", "stopped"),
					resource.TestCheckResourceAttr(mlDatafeedStateResourceName, "force", "false"),
					resource.TestCheckResourceAttrSet(mlDatafeedStateResourceName, "id"),
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
				ResourceName:                         mlDatafeedStateResourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "datafeed_id",
				ImportStateVerifyIgnore:              []string{"force", "datafeed_timeout", "id"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[mlDatafeedStateResourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", mlDatafeedStateResourceName)
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
	resourceName := mlDatafeedStateResourceName

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
					resource.TestCheckResourceAttr(resourceName, "datafeed_id", datafeedID),
					resource.TestCheckResourceAttr(resourceName, "state", "started"),
					resource.TestCheckResourceAttr(resourceName, "start", "2024-01-01T00:00:00Z"),
					resource.TestCheckResourceAttr(resourceName, "end", "2024-01-02T00:00:00Z"),
					resource.TestCheckResourceAttr(resourceName, "datafeed_timeout", "60s"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_times_updated_timeout"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
					"index_name":  config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", "stopped"),
					resource.TestCheckResourceAttr(resourceName, "datafeed_timeout", "90s"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
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

// TestAccResourceMLDatafeedState_timesMultiStep exercises multi-step coverage for
// start/end values across stopped-state transitions. It verifies that start/end
// are stored when set, can be updated, and that end is absent from state when
// removed from config.
func TestAccResourceMLDatafeedState_timesMultiStep(t *testing.T) {
	jobID := fmt.Sprintf("test-job-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	datafeedID := fmt.Sprintf("test-datafeed-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	indexName := fmt.Sprintf("test-datafeed-index-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	resourceName := mlDatafeedStateResourceName

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("initial_with_times"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
					"index_name":  config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", "started"),
					resource.TestCheckResourceAttr(resourceName, "start", "2022-01-01T00:00:00Z"),
					resource.TestCheckResourceAttr(resourceName, "end", "2022-01-31T00:00:00Z"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("stopped_1"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
					"index_name":  config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", "stopped"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("new_times"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
					"index_name":  config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", "stopped"),
					resource.TestCheckResourceAttr(resourceName, "start", "2023-06-01T00:00:00Z"),
					resource.TestCheckResourceAttr(resourceName, "end", "2023-06-30T00:00:00Z"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("end_omitted"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
					"index_name":  config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", "stopped"),
					resource.TestCheckNoResourceAttr(resourceName, "end"),
				),
			},
		},
	})
}

// TestAccResourceMLDatafeedState_timeouts verifies that the timeouts.create and
// timeouts.update attributes round-trip correctly through create and update.
func TestAccResourceMLDatafeedState_timeouts(t *testing.T) {
	jobID := fmt.Sprintf("test-job-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	datafeedID := fmt.Sprintf("test-datafeed-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	indexName := fmt.Sprintf("test-datafeed-index-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	resourceName := mlDatafeedStateResourceName

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_timeouts"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
					"index_name":  config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", "stopped"),
					resource.TestCheckResourceAttr(resourceName, "timeouts.create", "5m"),
					resource.TestCheckResourceAttr(resourceName, "timeouts.update", "5m"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_updated_timeouts"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
					"index_name":  config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", "stopped"),
					resource.TestCheckResourceAttr(resourceName, "timeouts.create", "10m"),
					resource.TestCheckResourceAttr(resourceName, "timeouts.update", "10m"),
				),
			},
		},
	})
}

// TestAccResourceMLDatafeedState_explicitConnection exercises the
// elasticsearch_connection block on the resource, asserting that the endpoint
// and insecure fields are stored correctly. The datafeed is kept in the
// "stopped" state so the API state always matches the config, making
// ImportStateVerify reliable without needing to ignore the "state" attribute.
func TestAccResourceMLDatafeedState_explicitConnection(t *testing.T) {
	endpoints := testAccMLDatafeedStateESEndpoints()
	if len(endpoints) == 0 {
		t.Skip("ELASTICSEARCH_ENDPOINTS must be set to run this test")
	}
	endpointVars := make([]config.Variable, 0, len(endpoints))
	for _, ep := range endpoints {
		endpointVars = append(endpointVars, config.StringVariable(ep))
	}

	jobID := fmt.Sprintf("test-job-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	datafeedID := fmt.Sprintf("test-datafeed-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	indexName := fmt.Sprintf("test-datafeed-index-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	resourceName := mlDatafeedStateResourceName

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
					"endpoints":   config.ListVariable(endpointVars...),
					"api_key":     config.StringVariable(os.Getenv("ELASTICSEARCH_API_KEY")),
					"username":    config.StringVariable(os.Getenv("ELASTICSEARCH_USERNAME")),
					"password":    config.StringVariable(os.Getenv("ELASTICSEARCH_PASSWORD")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "datafeed_id", datafeedID),
					resource.TestCheckResourceAttr(resourceName, "state", "stopped"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.0.endpoints.#", fmt.Sprintf("%d", len(endpoints))),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.0.endpoints.0", endpoints[0]),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.0.insecure", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				ProtoV6ProviderFactories:             acctest.Providers,
				ConfigDirectory:                      acctest.NamedTestCaseDirectory("stopped"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIgnore:              []string{"elasticsearch_connection", "force", "datafeed_timeout", "id"},
				ImportStateVerifyIdentifierAttribute: "datafeed_id",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}
					return rs.Primary.Attributes["datafeed_id"], nil
				},
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
					"index_name":  config.StringVariable(indexName),
					"endpoints":   config.ListVariable(endpointVars...),
					"api_key":     config.StringVariable(os.Getenv("ELASTICSEARCH_API_KEY")),
					"username":    config.StringVariable(os.Getenv("ELASTICSEARCH_USERNAME")),
					"password":    config.StringVariable(os.Getenv("ELASTICSEARCH_PASSWORD")),
				},
			},
		},
	})
}

func testAccMLDatafeedStateESEndpoints() []string {
	rawEndpoints := os.Getenv("ELASTICSEARCH_ENDPOINTS")
	parts := strings.Split(rawEndpoints, ",")
	endpoints := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			endpoints = append(endpoints, part)
		}
	}
	return endpoints
}
