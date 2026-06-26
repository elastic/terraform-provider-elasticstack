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
  name        = "acc-kibana-tags-ds-${var.space_id}"
  description = "Kibana space for tags data source acceptance test"
}

resource "elasticstack_kibana_tag" "test" {
  space_id = elasticstack_kibana_space.test.space_id
  name     = "tf-acc-tag-space-${var.suffix}"
  color    = "#334455"
}

data "elasticstack_kibana_tags" "test" {
  space_id = elasticstack_kibana_space.test.space_id

  depends_on = [
    elasticstack_kibana_tag.test,
  ]
}
