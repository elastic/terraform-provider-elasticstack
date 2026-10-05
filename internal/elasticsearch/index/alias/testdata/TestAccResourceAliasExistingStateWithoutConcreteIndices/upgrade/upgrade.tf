variable "alias_name" {
  description = "The alias name"
  type        = string
}

variable "index_name" {
  description = "The read index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_alias" "test_alias" {
  name = var.alias_name

  read_indices = [{
    name = var.index_name
  }]
}
