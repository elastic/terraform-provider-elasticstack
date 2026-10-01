variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title = var.dashboard_title

  query = {
    language = "lucene"
    text     = "status:200"
  }
}
