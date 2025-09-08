package anomaly_detector_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceAnomalyDetectorJob(t *testing.T) {
	jobID := fmt.Sprintf("test-anomaly-detector-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAnomalyDetectorJobBasic(jobID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "job_id", jobID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "description", "Test anomaly detection job"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "analysis_config.bucket_span", "15m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "analysis_config.detectors.0.function", "count"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "data_description.time_field", "@timestamp"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "data_description.time_format", "epoch_ms"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_anomaly_detector.test", "create_time"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "job_type", "anomaly_detector"),
				),
			},
			{
				Config: testAccResourceAnomalyDetectorJobUpdated(jobID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "job_id", jobID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "description", "Updated test anomaly detection job"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "groups.#", "1"),
				),
			},
		},
	})
}

func testAccResourceAnomalyDetectorJobBasic(jobID string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_anomaly_detector" "test" {
  job_id      = "%s"
  description = "Test anomaly detection job"

  analysis_config = {
    bucket_span = "15m"
    detectors = [
      {
        function              = "count"
        detector_description = "Count detector"
      }
    ]
  }

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }
}
`, jobID)
}

func testAccResourceAnomalyDetectorJobUpdated(jobID string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_anomaly_detector" "test" {
  job_id      = "%s"
  description = "Updated test anomaly detection job"

  analysis_config = {
    bucket_span = "15m"
    detectors = [
      {
        function              = "count"
        detector_description = "Count detector"
      }
    ]
  }

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }

  groups = ["test-group"]
}
`, jobID)
}
