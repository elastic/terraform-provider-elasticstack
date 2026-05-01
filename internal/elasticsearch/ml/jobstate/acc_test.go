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

const mlJobStateResourceName = "elasticstack_elasticsearch_ml_job_state.test"

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
					resource.TestCheckResourceAttr(mlJobStateResourceName, "job_id", jobID),
					resource.TestCheckResourceAttr(mlJobStateResourceName, "state", "opened"),
					resource.TestCheckResourceAttr(mlJobStateResourceName, "force", "false"),
					resource.TestCheckResourceAttr(mlJobStateResourceName, "job_timeout", "30s"),
					resource.TestCheckResourceAttrSet(mlJobStateResourceName, "id"),
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
					resource.TestCheckResourceAttr(mlJobStateResourceName, "job_id", jobID),
					resource.TestCheckResourceAttr(mlJobStateResourceName, "state", "closed"),
					resource.TestCheckResourceAttr(mlJobStateResourceName, "force", "false"),
					resource.TestCheckResourceAttr(mlJobStateResourceName, "job_timeout", "30s"),
					resource.TestCheckResourceAttrSet(mlJobStateResourceName, "id"),
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
					resource.TestCheckResourceAttr(mlJobStateResourceName, "job_id", jobID),
					resource.TestCheckResourceAttr(mlJobStateResourceName, "state", "opened"),
					resource.TestCheckResourceAttr(mlJobStateResourceName, "force", "true"),
					resource.TestCheckResourceAttr(mlJobStateResourceName, "job_timeout", "1m"),
					resource.TestCheckResourceAttrSet(mlJobStateResourceName, "id"),
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
				ResourceName:            mlJobStateResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force", "job_timeout", "timeouts"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources[mlJobStateResourceName]
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
					resource.TestCheckResourceAttr(mlJobStateResourceName, "job_id", jobID),
					resource.TestCheckResourceAttr(mlJobStateResourceName, "state", "opened"),
					resource.TestCheckResourceAttr(mlJobStateResourceName, "force", "false"),
					resource.TestCheckResourceAttr(mlJobStateResourceName, "job_timeout", "30s"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("opened"),
				ConfigVariables: config.Variables{
					"job_id": config.StringVariable(jobID),
				},
				ResourceName:      mlJobStateResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources[mlJobStateResourceName]
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
	resourceName := mlJobStateResourceName

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

// TestAccResourceMLJobStateForceAndTimeoutUpdate verifies that force and job_timeout can be
// updated independently of job state, and that defaults are correctly restored.
// Step 1 creates with defaults, step 2 updates both attributes while keeping state=opened,
// step 3 (reversal) restores defaults and re-verifies.
func TestAccResourceMLJobStateForceAndTimeoutUpdate(t *testing.T) {
	jobID := fmt.Sprintf("test-ml-job-upd-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	resourceName := mlJobStateResourceName

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Step 1: create with default force and job_timeout.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("opened"),
				ConfigVariables: config.Variables{
					"job_id": config.StringVariable(jobID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "job_id", jobID),
					resource.TestCheckResourceAttr(resourceName, "state", "opened"),
					resource.TestCheckResourceAttr(resourceName, "force", "false"),
					resource.TestCheckResourceAttr(resourceName, "job_timeout", "30s"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			// Step 2: update force=true and job_timeout="1m" while keeping state=opened.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("opened_with_options"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"force":       config.BoolVariable(true),
					"job_timeout": config.StringVariable("1m"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", "opened"),
					resource.TestCheckResourceAttr(resourceName, "force", "true"),
					resource.TestCheckResourceAttr(resourceName, "job_timeout", "1m"),
				),
			},
			// Step 3: reversal - omit force and job_timeout to restore defaults.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("opened"),
				ConfigVariables: config.Variables{
					"job_id": config.StringVariable(jobID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", "opened"),
					resource.TestCheckResourceAttr(resourceName, "force", "false"),
					resource.TestCheckResourceAttr(resourceName, "job_timeout", "30s"),
				),
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
	resourceName := mlJobStateResourceName

	// Build per-endpoint assertions for all indices.
	openedChecks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(resourceName, "job_id", jobID),
		resource.TestCheckResourceAttr(resourceName, "state", "opened"),
		resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.#", "1"),
		resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.0.endpoints.#", fmt.Sprintf("%d", len(endpoints))),
		resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.0.insecure", "true"),
	}
	for i, ep := range endpoints {
		openedChecks = append(openedChecks,
			resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("elasticsearch_connection.0.endpoints.%d", i), ep),
		)
	}
	// Assert the active auth field based on which credentials are available.
	if apiKey := os.Getenv("ELASTICSEARCH_API_KEY"); apiKey != "" {
		openedChecks = append(openedChecks,
			resource.TestCheckResourceAttrSet(resourceName, "elasticsearch_connection.0.api_key"),
		)
	} else if username := os.Getenv("ELASTICSEARCH_USERNAME"); username != "" {
		openedChecks = append(openedChecks,
			resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.0.username", username),
		)
	}

	closedChecks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(resourceName, "job_id", jobID),
		resource.TestCheckResourceAttr(resourceName, "state", "closed"),
		resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.#", "1"),
		resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.0.endpoints.#", fmt.Sprintf("%d", len(endpoints))),
		resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.0.insecure", "true"),
	}
	for i, ep := range endpoints {
		closedChecks = append(closedChecks,
			resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("elasticsearch_connection.0.endpoints.%d", i), ep),
		)
	}

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
				Check: resource.ComposeTestCheckFunc(openedChecks...),
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
				Check: resource.ComposeTestCheckFunc(closedChecks...),
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
			// Step 5: re-open using a dedicated username/password-only config so the
			// username field can be explicitly asserted in state.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("opened_userpass"),
				ConfigVariables: config.Variables{
					"job_id":    config.StringVariable(jobID),
					"endpoints": config.ListVariable(endpointVars...),
					"username":  config.StringVariable(os.Getenv("ELASTICSEARCH_USERNAME")),
					"password":  config.StringVariable(os.Getenv("ELASTICSEARCH_PASSWORD")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", "opened"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.0.endpoints.#", fmt.Sprintf("%d", len(endpoints))),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.0.username", os.Getenv("ELASTICSEARCH_USERNAME")),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.0.insecure", "true"),
				),
			},
		},
	})
}

