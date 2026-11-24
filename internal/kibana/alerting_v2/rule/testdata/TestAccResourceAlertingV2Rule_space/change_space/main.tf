variable "name" {
  description = "Unique rule name"
  type        = string
}

variable "rule_id" {
  description = "Rule identifier"
  type        = string
}

variable "space_id" {
  description = "Kibana space the rule was created in"
  type        = string
}

variable "new_space_id" {
  description = "Kibana space the rule is moved to"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id = var.space_id
  name     = var.space_id
}

resource "elasticstack_kibana_space" "changed" {
  space_id = var.new_space_id
  name     = var.new_space_id
}

resource "elasticstack_kibana_alerting_v2_rule" "test" {
  space_id = elasticstack_kibana_space.changed.space_id
  rule_id  = var.rule_id
  kind     = "alert"

  metadata = {
    name = var.name
  }

  schedule = {
    every = "1m"
  }

  query = {
    base = "FROM logs-* | STATS count = COUNT(*) BY host.name | WHERE count > 100"
  }

  recovery = {
    strategy = "no_breach"
  }

  no_data = {
    strategy = "ignore"
  }
}
