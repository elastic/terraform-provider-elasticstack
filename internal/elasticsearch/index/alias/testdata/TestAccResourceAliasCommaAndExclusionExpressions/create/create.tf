variable "alias_name" {
  type = string
}

variable "comma_expression" {
  type = string
}

variable "exclusion_expression" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_alias" "test_alias" {
  name = var.alias_name

  read_indices = [
    {
      name = var.comma_expression
    },
    {
      name = var.exclusion_expression
    },
  ]
}
