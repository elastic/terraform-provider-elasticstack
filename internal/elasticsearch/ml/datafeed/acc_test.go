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

package datafeed_test

import (
	_ "embed"
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

func TestAccResourceDatafeed(t *testing.T) {
	jobID := fmt.Sprintf("test-job-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	datafeedID := fmt.Sprintf("test-datafeed-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "datafeed_id", datafeedID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "job_id", jobID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "indices.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "indices.0", "test-index-*"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_datafeed.test", "query"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "datafeed_id", datafeedID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "job_id", jobID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "indices.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "indices.0", "test-index-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "indices.1", "test-index-2-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "scroll_size", "1000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "frequency", "60s"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_datafeed.test", "query"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ResourceName:             "elasticstack_elasticsearch_ml_datafeed.test",
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["elasticstack_elasticsearch_ml_datafeed.test"]
					return rs.Primary.ID, nil
				},
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
				},
			},
		},
	})
}

func TestAccResourceDatafeedComprehensive(t *testing.T) {
	jobID := fmt.Sprintf("test-job-comprehensive-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	datafeedID := fmt.Sprintf("test-datafeed-comprehensive-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	// Expected normalized JSON for query (create step): keys sorted by json.Marshal
	const queryCreate = `{"bool":{"must":[{"range":{"@timestamp":{"gte":"now-1h"}}},{"term":{"status":"active"}}]}}`
	// Expected normalized JSON for query (update step)
	const queryUpdate = `{"bool":{"must":[{"range":{"@timestamp":{"gte":"now-2h"}}},{"term":{"status":"updated"}}]}}`

	// Expected normalized JSON for runtime_mappings: "script" < "type" alphabetically
	const runtimeMappingsCreate = `{"hour_of_day":{"script":{"source":"emit(doc['@timestamp'].value.hour)"},"type":"long"}}`
	const runtimeMappingsUpdate = `{"day_of_week":{"script":{"source":"emit(doc['@timestamp'].value.dayOfWeek)"},"type":"long"}}`

	// Expected normalized JSON for script_fields after Elasticsearch stores them.
	// Elasticsearch does NOT add lang:"painless" to the stored representation.
	// Field names sorted: "double_value" < "status_upper".
	const scriptFieldsCreate = `{"double_value":{"script":{"source":"doc['value'].value * 2"}},` +
		`"status_upper":{"script":{"source":"doc['status'].value.toUpperCase()"}}}`
	// Update step changes double_value→triple_value and status_upper→status_lower.
	// "status_lower" < "triple_value" alphabetically.
	const scriptFieldsUpdate = `{"status_lower":{"script":{"source":"doc['status'].value.toLowerCase()"}},` +
		`"triple_value":{"script":{"source":"doc['value'].value * 3"}}}`

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
				},
				Check: resource.ComposeTestCheckFunc(
					// Basic attributes
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "datafeed_id", datafeedID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "job_id", jobID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "indices.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "indices.0", "test-index-1-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "indices.1", "test-index-2-*"),

					// Exact JSON assertions for query, script_fields, runtime_mappings
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "query", queryCreate),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "script_fields", scriptFieldsCreate),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "runtime_mappings", runtimeMappingsCreate),

					// Performance settings
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "scroll_size", "500"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "frequency", "30s"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "query_delay", "60s"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "max_empty_searches", "10"),

					// Chunking config
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "chunking_config.mode", "manual"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "chunking_config.time_span", "1h"),

					// Delayed data check config
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "delayed_data_check_config.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "delayed_data_check_config.check_window", "2h"),

					// Indices options
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "indices_options.expand_wildcards.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "indices_options.expand_wildcards.0", "open"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "indices_options.expand_wildcards.1", "closed"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "indices_options.ignore_unavailable", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "indices_options.allow_no_indices", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "indices_options.ignore_throttled", "false"),

					// Computed fields
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_datafeed.test", "id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
				},
				Check: resource.ComposeTestCheckFunc(
					// Verify updates - basic attributes
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "datafeed_id", datafeedID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "job_id", jobID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "indices.#", "3"),              // Updated to 3 indices
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "indices.2", "test-index-3-*"), // New index added

					// Verify exact JSON updates
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "query", queryUpdate),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "script_fields", scriptFieldsUpdate),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "runtime_mappings", runtimeMappingsUpdate),

					// Verify updated performance settings
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "scroll_size", "1000"),      // Updated from 500
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "frequency", "60s"),         // Updated from 30s
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "query_delay", "120s"),      // Updated from 60s
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "max_empty_searches", "20"), // Updated from 10

					// Verify updated chunking config
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "chunking_config.mode", "manual"),  // Keep manual mode
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "chunking_config.time_span", "2h"), // Updated from 1h to 2h

					// Verify updated delayed data check config
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "delayed_data_check_config.enabled", "false"),   // Updated from true
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "delayed_data_check_config.check_window", "4h"), // Updated from 2h

					// Verify updated indices options
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "indices_options.expand_wildcards.#", "1"), // Updated to 1
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "indices_options.expand_wildcards.0", "open"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "indices_options.ignore_unavailable", "false"), // Updated from true
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "indices_options.allow_no_indices", "true"),    // Updated from false
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "indices_options.ignore_throttled", "true"),    // Updated from false
				),
			},
		},
	})
}

