variable "name" {
  type = string
}

variable "snapshot_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_snapshot_repository" "repo" {
  name = "${var.name}-repo"

  fs {
    location = "/tmp/snapshots"
  }
}

resource "elasticstack_elasticsearch_index" "source" {
  name = "${var.name}-idx"

  mappings = jsonencode({
    properties = {
      field1 = { type = "keyword" }
    }
  })

  deletion_protection = false
}

action "elasticstack_elasticsearch_snapshot_create" "create" {
  config {
    repository           = elasticstack_elasticsearch_snapshot_repository.repo.name
    snapshot             = var.snapshot_name
    indices              = [elasticstack_elasticsearch_index.source.name]
    include_global_state = false
    metadata             = jsonencode({ created_by = "terraform", env = "test" })
    wait_for_completion  = true
  }
}

resource "terraform_data" "trigger_create" {
  depends_on = [
    elasticstack_elasticsearch_index.source,
    elasticstack_elasticsearch_snapshot_repository.repo,
  ]

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.elasticstack_elasticsearch_snapshot_create.create]
    }
  }
}
