provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name    = var.role_name
  cluster = ["monitor"]

  # When allow_restricted_indices is false (or omitted, which defaults to false),
  # the API returns nil (omitempty). The provider should return false, not null, to prevent plan drift.
  indices {
    names                     = ["index1", "index2"]
    privileges                = ["read", "view_index_metadata"]
    allow_restricted_indices  = false
    field_security {
      grant  = ["*"]
      # Test case: empty except array should not cause plan drift when API omits it
      except = []
    }
  }

  metadata = jsonencode({
    version = 1
  })
}
