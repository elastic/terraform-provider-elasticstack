variable "space_id" {
  description = "The ID of the Kibana space"
  type        = string
}

variable "long_key" {
  description = "A setting key longer than the 1024 characters Kibana accepts"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_advanced_settings" "test" {
  space_id = var.space_id

  json_settings = {
    (var.long_key) = "0"
  }
}
