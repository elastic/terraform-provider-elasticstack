variable "job_id" {
  type = string
}

variable "cal_a" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "test" {
  job_id      = var.job_id
  description = "ACC ML calendars: explicit empty calendars on create"

  calendars = toset([])

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
    model_memory_limit = "10mb"
  }

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }

  allow_lazy_open = true
}
