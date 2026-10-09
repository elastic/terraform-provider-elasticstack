variable "space_a_id" {
  description = "The ID of the first Kibana space"
  type        = string
}

variable "space_b_id" {
  description = "The ID of the second Kibana space"
  type        = string
}

variable "active_space_id" {
  description = "The space_id currently assigned to the prebuilt rules resource"
  type        = string
}

resource "elasticstack_kibana_space" "space_a" {
  space_id    = var.space_a_id
  name        = "Space A ${var.space_a_id}"
  description = "Test space A for prebuilt rules space_id replace test"
}

resource "elasticstack_kibana_space" "space_b" {
  space_id    = var.space_b_id
  name        = "Space B ${var.space_b_id}"
  description = "Test space B for prebuilt rules space_id replace test"
}

resource "elasticstack_kibana_install_prebuilt_rules" "test" {
  space_id = var.active_space_id

  depends_on = [
    elasticstack_kibana_space.space_a,
    elasticstack_kibana_space.space_b,
  ]
}
