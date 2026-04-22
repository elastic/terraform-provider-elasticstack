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
        function      = "count"
        by_field_name = "mlcategory"
      }
    ]
    per_partition_categorization = {
      enabled      = false
      stop_on_warn = false
    }
  }

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }
}
