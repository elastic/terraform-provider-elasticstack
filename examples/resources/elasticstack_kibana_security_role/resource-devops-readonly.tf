provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_role" "devops_readonly" {
  name = "devops_readonly"

  elasticsearch {
    cluster = ["monitor"]

    indices {
      names      = ["logs-*", "metrics-*", "traces-*"]
      privileges = ["read", "view_index_metadata"]
    }
  }

  kibana {
    feature {
      name       = "apm"
      privileges = ["read"]
    }

    feature {
      name       = "fleet"
      privileges = ["read"]
    }

    feature {
      name       = "infrastructure"
      privileges = ["read"]
    }

    feature {
      name       = "logs"
      privileges = ["read"]
    }

    spaces = ["operations"]
  }
}
