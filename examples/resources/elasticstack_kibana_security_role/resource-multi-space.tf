provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_role" "multi_space" {
  name = "multi_space"

  elasticsearch {
    cluster = ["monitor"]

    indices {
      names      = ["logs-*", "metrics-*", "traces-*"]
      privileges = ["read", "view_index_metadata"]
    }
  }

  kibana {
    base   = ["all"]
    spaces = ["dev", "staging"]
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

    spaces = ["prod"]
  }
}
