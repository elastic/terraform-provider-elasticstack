variable "model_alias" {
  description = "The model alias"
  type        = string
}

variable "model_b" {
  description = "Second trained model ID"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_trained_model_alias" "test" {
  model_alias = var.model_alias
  model_id    = var.model_b
}
