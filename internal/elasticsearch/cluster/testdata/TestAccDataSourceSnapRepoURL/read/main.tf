variable "name" {
  description = "The snapshot repository name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_snapshot_repository" "test_url_repo" {
  name = var.name

  url {
    url = "file:/tmp"
  }
}

data "elasticstack_elasticsearch_snapshot_repository" "test_url_repo" {
  name = elasticstack_elasticsearch_snapshot_repository.test_url_repo.name
}
