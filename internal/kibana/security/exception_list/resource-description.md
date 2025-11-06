Manages a Kibana Exception List. Exception lists are containers for exception items used to prevent security rules from generating alerts.

See the [Kibana Exceptions API documentation](https://www.elastic.co/docs/api/doc/kibana/group/endpoint-security-exceptions-api) for more details.

## Example Usage

```terraform
resource "elasticstack_kibana_security_exception_list" "example" {
  list_id       = "my-exception-list"
  name          = "My Exception List"
  description   = "List of exceptions for security rules"
  type          = "detection"
  namespace_type = "single"
  
  tags = ["tag1", "tag2"]
}
```
