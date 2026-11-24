variable "first_space_id" {
  description = "The ID of the first Kibana space"
  type        = string
}

variable "second_space_id" {
  description = "The ID of the second Kibana space"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_advanced_settings" "first" {
  space_id = var.first_space_id

  string_settings = {
    "dateFormat:tz" = "Europe/Rome"
  }
}

resource "elasticstack_kibana_advanced_settings" "second" {
  space_id = var.second_space_id

  string_settings = {
    "dateFormat:tz" = "Europe/London"
  }
}
