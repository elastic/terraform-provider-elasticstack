Manages Kibana security lists (also known as value lists). Security lists are used by exception items to define sets of values for matching or excluding in security rules.

Relevant Kibana docs can be found [here](https://www.elastic.co/docs/api/doc/kibana/group/endpoint-security-lists-api).

## Notes

- Security lists define the type of data they can contain via the `type` attribute
- Once created, the `type` of a list cannot be changed
- Lists can be referenced by exception items to create more sophisticated matching rules
- The `list_id` is auto-generated if not provided
