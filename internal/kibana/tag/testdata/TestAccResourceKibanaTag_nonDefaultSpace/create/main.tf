variable "suffix" {
  type = string
}

variable "space_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "acc-kibana-tag-${var.space_id}"
  description = "Kibana space for tag acceptance test"
}

resource "elasticstack_kibana_tag" "test" {
  space_id = elasticstack_kibana_space.test.space_id
  name     = "tf-acc-tag-space-${var.suffix}"
  color    = "#445566"
}
