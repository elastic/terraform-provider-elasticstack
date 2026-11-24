resource "elasticstack_kibana_space" "ops" {
  space_id = "ops"
  name     = "Operations"
}

resource "elasticstack_kibana_advanced_settings" "ops" {
  space_id = elasticstack_kibana_space.ops.space_id

  # Settings sent to Kibana as plain strings. Settings that store JSON as a
  # string, such as timepicker:timeDefaults, belong here as well.
  string_settings = {
    "dateFormat:tz"           = "Europe/Rome"
    "timepicker:timeDefaults" = jsonencode({ from = "now-7d", to = "now" })
  }

  # Settings sent to Kibana as JSON, keeping their boolean, number or array type.
  json_settings = {
    "csv:quoteValues"    = jsonencode(false)
    "truncate:maxHeight" = jsonencode(0)
    "defaultColumns"     = jsonencode(["@timestamp", "message"])
  }
}
