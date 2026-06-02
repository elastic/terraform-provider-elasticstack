variable "model_alias" {
  description = "The model alias"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_trained_model_alias" "test" {
  model_alias = var.model_alias
  model_id    = "lang_ident_model_1"
  reassign    = true
}
