variable "transform_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_transform" "test" {
  name        = var.transform_name
  description = "neither pivot nor latest"

  source {
    indices = ["source_index_for_transform"]
  }

  destination {
    index = "dest_index_for_transform"
  }

  defer_validation = true
}
