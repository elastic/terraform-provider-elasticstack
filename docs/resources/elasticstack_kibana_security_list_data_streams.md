---
subcategory: "Kibana"
layout: ""
page_title: "Elasticstack: elasticstack_kibana_security_list_data_streams Resource"
description: |-
  Creates and manages `.lists` and `.items` data streams for security lists in Kibana.
---

# Resource: <no value>

<no value>

Before you can start working with exceptions that use value lists, you must create the `.lists` and `.items` data streams for the relevant Kibana space. Once these data streams are created, your role needs privileges to manage rules.

See the [Elastic documentation](https://www.elastic.co/docs/api/doc/kibana/operation/operation-createlistindex) for more details.

## Example Usage

```terraform
# Create list data streams in the default space
resource "elasticstack_kibana_security_list_data_streams" "default" {
}

# Create list data streams in a custom space
resource "elasticstack_kibana_security_list_data_streams" "custom" {
  space_id = "my-space"
}
```

## Argument Reference

The following arguments are supported:

<no value>

## Import

Import is supported using the following syntax:

```shell
# List data streams can be imported using the space ID
terraform import elasticstack_kibana_security_list_data_streams.default default
terraform import elasticstack_kibana_security_list_data_streams.custom my-space
```
