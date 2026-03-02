variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name                = var.name
  type                = "query"
  query               = "process.name:* AND user.name:*"
  language            = "kuery"
  enabled             = true
  description         = "Updated test building block security detection rule"
  severity            = "medium"
  risk_score          = 40
  from                = "now-6m"
  to                  = "now"
  interval            = "5m"
  data_view_id        = "updated-building-block-data-view-id"
  namespace           = "updated-building-block-namespace"
  building_block_type = "default"
  author              = ["Test Author"]
  tags                = ["building-block", "test"]
  license             = "Elastic License v2"
}

