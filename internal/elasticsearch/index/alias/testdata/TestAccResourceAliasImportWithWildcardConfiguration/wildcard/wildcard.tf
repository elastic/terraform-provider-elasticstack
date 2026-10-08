variable "alias_name" {
  description = "The alias name"
  type        = string
}

variable "read_indices_pattern" {
  description = "The wildcard expression for read index targets"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_alias" "test_alias" {
  name = var.alias_name

  read_indices = [{
    name = var.read_indices_pattern
  }]
}
