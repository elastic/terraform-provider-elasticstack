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

resource "elasticstack_kibana_tag" "duplicate" {
  tag_id = var.tag_id
  name   = "tf-acc-tag-duplicate-id-other-${var.suffix}"
  color  = "#00FF00"

  depends_on = [
    elasticstack_kibana_tag.test,
  ]
}
