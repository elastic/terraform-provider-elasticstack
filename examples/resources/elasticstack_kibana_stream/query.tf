resource "elasticstack_kibana_stream" "nginx_errors" {
  name     = "logs.nginx.errors-view"
  space_id = "default"

  query_config = {
    # `view` is computed ($.<stream name>). FROM must use $.{parent prefix} notation (last dot segment stripped).
    esql = "FROM $.logs.nginx | WHERE http.response.status_code >= 400"
  }
}
