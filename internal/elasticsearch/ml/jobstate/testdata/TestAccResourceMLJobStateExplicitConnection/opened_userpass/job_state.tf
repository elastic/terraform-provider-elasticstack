variable "job_id" {
  description = "The job ID for the ML job"
  type        = string
}

variable "endpoints" {
  type = list(string)
}

variable "username" {
  type    = string
  default = ""
}

variable "password" {
  type    = string
  default = ""
}

provider "elasticstack" {
  elasticsearch {}
}

# Anomaly job uses the default provider connection; job_state uses an explicit
# username/password elasticsearch_connection to exercise that auth branch.
resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "test" {
  job_id      = var.job_id
  description = "Test anomaly detection job for explicit connection username/password coverage"

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
    username  = var.username
    password  = var.password
    insecure  = true
  }
}
