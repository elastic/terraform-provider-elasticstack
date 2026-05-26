variable "calendar_id" {
  type = string
}

variable "job_id_a" {
  type = string
}

variable "job_id_b" {
  type = string
}

variable "attach_job" {
  type        = string
  description = "Which anomaly job to attach: \"a\" or \"b\""
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "a" {
  job_id      = var.job_id_a
  description = "ACC job A for ml_calendar_job replace job"

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

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "b" {
  job_id      = var.job_id_b
  description = "ACC job B for ml_calendar_job replace job"

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
  job_id      = var.attach_job == "a" ? elasticstack_elasticsearch_ml_anomaly_detection_job.a.job_id : elasticstack_elasticsearch_ml_anomaly_detection_job.b.job_id
}
