provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name = var.template_name

  priority       = 42
  index_patterns = ["${var.template_name}-logs-*"]

  template {
    alias {
      name = "my_template_test"
    }

    settings = jsonencode({
      number_of_shards = "3"
    })
  }
}

resource "elasticstack_elasticsearch_index_template" "test2" {
  name = "${var.template_name}-stream"

  index_patterns = ["index-pattern-streams*"]

  data_stream {
    hidden = true
  }
}
