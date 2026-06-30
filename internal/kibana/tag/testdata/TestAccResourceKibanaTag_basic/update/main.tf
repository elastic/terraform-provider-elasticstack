variable "suffix" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_tag" "test" {
  name  = "tf-acc-tag-updated-${var.suffix}"
  color = "#FF0000"
}
