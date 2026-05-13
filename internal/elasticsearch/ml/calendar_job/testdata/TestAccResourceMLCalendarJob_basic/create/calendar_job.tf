variable "calendar_id" {
  type = string
}

variable "job_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "job" {
  job_id      = var.job_id
  description = "ACC job for ml_calendar_job"

  analysis_config = {
    bucket_span = "15m"
    detectors = [
      {
        function = "count"
      }
    ]
  }

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }
}

resource "elasticstack_elasticsearch_ml_calendar_job" "test" {
  calendar_id = var.calendar_id
  job_id = elasticstack_elasticsearch_ml_anomaly_detection_job.job.job_id
}
