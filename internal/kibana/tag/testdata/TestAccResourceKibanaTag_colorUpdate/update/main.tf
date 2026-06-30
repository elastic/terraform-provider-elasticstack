variable "suffix" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_tag" "test" {
  name  = "tf-acc-tag-color-${var.suffix}"
  color = "#00FF00"
}
