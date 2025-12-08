variable "space_id" {
    type        = string
    description = "The ID of the Kibana space where prebuilt rules should be installed."
    default     = "default"
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "Test Space for Prebuilt Rules"
  description = "A Kibana space created for acceptance testing of prebuilt rules."
}

resource "elasticstack_kibana_install_prebuilt_rules" "test" {
  space_id = var.space_id
}