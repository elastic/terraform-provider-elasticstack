variable "job_id" {
  description = "The ML job ID"
  type        = string
}

variable "datafeed_id" {
  description = "The ML datafeed ID"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

# First create the ML job
resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "test" {
  job_id      = var.job_id
  description = "Test job for datafeed"

  analysis_config = {
    bucket_span = "15m"
    detectors = [
      {
        function             = "count"
        detector_description = "Count detector"
      }
    ]
  }

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }
}

# Then create the datafeed with updated configuration
resource "elasticstack_elasticsearch_ml_datafeed" "test" {
  datafeed_id = var.datafeed_id
  job_id      = elasticstack_elasticsearch_ml_anomaly_detection_job.test.job_id
  indices     = ["test-index-*", "test-index-2-*"]  # Added second index

  query = jsonencode({
    match_all = {
      boost = 1
    }
  })

  # Added performance settings to validate update
  scroll_size = 1000
  frequency   = "60s"

  depends_on = [elasticstack_elasticsearch_ml_anomaly_detection_job.test]
}