// TestAccResourceDatafeedAggregations verifies that a datafeed configured with
// aggregations (mutually exclusive with script_fields) is stored and read back
// with the exact normalized JSON value.
func TestAccResourceDatafeedAggregations(t *testing.T) {
	jobID := fmt.Sprintf("test-job-aggs-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	datafeedID := fmt.Sprintf("test-datafeed-aggs-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	// Expected normalized JSON for aggregations after ES stores and Go's json.Marshal re-serialises them.
	// HCL jsonencode sorts keys alphabetically: "aggregations" < "date_histogram".
	// Within date_histogram: "field" < "fixed_interval" < "time_zone".
	const expectedAggregations = `{"buckets":{"aggregations":{"@timestamp":{"max":{"field":"@timestamp"}}},"date_histogram":{"field":"@timestamp","fixed_interval":"15m","time_zone":"UTC"}}}`

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "datafeed_id", datafeedID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "job_id", jobID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "indices.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "aggregations", expectedAggregations),
					// script_fields must be absent (mutually exclusive with aggregations)
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_ml_datafeed.test", "script_fields"),
				),
			},
		},
	})
}

// TestAccResourceDatafeedExplicitConnection exercises the elasticsearch_connection block
// on the ML datafeed resource, covering both the create and update paths, and verifies
// that import ignores the sensitive connection block.
func TestAccResourceDatafeedExplicitConnection(t *testing.T) {
	endpoints := testAccDatafeedESEndpoints()
	if len(endpoints) == 0 {
		t.Skip("ELASTICSEARCH_ENDPOINTS must be set to run this test")
	}
	endpointVars := make([]config.Variable, 0, len(endpoints))
	for _, endpoint := range endpoints {
		endpointVars = append(endpointVars, config.StringVariable(endpoint))
	}

	jobID := fmt.Sprintf("test-job-conn-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	datafeedID := fmt.Sprintf("test-datafeed-conn-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	const resourceAddr = "elasticstack_elasticsearch_ml_datafeed.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Step 1: create with explicit connection (api_key if available, else username/password)
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
					"endpoints":   config.ListVariable(endpointVars...),
					"api_key":     config.StringVariable(os.Getenv("ELASTICSEARCH_API_KEY")),
					"username":    config.StringVariable(os.Getenv("ELASTICSEARCH_USERNAME")),
					"password":    config.StringVariable(os.Getenv("ELASTICSEARCH_PASSWORD")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "datafeed_id", datafeedID),
					resource.TestCheckResourceAttr(resourceAddr, "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr(resourceAddr, "elasticsearch_connection.0.endpoints.#", fmt.Sprintf("%d", len(endpoints))),
					resource.TestCheckResourceAttr(resourceAddr, "elasticsearch_connection.0.endpoints.0", endpoints[0]),
					resource.TestCheckResourceAttr(resourceAddr, "elasticsearch_connection.0.insecure", "true"),
				),
			},
			// Step 2: import; sensitive connection block is intentionally ignored
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
					"endpoints":   config.ListVariable(endpointVars...),
					"api_key":     config.StringVariable(os.Getenv("ELASTICSEARCH_API_KEY")),
					"username":    config.StringVariable(os.Getenv("ELASTICSEARCH_USERNAME")),
					"password":    config.StringVariable(os.Getenv("ELASTICSEARCH_PASSWORD")),
				},
				ResourceName:            resourceAddr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"elasticsearch_connection"},
			},
			// Step 3: update indices while keeping the same explicit connection (username/password)
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
					"endpoints":   config.ListVariable(endpointVars...),
					"username":    config.StringVariable(os.Getenv("ELASTICSEARCH_USERNAME")),
					"password":    config.StringVariable(os.Getenv("ELASTICSEARCH_PASSWORD")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "datafeed_id", datafeedID),
					resource.TestCheckResourceAttr(resourceAddr, "indices.#", "2"),
					resource.TestCheckResourceAttr(resourceAddr, "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr(resourceAddr, "elasticsearch_connection.0.endpoints.#", fmt.Sprintf("%d", len(endpoints))),
					resource.TestCheckResourceAttr(resourceAddr, "elasticsearch_connection.0.endpoints.0", endpoints[0]),
					resource.TestCheckResourceAttr(resourceAddr, "elasticsearch_connection.0.insecure", "true"),
				),
			},
			// Step 4: re-import after update to confirm connection block is ignored
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
					"endpoints":   config.ListVariable(endpointVars...),
					"username":    config.StringVariable(os.Getenv("ELASTICSEARCH_USERNAME")),
					"password":    config.StringVariable(os.Getenv("ELASTICSEARCH_PASSWORD")),
				},
				ResourceName:            resourceAddr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"elasticsearch_connection"},
			},
		},
	})
}

