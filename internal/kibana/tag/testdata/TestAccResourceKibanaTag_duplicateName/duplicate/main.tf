variable "suffix" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_tag" "test" {
  name  = "tf-acc-tag-duplicate-name-${var.suffix}"
  color = "#FF0000"
}

resource "elasticstack_kibana_tag" "duplicate" {
  name  = "tf-acc-tag-duplicate-name-${var.suffix}"
  color = "#00FF00"

  depends_on = [
    elasticstack_kibana_tag.test,
  ]
}
