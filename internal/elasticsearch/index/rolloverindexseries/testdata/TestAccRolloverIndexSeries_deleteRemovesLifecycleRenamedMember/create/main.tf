variable "prefix" {
  type = string
}

variable "rollover_alias" {
  type = string
}

variable "deletion_protection" {
  type = bool
}

variable "update_timeout" {
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
  prefix              = var.prefix
  rollover_alias      = var.rollover_alias
  deletion_protection = var.deletion_protection

  timeouts = {
    update = var.update_timeout
  }

  depends_on = [elasticstack_elasticsearch_index_template.test]
}
