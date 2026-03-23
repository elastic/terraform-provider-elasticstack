Mapping for fields in the index.

If specified, this mapping can include: field names, [field data types](https://www.elastic.co/guide/en/elasticsearch/reference/current/mapping-types.html), [mapping parameters](https://www.elastic.co/guide/en/elasticsearch/reference/current/mapping-params.html).

**NOTE:**
- Changing datatypes in the existing _mappings_ will force index to be re-created.
- Removing a field will be ignored by default (same as Elasticsearch). You need to recreate the index to remove the field completely.
