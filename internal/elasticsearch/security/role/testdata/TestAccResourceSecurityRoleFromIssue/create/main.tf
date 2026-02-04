variable "role_name" {
  description = "Name of the security role"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test_empty_sets" {
  name    = var.role_name
  cluster = [
    "manage_index_templates",
    "manage_ilm",
    "manage_pipeline",
    "manage_transform"
  ]

  indices {
    names      = ["slo-*", ".slo-*"]
    privileges = ["all"]
    field_security {
      grant  = ["*"]
      except = []
    }
    allow_restricted_indices = false
  }

  applications {
    application = "kibana-.kibana"
    privileges  = ["feature_slo.all"]
    resources   = ["*"]
  }

  run_as = []

  metadata = jsonencode({})
}
