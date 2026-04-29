variable "job_id" {
  description = "The job ID for the anomaly detection job"
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

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "test" {
  job_id      = var.job_id
  description = "Updated anomaly detection job with explicit connection"

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

  elasticsearch_connection {
    endpoints = var.endpoints
    username  = var.username
    password  = var.password
    insecure  = true
  }
}
