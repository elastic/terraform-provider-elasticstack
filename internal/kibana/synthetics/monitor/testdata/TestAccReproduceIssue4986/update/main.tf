variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_synthetics_monitor" "issue_4986" {
  name      = "Issue 4986 Monitor Updated - ${var.name}"
  locations = ["us_west"]
  params    = jsonencode({ foo = "baz" })
  http = {
    url = "http://localhost:5601/status"
  }
}
