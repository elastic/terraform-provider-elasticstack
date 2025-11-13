provider "elasticstack" {
  elasticsearch {}
}

# Try to manage state of a non-existent ML job
resource "elasticstack_elasticsearch_ml_job_state" "test" {
  job_id = "non-existent-ml-job"
  state  = "opened"
}