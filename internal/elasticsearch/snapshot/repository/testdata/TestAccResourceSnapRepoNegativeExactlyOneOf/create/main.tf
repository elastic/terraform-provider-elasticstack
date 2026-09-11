variable "name" {
  description = "The snapshot repository name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_snapshot_repository" "test_negative" {
  name = var.name

  fs {
    location = "/tmp"
  }

  url {
    url = "file:/tmp"
  }
}
