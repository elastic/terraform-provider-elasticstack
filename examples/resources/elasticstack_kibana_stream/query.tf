resource "elasticstack_kibana_stream" "nginx_errors" {
  name     = "logs.nginx.errors-view"
  space_id = "default"

  query_config = {
    esql = "FROM logs.nginx | WHERE http.response.status_code >= 400"
    view = "logs-nginx-errors"
  }
}
