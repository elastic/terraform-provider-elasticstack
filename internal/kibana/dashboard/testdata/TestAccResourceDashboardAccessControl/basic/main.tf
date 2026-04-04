variable "dashboard_title" {
  type = string
}

variable "dashboard_username" {
  type = string
}

variable "dashboard_password" {
  type      = string
  sensitive = true
}

provider "elasticstack" {
  alias = "dashboard_user"
  elasticsearch {
    username = var.dashboard_username
    password = var.dashboard_password
  }
  kibana {
    username = var.dashboard_username
    password = var.dashboard_password
  }
}

resource "elasticstack_elasticsearch_security_user" "dashboard_user" {
  username  = var.dashboard_username
  password  = var.dashboard_password
  full_name = "Dashboard Access Control Test User"
  roles     = ["kibana_admin"]
}

resource "elasticstack_kibana_dashboard" "test" {
  provider = elasticstack.dashboard_user

  depends_on = [elasticstack_elasticsearch_security_user.dashboard_user]

  title       = var.dashboard_title
  description = "Test dashboard with access control"

  time_from = "now-15m"
  time_to   = "now"

  refresh_interval_pause = true
  refresh_interval_value = 90000

  query_language = "kql"
  query_text     = ""

  access_control = {
    access_mode = "write_restricted"
  }
}
