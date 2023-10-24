
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_role" "example" {
  name = "sample_role"
  elasticsearch {
    cluster = ["create_snapshot"]
    indices {
      field_security {
        grant  = ["test"]
        except = []
      }
      names      = ["test"]
      privileges = ["create", "read", "write"]
    }
  }
  kibana {
    base   = ["all"]
    spaces = ["default"]
  }
  kibana {
    feature {
      name       = "actions"
      privileges = ["read"]
    }
    feature {
      name       = "discover"
      privileges = ["minimal_read", "url_create", "store_search_session"]
    }
    feature {
      name       = "observabilityCases"
      privileges = ["minimal_read", "cases_delete"]
    }
    feature {
      name       = "osquery"
      privileges = ["minimal_read", "live_queries_all", "run_saved_queries", "saved_queries_read", "packs_all"]
    }
    feature {
      name       = "rulesSettings"
      privileges = ["minimal_read", "readFlappingSettings"]
    }
    feature {
      name       = "securitySolutionCases"
      privileges = ["minimal_read", "cases_delete"]
    }

    spaces = ["Default"]
  }
}
