provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_role" "security_analyst" {
  name = "security_analyst"

  elasticsearch {
    cluster = ["monitor"]

    indices {
      names      = ["logs-*", "metrics-*", "traces-*", ".alerts-security.*"]
      privileges = ["read", "view_index_metadata"]
    }
  }

  kibana {
    feature {
      name       = "actions"
      privileges = ["all"]
    }

    feature {
      name       = "alerting"
      privileges = ["all"]
    }

    feature {
      name       = "osquery"
      privileges = ["all"]
    }

    feature {
      name       = "rulesSettings"
      privileges = ["all"]
    }

    feature {
      name       = "securitySolutionCases"
      privileges = ["all"]
    }

    feature {
      name       = "siem"
      privileges = ["all"]
    }

    spaces = ["security"]
  }
}
