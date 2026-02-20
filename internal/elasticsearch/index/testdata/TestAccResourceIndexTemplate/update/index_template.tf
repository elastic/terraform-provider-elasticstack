provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name = var.template_name

  index_patterns = ["${var.template_name}-logs-*"]

  template {
    alias {
      name = "my_template_test"
    }
    alias {
      name = "alias2"
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
    hidden = false
  }

  template {}
}
