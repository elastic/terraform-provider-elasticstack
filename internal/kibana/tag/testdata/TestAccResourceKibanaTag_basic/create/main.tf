variable "suffix" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_tag" "test" {
  name  = "tf-acc-tag-${var.suffix}"
  color = "#FF0000"
}
