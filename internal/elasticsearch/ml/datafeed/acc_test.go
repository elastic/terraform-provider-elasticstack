package datafeed_test

import (
	_ "embed"
	"fmt"
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

					// Query and data retrieval
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_datafeed.test", "query"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_datafeed.test", "script_fields"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_datafeed.test", "runtime_mappings"),

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

					// Verify JSON fields are updated
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_datafeed.test", "query"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_datafeed.test", "script_fields"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_datafeed.test", "runtime_mappings"),
				),
			},
		},
	})
}
