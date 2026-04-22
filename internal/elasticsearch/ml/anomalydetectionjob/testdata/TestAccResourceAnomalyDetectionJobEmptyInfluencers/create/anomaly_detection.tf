variable "job_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "test" {
  job_id = var.job_id

  analysis_config = {
    bucket_span = "3h"
    detectors = [
      {
        function        = "count"
        over_field_name = "clientip"
      }
    ]
    influencers = []
  }

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }
}
