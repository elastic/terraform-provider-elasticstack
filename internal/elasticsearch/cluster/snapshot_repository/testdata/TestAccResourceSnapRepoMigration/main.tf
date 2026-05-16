variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_snapshot_repository" "test_migration" {
  name = var.name
  fs {
    location = "/tmp"
  }
}
