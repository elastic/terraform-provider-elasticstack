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
    "custom:list"   = "[ \"a\",   \"b\" ]"
    "custom:object" = "{\"b\":{\"d\":3,\"c\":2},\"a\":1}"
  }
}
