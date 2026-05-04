variable "job_id" {
  description = "The job ID for the ML job"
  type        = string
}

variable "endpoints" {
  type = list(string)
}

variable "api_key" {
  type    = string
  default = ""
}

provider "elasticstack" {
  elasticsearch {}
}

# Anomaly job uses the default provider connection; job_state uses an explicit
# api_key elasticsearch_connection to exercise that auth branch.
resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "test" {
  job_id      = var.job_id
  description = "Test anomaly detection job for explicit connection api_key coverage"

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

resource "elasticstack_elasticsearch_ml_job_state" "test" {
  job_id = elasticstack_elasticsearch_ml_anomaly_detection_job.test.job_id
  state  = "opened"

  elasticsearch_connection {
    endpoints = var.endpoints
    api_key   = var.api_key
    insecure  = true
  }
}
