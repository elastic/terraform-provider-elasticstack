variable "space_id" {
  description = "The ID of the Kibana space"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "terraform_data" "source" {
  input = "not json"
}

resource "elasticstack_kibana_advanced_settings" "test" {
  space_id = var.space_id

  json_settings = {
    "truncate:maxHeight" = terraform_data.source.output
    "csv:quoteValues"    = "false"
  }
}
