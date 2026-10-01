variable "space_id" {
  description = "The ID of the Kibana space"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id = var.space_id
  name     = "Advanced settings ${var.space_id}"
}

resource "elasticstack_kibana_advanced_settings" "test" {
  space_id = elasticstack_kibana_space.test.space_id
  settings = {
    "dateFormat:tz"                         = jsonencode("Europe/Berlin")
    "discover:sampleSize"                   = jsonencode(321)
    "courier:ignoreFilterIfFieldNotInIndex" = jsonencode(true)
    "defaultColumns"                        = jsonencode(["host.name", "message"])
    "timepicker:timeDefaults"               = jsonencode(jsonencode({ from = "now-30m", to = "now" }))
  }
}
