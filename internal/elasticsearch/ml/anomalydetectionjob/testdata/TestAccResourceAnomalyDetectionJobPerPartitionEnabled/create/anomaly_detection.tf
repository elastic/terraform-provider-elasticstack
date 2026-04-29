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
    detectors = [
      {
        function             = "count"
        by_field_name        = "mlcategory"
        partition_field_name = "service"
      }
    ]
    per_partition_categorization = {
      enabled      = true
      stop_on_warn = true
    }
  }

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }
}
