resource "elasticstack_elasticsearch_index" "ml_datafeed_index" {
  name = "ml-datafeed-data"
  mappings = jsonencode({
    properties = {
      "@timestamp" = {
        type = "date"
      }
      value = {
        type = "double"
      }
      user = {
        type = "keyword"
      }
    }
  })
}

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "example" {
  job_id      = "example-anomaly-job"
  description = "Example anomaly detection job"

  analysis_config {
    bucket_span = "15m"
    detectors {
      function      = "mean"
      field_name    = "value"
      by_field_name = "user"
    }
  }

  data_description {
    time_field = "@timestamp"
  }
}

resource "elasticstack_elasticsearch_ml_datafeed" "example" {
  datafeed_id = "example-datafeed"
  job_id      = elasticstack_elasticsearch_ml_anomaly_detection_job.example.job_id
  indices     = [elasticstack_elasticsearch_index.ml_datafeed_index.name]

  query = jsonencode({
    bool = {
      must = [
        {
          range = {
            "@timestamp" = {
              gte = "now-7d"
            }
          }
        }
      ]
    }
  })
}

resource "elasticstack_elasticsearch_ml_datafeed_state" "example" {
  datafeed_id = elasticstack_elasticsearch_ml_datafeed.example.datafeed_id
  state       = "started"
  force       = false

  depends_on = [
    elasticstack_elasticsearch_ml_datafeed.example,
    elasticstack_elasticsearch_ml_anomaly_detection_job.example
  ]
}