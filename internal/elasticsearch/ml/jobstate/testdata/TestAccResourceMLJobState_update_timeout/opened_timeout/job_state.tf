variable "job_id" {
  description = "The ML job ID"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

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

# Attempt to update to "opened" with a very short timeout. The job requires 2 GB
# of model memory and allow_lazy_open is set, so the open call returns immediately
# but the job stays in "opening" until a node with enough memory is available.
# With a 10 s update timeout the waitForJobState loop times out first.
resource "elasticstack_elasticsearch_ml_job_state" "test" {
  job_id = elasticstack_elasticsearch_ml_anomaly_detection_job.test.job_id
  state  = "opened"

  timeouts = {
    update = "10s"
  }
}
