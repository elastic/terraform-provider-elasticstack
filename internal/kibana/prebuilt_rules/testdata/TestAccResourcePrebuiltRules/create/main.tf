variable "space_id" {
    type        = string
    description = "The ID of the Kibana space where prebuilt rules should be installed."
    default     = "default"
}

resource "elasticstack_kibana_install_prebuilt_rules" "test" {
  space_id = var.space_id
}