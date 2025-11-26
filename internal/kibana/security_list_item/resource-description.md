---
subcategory: "Kibana"
layout: ""
page_title: "Elasticstack: elasticstack_kibana_security_list_item Resource"
description: |-
  Manages items within Kibana security value lists.
---

# Resource: elasticstack_kibana_security_list_item

Manages items within Kibana security value lists. Value lists are containers for values that can be used within exception lists to define conditions. This resource allows you to add, update, and remove individual values (items) in those lists.

Value list items are used to store data values that match the type of their parent security list (e.g., IP addresses, keywords, etc.). These items can then be referenced in exception list entries to define exception conditions.

Kibana docs can be found [here](https://www.elastic.co/docs/api/doc/kibana/group/endpoint-security-lists-api)