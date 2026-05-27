provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_role" "data_analyst" {
  name = "data_analyst"

  elasticsearch {
    cluster = ["monitor"]

    indices {
      names      = ["logs-*", "metrics-*"]
      privileges = ["read", "view_index_metadata"]
    }
  }

  kibana {
    feature {
      name       = "dashboard"
      privileges = ["read"]
    }

    feature {
      name       = "discover"
      privileges = ["minimal_read", "url_create", "store_search_session"]
    }

    spaces = ["analytics"]
  }
}
