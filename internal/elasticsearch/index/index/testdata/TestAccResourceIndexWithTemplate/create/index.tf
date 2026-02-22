variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = var.index_name
  index_patterns = [var.index_name]

  template {
    settings = jsonencode({
      default_pipeline = ".fleet_final_pipeline-1"
      lifecycle        = { name = ".monitoring-8-ilm-policy" }
    })

    mappings = jsonencode({
      dynamic_templates = [
        {
          strings_as_ip = {
            match_mapping_type = "string"
            match              = "ip*"
            runtime = {
              type = "ip"
            }
          }
        },
      ]
    })
  }
}

resource "elasticstack_elasticsearch_index" "test" {
  name                = var.index_name
  deletion_protection = false

  alias = [
    {
      name           = "${var.index_name}-alias"
      is_write_index = true
    },
  ]

  lifecycle {
    ignore_changes = [mappings]
  }

  depends_on = [elasticstack_elasticsearch_index_template.test]
}
