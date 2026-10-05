variable "alias_name" {
  type = string
}

variable "write_index_name" {
  type = string
}

variable "read_indices_expression" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_alias" "test_alias" {
  name = var.alias_name

  write_index = {
    name = var.write_index_name
  }

  read_indices = [{
    name = var.read_indices_expression
  }]
}
