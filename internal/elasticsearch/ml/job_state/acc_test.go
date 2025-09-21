package job_state_test

import (
	"fmt"
	"regexp"
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
