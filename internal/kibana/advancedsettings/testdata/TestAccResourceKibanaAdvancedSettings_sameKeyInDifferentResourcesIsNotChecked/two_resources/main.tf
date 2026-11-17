variable "space_id" {
  description = "The ID of the Kibana space"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_advanced_settings" "first" {
  space_id = var.space_id

  string_settings = {
    "dateFormat:tz" = "Europe/Rome"
  }
}

resource "elasticstack_kibana_advanced_settings" "second" {
  space_id = var.space_id

  string_settings = {
    "dateFormat:tz" = "Europe/London"
  }
}
