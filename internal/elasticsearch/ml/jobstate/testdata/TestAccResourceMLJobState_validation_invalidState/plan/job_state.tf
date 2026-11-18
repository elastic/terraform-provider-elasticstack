provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_job_state" "test" {
  job_id = "test-ml-job-state-validation"
  state  = "paused"
}
