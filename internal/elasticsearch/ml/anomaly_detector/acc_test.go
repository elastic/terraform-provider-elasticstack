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
				Config: testAccResourceAnomalyDetectorJobComprehensive(jobID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "job_id", jobID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "description", "Comprehensive test anomaly detection job"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "groups.#", "2"),
					// Analysis config checks
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "analysis_config.bucket_span", "10m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "analysis_config.detectors.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "analysis_config.detectors.0.function", "count"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "analysis_config.detectors.1.function", "mean"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "analysis_config.detectors.1.field_name", "response_time"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "analysis_config.influencers.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "analysis_config.influencers.0", "status_code"),
					// Analysis limits checks
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "analysis_limits.model_memory_limit", "100mb"),
					// Data description checks
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "data_description.time_field", "@timestamp"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "data_description.time_format", "epoch_ms"),
					// Datafeed config checks
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "datafeed_config.datafeed_id", jobID+"-datafeed"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "datafeed_config.indices.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "datafeed_config.indices.0", "test-index-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "datafeed_config.scroll_size", "1000"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "datafeed_config.frequency", "30s"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "datafeed_config.chunking_config.mode", "auto"),
					// Model plot config checks
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "model_plot_config.enabled", "true"),
					// Other settings checks
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "allow_lazy_open", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "background_persist_interval", "1h"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "custom_settings", "{\"custom_key\": \"custom_value\"}"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "daily_model_snapshot_retention_after_days", "3"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "model_snapshot_retention_days", "7"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "renormalization_window_days", "14"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "results_retention_days", "30"),
					// Computed fields
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_anomaly_detector.test", "create_time"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "job_type", "anomaly_detector"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_anomaly_detector.test", "job_version"),
				),
			},
			{
				Config: testAccResourceAnomalyDetectorJobUpdated(jobID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "job_id", jobID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "description", "Updated test anomaly detection job"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "groups.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "groups.0", "test-group"),
					// Verify that updatable fields were actually updated
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "analysis_limits.model_memory_limit", "200mb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "model_plot_config.enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "allow_lazy_open", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "background_persist_interval", "2h"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "custom_settings", "{\"updated_key\": \"updated_value\"}"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "daily_model_snapshot_retention_after_days", "5"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "model_snapshot_retention_days", "14"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "renormalization_window_days", "30"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detector.test", "results_retention_days", "60"),
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

func testAccResourceAnomalyDetectorJobComprehensive(jobID string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_anomaly_detector" "test" {
  job_id      = "%s"
  description = "Comprehensive test anomaly detection job"
  groups      = ["test-group", "ml-group"]

  analysis_config = {
    bucket_span = "10m"
    detectors = [
      {
        function              = "count"
        detector_description = "Count detector"
      },
      {
        function              = "mean"
        field_name           = "response_time"
        detector_description = "Mean response time detector"
      }
    ]
    influencers = ["status_code"]
  }

  analysis_limits = {
    model_memory_limit = "100mb"
  }

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }

  datafeed_config = {
    datafeed_id = "%s-datafeed"
    indices     = ["test-index-*"]
    scroll_size = 1000
    frequency   = "30s"
    
    chunking_config = {
      mode = "auto"
    }
  }

  model_plot_config = {
    enabled = true
  }

  allow_lazy_open = true
  background_persist_interval = "1h"
  custom_settings = "{\"custom_key\": \"custom_value\"}"
  daily_model_snapshot_retention_after_days = 3
  model_snapshot_retention_days = 7
  renormalization_window_days = 14
  results_retention_days = 30
}
`, jobID, jobID)
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
  
  # Test updating some of the other updatable fields
  analysis_limits = {
    model_memory_limit = "200mb"
  }
  
  model_plot_config = {
    enabled = false
  }
  
  allow_lazy_open = false
  background_persist_interval = "2h"
  custom_settings = "{\"updated_key\": \"updated_value\"}"
  daily_model_snapshot_retention_after_days = 5
  model_snapshot_retention_days = 14
  renormalization_window_days = 30
  results_retention_days = 60
}
`, jobID)
}
