provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "for_open_example" {
  job_id      = "example-ml-job-open-example"
  description = "Example anomaly detection job"

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

resource "elasticstack_elasticsearch_ml_job_state" "example" {
  job_id = elasticstack_elasticsearch_ml_anomaly_detection_job.for_open_example.job_id
  state  = "opened"

  force       = false
  job_timeout = "30s"

  timeouts = {
    create = "5m"
    update = "5m"
  }

  depends_on = [elasticstack_elasticsearch_ml_anomaly_detection_job.for_open_example]
}

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "for_close_example" {
  job_id      = "example-ml-job-close-example"
  description = "Example anomaly detection job for force-close illustration"

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

resource "elasticstack_elasticsearch_ml_job_state" "example_with_options" {
  job_id = elasticstack_elasticsearch_ml_anomaly_detection_job.for_close_example.job_id
  state  = "closed"

  force       = true
  job_timeout = "2m"

  timeouts = {
    create = "10m"
    update = "3m"
  }

  depends_on = [elasticstack_elasticsearch_ml_anomaly_detection_job.for_close_example]
}
