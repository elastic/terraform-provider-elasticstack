variable "suffix" {
  type = string
}

variable "tag_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_tag" "test" {
  tag_id = var.tag_id
  name   = "tf-acc-tag-duplicate-id-${var.suffix}"
  color  = "#FF0000"
}
