package job_state_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceMLJobState(t *testing.T) {
	jobID := fmt.Sprintf("test-ml-job-state-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceMLJobStateConfig(jobID, "opened"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "job_id", jobID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "state", "opened"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "force", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "job_timeout", "30s"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_job_state.test", "id"),
					// Verify that the ML job was created by the anomaly detector resource
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "job_id", jobID),
				),
			},
			{
				Config: testAccResourceMLJobStateConfig(jobID, "closed"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "job_id", jobID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "state", "closed"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "force", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "job_timeout", "30s"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_job_state.test", "id"),
				),
			},
			{
				Config: testAccResourceMLJobStateConfigWithOptions(jobID, "opened", true, "1m"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "job_id", jobID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "state", "opened"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "force", "true"),
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceMLJobStateNonExistent,
				ExpectError: regexp.MustCompile(`ML job .* does not exist`),
			},
		},
	})
}

func TestAccResourceMLJobStateImport(t *testing.T) {
	jobID := fmt.Sprintf("test-ml-job-state-import-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceMLJobStateConfig(jobID, "opened"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "job_id", jobID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_job_state.test", "state", "opened"),
				),
			},
			{
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

func testAccResourceMLJobStateConfig(jobID, state string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

# First create an ML anomaly detection job
resource "elasticstack_elasticsearch_ml_anomaly_detector" "test" {
  job_id      = "%s"
  description = "Test anomaly detection job for state management"

  analysis_config = {
    bucket_span = "15m"
    detectors = [
      {
        function              = "count"
        detector_description = "Count detector"
      }
    ]
  }

  analysis_limits = {
    model_memory_limit = "100mb"
  }

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }
}

# Then manage the state of that ML job
resource "elasticstack_elasticsearch_ml_job_state" "test" {
  job_id = elasticstack_elasticsearch_ml_anomaly_detector.test.job_id
  state  = "%s"
  
  depends_on = [elasticstack_elasticsearch_ml_anomaly_detector.test]
}
`, jobID, state)
}

func testAccResourceMLJobStateConfigWithOptions(jobID, state string, force bool, timeout string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

# First create an ML anomaly detection job
resource "elasticstack_elasticsearch_ml_anomaly_detector" "test" {
  job_id      = "%s"
  description = "Test anomaly detection job for state management with options"

  analysis_config = {
    bucket_span = "15m"
    detectors = [
      {
        function              = "count"
        detector_description = "Count detector"
      }
    ]
  }

  analysis_limits = {
    model_memory_limit = "100mb"
  }

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }
}

# Then manage the state of that ML job with custom options
resource "elasticstack_elasticsearch_ml_job_state" "test" {
  job_id         = elasticstack_elasticsearch_ml_anomaly_detector.test.job_id
  state          = "%s"
  force          = %t
  job_timeout    = "%s"
  
  depends_on = [elasticstack_elasticsearch_ml_anomaly_detector.test]
}
`, jobID, state, force, timeout)
}

const testAccResourceMLJobStateNonExistent = `
provider "elasticstack" {
  elasticsearch {}
}

# Try to manage state of a non-existent ML job
resource "elasticstack_elasticsearch_ml_job_state" "test" {
  job_id = "non-existent-ml-job"
  state  = "opened"
}
`
