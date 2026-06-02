variable "name" {
  description = "The snapshot repository name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_snapshot_repository" "test_fs_repo" {
  name = var.name
}
