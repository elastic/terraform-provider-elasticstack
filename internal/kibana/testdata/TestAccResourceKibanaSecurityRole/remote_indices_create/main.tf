variable "role_name" {
  description = "The role name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_role" "test" {
  name = var.role_name
  elasticsearch {
    cluster = ["create_snapshot"]
    indices {
      field_security {
        grant  = ["sample"]
        except = []
      }
      names      = ["sample"]
      privileges = ["create", "read", "write"]
    }
    remote_indices {
      clusters = ["test-cluster"]
      field_security {
        grant  = ["sample"]
        except = []
      }
      names      = ["sample"]
      privileges = ["create", "read", "write"]
    }
  }
  kibana {
    feature {
      name       = "actions"
      privileges = ["read"]
    }
    feature {
      name       = "advancedSettings"
      privileges = ["read"]
    }
    feature {
      name       = "discover"
      privileges = ["minimal_read", "url_create", "store_search_session"]
    }
    feature {
      name       = "generalCases"
      privileges = ["minimal_read", "cases_delete"]
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
    feature {
      name       = "visualize"
      privileges = ["minimal_read", "url_create"]
    }

    spaces = ["default"]
  }
}
