variable "job_id" {
  description = "The ML job ID"
  type        = string
}

variable "datafeed_id" {
  description = "The ML datafeed ID"
  type        = string
}

variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name                = var.index_name
  deletion_protection = false
  mappings = jsonencode({
    properties = {
      "@timestamp" = {
        type = "date"
      }
      nginx = {
        properties = {
          access = {
            properties = {
              body_sent = {
                properties = {
                  bytes = {
                    type = "long"
                  }
                }
              }
              geoip = {
                properties = {
                  city_name = {
                    type = "keyword"
                  }
                }
              }
              user_agent = {
                properties = {
                  build = {
                    type = "keyword"
                  }
                }
              }
            }
          }
        }
      }
    }
  })
}

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "nginx" {
  job_id      = var.job_id
  description = "Anomaly detection for network traffic"
  analysis_config = {
    bucket_span = "15m"
    detectors = [
      {
        function             = "count"
        detector_description = "count"
      },
      {
        function             = "mean"
        field_name           = "nginx.access.body_sent.bytes"
        detector_description = "mean(\"nginx.access.body_sent.bytes\")"
      }
    ]
    influencers        = ["nginx.access.geoip.city_name", "nginx.access.user_agent.build"]
    model_prune_window = "30d"
  }
  analysis_limits = {
    model_memory_limit            = "10MB"
    categorization_examples_limit = 4
  }
  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }
  model_snapshot_retention_days             = 10
  daily_model_snapshot_retention_after_days = 1
}

resource "elasticstack_elasticsearch_ml_datafeed" "datafeed_nginx" {
  datafeed_id = var.datafeed_id
  job_id      = elasticstack_elasticsearch_ml_anomaly_detection_job.nginx.job_id
  query = jsonencode({
    bool = {
      must = [
        {
          match_all = {}
        }
      ]
    }
  })
  indices = [elasticstack_elasticsearch_index.test.name]
}

resource "elasticstack_elasticsearch_ml_job_state" "nginx" {
  job_id = elasticstack_elasticsearch_ml_anomaly_detection_job.nginx.job_id
  state  = "closed"
}

resource "elasticstack_elasticsearch_ml_datafeed_state" "nginx" {
  datafeed_id = elasticstack_elasticsearch_ml_datafeed.datafeed_nginx.datafeed_id
  state       = "stopped"
  force       = true

  depends_on = [
    elasticstack_elasticsearch_ml_datafeed.datafeed_nginx,
    elasticstack_elasticsearch_ml_job_state.nginx
  ]
}