// TestAccResourceDatafeedChunkingModes exercises chunking_config.mode = "auto" and "off",
// verifying that time_span is absent (null) in both cases since it is only valid for mode = "manual".
func TestAccResourceDatafeedChunkingModes(t *testing.T) {
	jobID := fmt.Sprintf("test-job-chunk-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	datafeedID := fmt.Sprintf("test-datafeed-chunk-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	const resourceAddr = "elasticstack_elasticsearch_ml_datafeed.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Step 1: mode = "auto" — time_span must not be set
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("auto"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "datafeed_id", datafeedID),
					resource.TestCheckResourceAttr(resourceAddr, "chunking_config.mode", "auto"),
					resource.TestCheckNoResourceAttr(resourceAddr, "chunking_config.time_span"),
				),
			},
			// Step 2: mode = "off" — time_span must not be set
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("off"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "datafeed_id", datafeedID),
					resource.TestCheckResourceAttr(resourceAddr, "chunking_config.mode", "off"),
					resource.TestCheckNoResourceAttr(resourceAddr, "chunking_config.time_span"),
				),
			},
		},
	})
}

// TestAccResourceDatafeedDelayedDataDisabled verifies the omission/unset path for
// delayed_data_check_config.check_window when the check is disabled, and also broadens
// coverage of indices_options.expand_wildcards with the "all" value.
func TestAccResourceDatafeedDelayedDataDisabled(t *testing.T) {
	jobID := fmt.Sprintf("test-job-delay-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	datafeedID := fmt.Sprintf("test-datafeed-delay-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	const resourceAddr = "elasticstack_elasticsearch_ml_datafeed.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable(jobID),
					"datafeed_id": config.StringVariable(datafeedID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "datafeed_id", datafeedID),

					// Disabled delayed data check without check_window: check_window must be absent/empty
					resource.TestCheckResourceAttr(resourceAddr, "delayed_data_check_config.enabled", "false"),
					resource.TestCheckNoResourceAttr(resourceAddr, "delayed_data_check_config.check_window"),

					// Broader expand_wildcards coverage: "hidden" value not covered by the comprehensive test
					resource.TestCheckResourceAttr(resourceAddr, "indices_options.expand_wildcards.#", "2"),
					resource.TestCheckResourceAttr(resourceAddr, "indices_options.expand_wildcards.0", "open"),
					resource.TestCheckResourceAttr(resourceAddr, "indices_options.expand_wildcards.1", "hidden"),
				),
			},
		},
	})
}

func TestAccResourceDatafeed_ImportNonExistent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("import_missing"),
				ConfigVariables: config.Variables{
					"job_id":      config.StringVariable("dummy-job"),
					"datafeed_id": config.StringVariable("dummy-datafeed"),
				},
				ResourceName:      "elasticstack_elasticsearch_ml_datafeed.test",
				ImportState:       true,
				ImportStateId:     "cluster-id/non-existent-datafeed-id",
				ImportStateVerify: false,
			},
		},
	})
}

// testAccDatafeedESEndpoints parses ELASTICSEARCH_ENDPOINTS into a slice of endpoint strings.
func testAccDatafeedESEndpoints() []string {
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
