Manages a Kibana Exception Item. Exception items define the specific query conditions used to prevent rules from generating alerts.

See the [Kibana Exceptions API documentation](https://www.elastic.co/docs/api/doc/kibana/group/endpoint-security-exceptions-api) for more details.

## Example Usage

```terraform
resource "elasticstack_kibana_security_exception_item" "example" {
  list_id       = elasticstack_kibana_security_exception_list.example.list_id
  item_id       = "my-exception-item"
  name          = "My Exception Item"
  description   = "Exclude specific processes from alerts"
  type          = "simple"
  namespace_type = "single"
  
  entries = jsonencode([
    {
      field = "process.name"
      operator = "included"
      type = "match"
      value = "my-process"
    }
  ])
  
  tags = ["tag1", "tag2"]
}
```
