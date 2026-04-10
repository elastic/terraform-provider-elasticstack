variable "tag_key" {
  type = string
}

variable "tag_value" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test_rule_1" {
  name        = "Test Rule 1 - ${var.tag_value}"
  type        = "query"
  query       = "event.action:test"
  language    = "kuery"
  description = "Test rule for enable_rule resource"
  severity    = "low"
  risk_score  = 21
  index       = ["logs-*"]
  tags        = ["${var.tag_key}: ${var.tag_value}", "test"]

  lifecycle {
    ignore_changes = [enabled]
  }
}

resource "elasticstack_kibana_security_detection_rule" "test_rule_2" {
  name        = "Test Rule 2 - ${var.tag_value}"
  type        = "query"
  query       = "event.action:test2"
  language    = "kuery"
  description = "Test rule for enable_rule resource"
  severity    = "low"
  risk_score  = 21
  index       = ["logs-*"]
  tags        = ["${var.tag_key}: ${var.tag_value}", "test"]

  lifecycle {
    ignore_changes = [enabled]
  }
}

resource "elasticstack_kibana_security_enable_rule" "test" {
  space_id           = "default"
  key                = var.tag_key
  value              = var.tag_value
  disable_on_destroy = true

  depends_on = [
    elasticstack_kibana_security_detection_rule.test_rule_1,
    elasticstack_kibana_security_detection_rule.test_rule_2
  ]
}
