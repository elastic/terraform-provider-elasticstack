variable "space_id" {
  description = "The ID of the Kibana space"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_advanced_settings" "test" {
  space_id = var.space_id

  json_settings = {
    "custom:ratio"  = "1.50"
    "custom:whole"  = "1.0"
    "custom:scaled" = "1e2"
    "custom:nested" = "{\"weights\": [1.0, 2.50], \"limit\": 1e2}"
  }
}
