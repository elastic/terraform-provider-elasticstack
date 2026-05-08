provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_snapshot_repository" "my_url_repo" {
  name = "my_url_repo"

  url {
    url = "https://example.com/repo"
  }
}

resource "elasticstack_elasticsearch_snapshot_repository" "my_fs_repo" {
  name = "my_fs_repo"

  fs {
    location                  = "/tmp"
    compress                  = true
    max_restore_bytes_per_sec = "10mb"
  }
}

data "elasticstack_elasticsearch_snapshot_repository" "my_fs_repo" {
  name = elasticstack_elasticsearch_snapshot_repository.my_fs_repo.name
}

data "elasticstack_elasticsearch_snapshot_repository" "my_url_repo" {
  name = elasticstack_elasticsearch_snapshot_repository.my_url_repo.name
}

output "repo_fs_location" {
  value = data.elasticstack_elasticsearch_snapshot_repository.my_fs_repo.fs[0].location
}

output "repo_url" {
  value = data.elasticstack_elasticsearch_snapshot_repository.my_url_repo.url[0].url
}
