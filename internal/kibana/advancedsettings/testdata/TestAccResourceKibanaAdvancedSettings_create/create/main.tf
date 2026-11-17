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

  string_settings = {
    "dateFormat:tz"   = "Europe/Rome"
    "custom:jsonText" = "[\"a\",\"b\"]"
    "custom:teamNote" = "hello"
  }

  json_settings = {
    "csv:quoteValues"    = "false"
    "truncate:maxHeight" = "0"
    "defaultColumns"     = "[\"@timestamp\",\"message\"]"
  }
}