// TestAccResourceMLJobStateExplicitConnectionAPIKey verifies the api_key auth path for
// the elasticsearch_connection block, asserting the key is stored in state and all
// configured endpoints are present. Skipped when ELASTICSEARCH_API_KEY is not set.
func TestAccResourceMLJobStateExplicitConnectionAPIKey(t *testing.T) {
	apiKey := os.Getenv("ELASTICSEARCH_API_KEY")
	if apiKey == "" {
		t.Skip("ELASTICSEARCH_API_KEY must be set to run this test")
	}
	endpoints := testAccMLJobStateESEndpoints()
	if len(endpoints) == 0 {
		t.Skip("ELASTICSEARCH_ENDPOINTS must be set to run this test")
	}
	endpointVars := make([]config.Variable, 0, len(endpoints))
	for _, endpoint := range endpoints {
		endpointVars = append(endpointVars, config.StringVariable(endpoint))
	}
	jobID := fmt.Sprintf("test-ml-job-apikey-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	resourceName := mlJobStateResourceName

	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(resourceName, "state", "opened"),
		resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.#", "1"),
		resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.0.endpoints.#", fmt.Sprintf("%d", len(endpoints))),
		resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.0.insecure", "true"),
		// api_key is sensitive but stored in state; verify it is present and correct.
		resource.TestCheckResourceAttr(resourceName, "elasticsearch_connection.0.api_key", apiKey),
	}
	for i, ep := range endpoints {
		checks = append(checks,
			resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("elasticsearch_connection.0.endpoints.%d", i), ep),
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("opened_apikey"),
				ConfigVariables: config.Variables{
					"job_id":    config.StringVariable(jobID),
					"endpoints": config.ListVariable(endpointVars...),
					"api_key":   config.StringVariable(apiKey),
				},
				Check: resource.ComposeTestCheckFunc(checks...),
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

// TestAccResourceMLJobState_update_timeout verifies that a short timeouts.update causes
// the update path to fail with "Operation timed out". Step 1 creates the resource in
// closed state (succeeds quickly since the job starts closed). Step 2 attempts to update
// to opened with a 10 s timeout; because allow_lazy_open is set and the job requires 2 GB
// of model memory, waitForJobState times out before the job reaches "opened".
func TestAccResourceMLJobState_update_timeout(t *testing.T) {
	jobID := fmt.Sprintf("test-job-upd-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Step 1: create the job resource in closed state - should succeed without timeout.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("closed"),
				ConfigVariables: config.Variables{
					"job_id": config.StringVariable(jobID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlJobStateResourceName, "state", "closed"),
				),
			},
			// Step 2: attempt to update to opened with a very short timeout - expect failure.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("opened_timeout"),
				ConfigVariables: config.Variables{
					"job_id": config.StringVariable(jobID),
				},
				ExpectError: regexp.MustCompile("Operation timed out"),
			},
		},
	})
}
