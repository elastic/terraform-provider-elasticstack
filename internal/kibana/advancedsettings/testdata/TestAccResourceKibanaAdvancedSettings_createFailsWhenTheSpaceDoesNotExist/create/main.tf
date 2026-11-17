variable "space_id" {
  description = "The ID of a Kibana space that does not exist"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_advanced_settings" "test" {
  space_id = var.space_id

  string_settings = {
    "custom:teamNote" = "hello"
  }
}
