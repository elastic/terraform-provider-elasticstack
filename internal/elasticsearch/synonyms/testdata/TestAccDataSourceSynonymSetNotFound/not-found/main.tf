variable "synonym_set_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_synonym_set" "test" {
  synonym_set_id = var.synonym_set_id
}
