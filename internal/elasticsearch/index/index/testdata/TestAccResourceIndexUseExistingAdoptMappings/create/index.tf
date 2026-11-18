variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

# Adopts an existing index that has legacy_field in its live mapping.
# The plan only declares new_field — legacy_field is intentionally omitted.
# After adoption the Put Mapping API should add new_field while legacy_field
# must remain present in the live cluster (Put Mapping cannot remove fields).
resource "elasticstack_elasticsearch_index" "test_adopt_mappings" {
  name         = var.index_name
  use_existing = true

  mappings = jsonencode({
    properties = {
      new_field = { type = "text" }
    }
  })

  deletion_protection = false
}
