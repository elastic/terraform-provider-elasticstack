resource "elasticstack_fleet_proxy" "output_proxy" {
  name = "Output Proxy"
  url  = "https://proxy.example.com:8080"
}

resource "elasticstack_fleet_output" "elasticsearch" {
  name     = "elasticsearch"
  type     = "elasticsearch"
  hosts    = ["https://elasticsearch:9200"]
  proxy_id = elasticstack_fleet_proxy.output_proxy.proxy_id
}
