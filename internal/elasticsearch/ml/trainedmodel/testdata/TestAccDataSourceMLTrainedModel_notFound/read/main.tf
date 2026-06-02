provider "elasticstack" {
  elasticsearch {}
}

variable "model_id" {
  type = string
}

data "elasticstack_elasticsearch_ml_trained_model" "test" {
  model_id = var.model_id
}
