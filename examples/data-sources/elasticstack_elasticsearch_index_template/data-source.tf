provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_index_template" "ilm-history-7" {
  name = "ilm-history-7"
}

output "template" {
  value = data.elasticstack_elasticsearch_index_template.ilm-history-7
}
