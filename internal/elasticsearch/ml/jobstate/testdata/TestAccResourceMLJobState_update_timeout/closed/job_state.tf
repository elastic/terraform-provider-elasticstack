variable "job_id" {
  description = "The ML job ID"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

# Create a high-memory job with allow_lazy_open so that opening it later will
# require waiting for resources - enabling the update-timeout scenario.
resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "test" {
  job_id      = var.job_id
  description = "Test job for update timeout testing"
  analysis_config = {
    bucket_span = "1h"
    detectors = [{
      function             = "count"
      detector_description = "count"
    }]
  }
  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }
  analysis_limits = {
    model_memory_limit = "2gb"
  }
  allow_lazy_open = true
}

# Start in closed state - no timeout concerns since the job is already closed
# when first created by Elasticsearch.
resource "elasticstack_elasticsearch_ml_job_state" "test" {
  job_id = elasticstack_elasticsearch_ml_anomaly_detection_job.test.job_id
  state  = "closed"
}
