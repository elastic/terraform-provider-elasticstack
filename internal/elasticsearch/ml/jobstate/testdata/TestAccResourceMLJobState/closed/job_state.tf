variable "job_id" {
  description = "The job ID for the ML job"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

# First create an ML anomaly detection job
resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "test" {
  job_id      = var.job_id
  description = "Test anomaly detection job for state management"

  analysis_config = {
    bucket_span = "15m"
    detectors = [
      {
        function             = "count"
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
  job_id = elasticstack_elasticsearch_ml_anomaly_detection_job.test.job_id
  state  = "closed"
}