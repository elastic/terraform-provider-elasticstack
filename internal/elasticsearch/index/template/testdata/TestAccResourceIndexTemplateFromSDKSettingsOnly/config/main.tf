variable "template_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "upgrade" {
  name           = var.template_name
  index_patterns = ["${var.template_name}-*"]

  template {
    settings = jsonencode({
      index = { number_of_replicas = 1 }
    })
  }
}
