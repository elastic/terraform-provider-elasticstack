variable "name" {
  type = string
}

variable "space_id" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "Test Space for Detection Rules"
  description = "Space for testing detection rules"
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  space_id    = elasticstack_kibana_space.test.space_id
  name        = var.name
  type        = "query"
  query       = "*:*"
  language    = "kuery"
  enabled     = true
  description = "Minimal test query security detection rule in custom space"
  severity    = "low"
  risk_score  = 21
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  index       = ["logs-*"]
}

