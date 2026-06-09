variable "model_id" {
  description = "The model ID for the trained model deployment"
  type        = string
}

variable "deployment_id" {
  description = "A custom deployment ID"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_trained_model_deployment" "test" {
  model_id      = var.model_id
  deployment_id = var.deployment_id
}
