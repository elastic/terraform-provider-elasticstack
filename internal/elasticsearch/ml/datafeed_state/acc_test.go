package datafeed_state_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
