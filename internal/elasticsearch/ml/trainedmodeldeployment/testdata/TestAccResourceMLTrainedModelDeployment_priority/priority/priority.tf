variable "model_id" {
  description = "The model ID for the trained model deployment"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_trained_model_deployment" "test" {
  model_id = var.model_id
  priority = "low"
}
