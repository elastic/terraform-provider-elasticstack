provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = var.template_name
  index_patterns = ["${var.template_name}-*"]

  data_stream {}

  template {
    data_stream_options {
      failure_store {
        enabled = true
      }
    }
  }
}
