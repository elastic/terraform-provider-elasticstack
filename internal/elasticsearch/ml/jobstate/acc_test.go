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

package jobstate_test

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceMLJobState(t *testing.T) {
	jobID := fmt.Sprintf("test-ml-job-state-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("opened"),
				ConfigVariables: config.Variables{
					"job_id": config.StringVariable(jobID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "job_id", jobID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "state", "opened"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "force", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "job_timeout", "30s"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_job_state.test", "id"),
					// Verify that the ML job was created by the anomaly detector resource
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "job_id", jobID),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("closed"),
				ConfigVariables: config.Variables{
					"job_id": config.StringVariable(jobID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "job_id", jobID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "state", "closed"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "force", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "job_timeout", "30s"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_job_state.test", "id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("opened_with_options"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"force":       config.BoolVariable(true),
					"job_timeout": config.StringVariable("1m"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "job_id", jobID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "state", "opened"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "force", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "job_timeout", "1m"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_job_state.test", "id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("opened_with_options"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"force":       config.BoolVariable(true),
					"job_timeout": config.StringVariable("1m"),
				},
				ResourceName:            "elasticstack_elasticsearch_ml_job_state.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force", "job_timeout", "timeouts"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["elasticstack_elasticsearch_ml_job_state.test"]
					return rs.Primary.ID, nil
				},
			},
		},
	})
}

func TestAccResourceMLJobStateNonExistent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("non_existent"),
				ExpectError:              regexp.MustCompile(`ML job .* does not exist`),
			},
		},
	})
}

func TestAccResourceMLJobStateImport(t *testing.T) {
	jobID := fmt.Sprintf("test-ml-job-state-import-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("opened"),
				ConfigVariables: config.Variables{
					"job_id": config.StringVariable(jobID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "job_id", jobID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "state", "opened"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "force", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "job_timeout", "30s"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("opened"),
				ConfigVariables: config.Variables{
					"job_id": config.StringVariable(jobID),
				},
				ResourceName:      "elasticstack_elasticsearch_ml_job_state.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["elasticstack_elasticsearch_ml_job_state.test"]
					return rs.Primary.ID, nil
				},
			},
		},
	})
}

// TestAccResourceMLJobStateImportWithOptions applies non-default force and job_timeout, then
// imports: Read only backfills null force/timeout to schema defaults, so import verify ignores
// those (same pattern as ml data feed state import tests).
func TestAccResourceMLJobStateImportWithOptions(t *testing.T) {
	jobID := fmt.Sprintf("test-ml-job-state-import-opts-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	resourceName := "elasticstack_elasticsearch_ml_job_state.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("opened_with_options"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"force":       config.BoolVariable(true),
					"job_timeout": config.StringVariable("1m"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "job_id", jobID),
					resource.TestCheckResourceAttr(resourceName, "state", "opened"),
					resource.TestCheckResourceAttr(resourceName, "force", "true"),
					resource.TestCheckResourceAttr(resourceName, "job_timeout", "1m"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("opened_with_options"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"force":       config.BoolVariable(true),
					"job_timeout": config.StringVariable("1m"),
				},
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force", "job_timeout", "timeouts"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources[resourceName]
					return rs.Primary.ID, nil
				},
			},
		},
	})
}

// TestAccResourceMLJobStateExplicitConnection exercises the elasticsearch_connection block
// (scoped Elasticsearch client) while the anomaly job still uses the default connection.
// After create+import, it updates state to closed under the same connection and re-imports
// (mirrors kibana_connection / APM scoped-connection parity).
func TestAccResourceMLJobStateExplicitConnection(t *testing.T) {
	endpoints := testAccMLJobStateESEndpoints()
	if len(endpoints) == 0 {
		t.Skip("ELASTICSEARCH_ENDPOINTS must be set to run this test")
	}
	endpointVars := make([]config.Variable, 0, len(endpoints))
	for _, endpoint := range endpoints {
		endpointVars = append(endpointVars, config.StringVariable(endpoint))
	}
	jobID := fmt.Sprintf("test-ml-job-state-explicit-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	resourceName := "elasticstack_elasticsearch_ml_job_state.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("opened"),
				ConfigVariables: config.Variables{
					"job_id":    config.StringVariable(jobID),
					"endpoints": config.ListVariable(endpointVars...),
					"api_key":   config.StringVariable(os.Getenv("ELASTICSEARCH_API_KEY")),
					"username":  config.StringVariable(os.Getenv("ELASTICSEARCH_USERNAME")),
					"password":  config.StringVariable(os.Getenv("ELASTICSEARCH_PASSWORD")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "job_id", jobID),
					resource.TestCheckResourceAttr(resourceName, "state", "opened"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.0.endpoints.#", fmt.Sprintf("%d", len(endpoints))),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.0.endpoints.0", endpoints[0]),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.0.insecure", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("opened"),
				ConfigVariables: config.Variables{
					"job_id":    config.StringVariable(jobID),
					"endpoints": config.ListVariable(endpointVars...),
					"api_key":   config.StringVariable(os.Getenv("ELASTICSEARCH_API_KEY")),
					"username":  config.StringVariable(os.Getenv("ELASTICSEARCH_USERNAME")),
					"password":  config.StringVariable(os.Getenv("ELASTICSEARCH_PASSWORD")),
				},
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"elasticsearch_connection"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources[resourceName]
					return rs.Primary.ID, nil
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("closed"),
				ConfigVariables: config.Variables{
					"job_id":    config.StringVariable(jobID),
					"endpoints": config.ListVariable(endpointVars...),
					"api_key":   config.StringVariable(os.Getenv("ELASTICSEARCH_API_KEY")),
					"username":  config.StringVariable(os.Getenv("ELASTICSEARCH_USERNAME")),
					"password":  config.StringVariable(os.Getenv("ELASTICSEARCH_PASSWORD")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "job_id", jobID),
					resource.TestCheckResourceAttr(resourceName, "state", "closed"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "elasticsearch_connection.0.endpoints.0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("closed"),
				ConfigVariables: config.Variables{
					"job_id":    config.StringVariable(jobID),
					"endpoints": config.ListVariable(endpointVars...),
					"api_key":   config.StringVariable(os.Getenv("ELASTICSEARCH_API_KEY")),
					"username":  config.StringVariable(os.Getenv("ELASTICSEARCH_USERNAME")),
					"password":  config.StringVariable(os.Getenv("ELASTICSEARCH_PASSWORD")),
				},
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"elasticsearch_connection"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources[resourceName]
					return rs.Primary.ID, nil
				},
			},
		},
	})
}

func testAccMLJobStateESEndpoints() []string {
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

func TestAccResourceMLJobState_timeouts(t *testing.T) {
	jobID := fmt.Sprintf("test-job-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	indexName := fmt.Sprintf("test-datafeed-index-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("timeouts"),
				ConfigVariables: config.Variables{
					"job_id":     config.StringVariable(jobID),
					"index_name": config.StringVariable(indexName),
				},
				ExpectError: regexp.MustCompile("Operation timed out"),
			},
		},
	})
}
