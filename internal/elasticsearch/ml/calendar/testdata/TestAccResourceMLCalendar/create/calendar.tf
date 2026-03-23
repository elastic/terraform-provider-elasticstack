variable "calendar_id" {
  description = "The calendar ID"
  type        = string
}

variable "job_id" {
  description = "The ML job ID"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "test" {
  job_id      = var.job_id
  description = "Test job for calendar"

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

resource "elasticstack_elasticsearch_ml_calendar" "test" {
  calendar_id = var.calendar_id
  description = "Test calendar"
  job_ids     = [elasticstack_elasticsearch_ml_anomaly_detection_job.test.job_id]
}
