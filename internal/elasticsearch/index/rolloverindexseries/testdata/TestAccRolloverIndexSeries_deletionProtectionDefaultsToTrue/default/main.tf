variable "prefix" {
  type = string
}

variable "rollover_alias" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = "${var.prefix}-template"
  index_patterns = ["${var.prefix}-*"]
  priority       = 500

  template {
    settings = jsonencode({
      number_of_replicas = 0
      refresh_interval   = "30s"
    })
  }
}

resource "elasticstack_elasticsearch_rollover_index_series" "test" {
  prefix         = var.prefix
  rollover_alias = var.rollover_alias

  depends_on = [elasticstack_elasticsearch_index_template.test]
}
