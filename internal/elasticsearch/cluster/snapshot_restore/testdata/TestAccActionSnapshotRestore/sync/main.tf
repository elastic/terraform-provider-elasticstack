variable "name" {
  type = string
}

variable "snapshot_name" {
  type = string
}

locals {
  index_name    = "${var.name}-idx"
  restored_name = "restored-${var.name}-idx"
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
  name = local.index_name

  mappings = jsonencode({
    properties = {
      field1 = { type = "keyword" }
    }
  })
}

action "elasticstack_elasticsearch_snapshot_create" "bootstrap" {
  repository          = elasticstack_elasticsearch_snapshot_repository.repo.name
  snapshot            = var.snapshot_name
  indices             = [elasticstack_elasticsearch_index.source.name]
  include_global_state = false
  wait_for_completion = true
}

resource "terraform_data" "trigger_create" {
  depends_on = [
    elasticstack_elasticsearch_index.source,
    elasticstack_elasticsearch_snapshot_repository.repo,
  ]

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.elasticstack_elasticsearch_snapshot_create.bootstrap]
    }
  }
}

action "elasticstack_elasticsearch_snapshot_restore" "restore" {
  repository           = elasticstack_elasticsearch_snapshot_repository.repo.name
  snapshot             = var.snapshot_name
  rename_pattern       = "(.+)"
  rename_replacement   = "restored-$1"
  include_global_state = false
  wait_for_completion  = true
}

resource "terraform_data" "trigger_restore" {
  depends_on = [terraform_data.trigger_create]

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.elasticstack_elasticsearch_snapshot_restore.restore]
    }
  }
}
