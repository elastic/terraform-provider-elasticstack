provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_trained_model_deployment" "test" {
  model_id              = "test-model"
  number_of_allocations = 1

  adaptive_allocations = {
    enabled = true
  }
}
