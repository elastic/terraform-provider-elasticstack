variable "suffix" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_tag" "tag_a" {
  name  = "tf-acc-tag-a-${var.suffix}"
  color = "#111111"
}

resource "elasticstack_kibana_tag" "tag_b" {
  name  = "tf-acc-tag-b-${var.suffix}"
  color = "#222222"
}

data "elasticstack_kibana_tags" "test" {
  query = var.suffix

  depends_on = [
    elasticstack_kibana_tag.tag_a,
    elasticstack_kibana_tag.tag_b,
  ]
}
