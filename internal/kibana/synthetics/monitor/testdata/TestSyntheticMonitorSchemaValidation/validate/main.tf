variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

resource "elasticstack_kibana_synthetics_monitor" "invalid-monitor" {
  name      = "TestHttpMonitor Invalid - ${var.name}"
  locations = ["us_central_qa"]
  namespace = "***"
  http = {
    url = "http://localhost:5601"
  }
}
