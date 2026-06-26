variable "suffix" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_tag" "test" {
  name        = "tf-acc-tag-import-${var.suffix}"
  color       = "#ABCDEF"
  description = "import acceptance test"
}
