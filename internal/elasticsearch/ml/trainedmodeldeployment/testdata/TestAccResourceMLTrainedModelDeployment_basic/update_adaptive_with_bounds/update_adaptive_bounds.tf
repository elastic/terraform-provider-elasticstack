variable "model_id" {
  description = "The model ID for the trained model deployment"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_trained_model_deployment" "test" {
  model_id = var.model_id

  adaptive_allocations = {
    enabled                   = true
    min_number_of_allocations = 1
    max_number_of_allocations = 3
  }
}
