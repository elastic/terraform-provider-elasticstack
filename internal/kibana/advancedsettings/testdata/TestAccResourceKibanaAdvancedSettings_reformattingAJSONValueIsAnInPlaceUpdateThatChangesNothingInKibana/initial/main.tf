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
    "custom:ratio"  = "1.5"
    "custom:list"   = "[\"a\",\"b\"]"
    "custom:object" = "{\"a\":1,\"b\":{\"c\":2,\"d\":3}}"
  }
}
