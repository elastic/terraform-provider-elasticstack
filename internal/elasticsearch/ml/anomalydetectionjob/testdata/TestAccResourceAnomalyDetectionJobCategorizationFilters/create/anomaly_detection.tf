variable "job_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "test" {
  job_id = var.job_id

  analysis_config = {
    bucket_span               = "1h"
    categorization_field_name = "message"
    categorization_filters = [
      "\\b\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\b",
      "\\b[A-Fa-f0-9]{8,}\\b",
    ]
    detectors = [
      {
        function      = "count"
        by_field_name = "mlcategory"
      }
    ]
  }

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }
}
