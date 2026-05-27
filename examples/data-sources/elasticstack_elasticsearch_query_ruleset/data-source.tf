data "elasticstack_elasticsearch_query_ruleset" "example" {
  ruleset_id = "my-search-rules"
}

output "rules" {
  value = data.elasticstack_elasticsearch_query_ruleset.example.rules
}
