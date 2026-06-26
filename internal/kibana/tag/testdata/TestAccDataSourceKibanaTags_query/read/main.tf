variable "suffix" {
  type = string
}

variable "query_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_tag" "tag_a" {
  name  = var.query_name
  color = "#111111"
}

resource "elasticstack_kibana_tag" "tag_b" {
  name  = "tf-acc-tag-other-${var.suffix}"
  color = "#222222"
}

data "elasticstack_kibana_tags" "test" {
  query = var.query_name
}
