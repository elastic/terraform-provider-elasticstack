variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
  name                = "my-index-${var.name}"
  deletion_protection = false
}

resource "elasticstack_kibana_slo" "test_slo" {
  name        = var.name
  description = "conflicting kql forms"

  kql_custom_indicator {
    index = "my-index-${var.name}"
    good  = "status: 200"
    good_kql = {
      kql_query = "http.response.status_code: 200"
    }
    total           = "*"
    timestamp_field = "custom_timestamp"
  }

  time_window {
    duration = "7d"
    type     = "rolling"
  }

  budgeting_method = "timeslices"

  objective {
    target           = 0.95
    timeslice_target = 0.95
    timeslice_window = "5m"
  }

  depends_on = [elasticstack_elasticsearch_index.my_index]
}
