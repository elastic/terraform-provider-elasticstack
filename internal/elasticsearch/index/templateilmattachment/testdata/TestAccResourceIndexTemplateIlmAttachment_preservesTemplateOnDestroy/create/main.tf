provider "elasticstack" {
  elasticsearch {}
}

# ILM policy to attach
resource "elasticstack_elasticsearch_index_lifecycle" "test" {
  name = var.policy_name

  hot {
    rollover {
      max_age = "1d"
    }
  }

  delete {
    min_age = "30d"
    delete {}
  }
}

# Attachment adds ILM to the pre-created @custom component template.
# On destroy we only remove the ILM setting; the template (created in test PreCheck) remains.
resource "elasticstack_elasticsearch_index_template_ilm_attachment" "test" {
  index_template = var.index_template
  lifecycle_name = elasticstack_elasticsearch_index_lifecycle.test.name
}
